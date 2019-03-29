// Package gjkr contains code that implements Distributed Key Generation protocol
// described in [GJKR 99].
//
// See http://docs.keep.network/random-beacon/dkg.html
//
//     [GJKR 99]: Gennaro R., Jarecki S., Krawczyk H., Rabin T. (1999) Secure
//         Distributed Key Generation for Discrete-Log Based Cryptosystems. In:
//         Stern J. (eds) Advances in Cryptology — EUROCRYPT ’99. EUROCRYPT 1999.
//         Lecture Notes in Computer Science, vol 1592. Springer, Berlin, Heidelberg
//         http://groups.csail.mit.edu/cis/pubs/stasio/vss.ps.gz
package gjkr

import (
	crand "crypto/rand"
	"fmt"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"github.com/keep-network/keep-core/pkg/beacon/relay/member"
	"github.com/keep-network/keep-core/pkg/net/ephemeral"
)

// GenerateEphemeralKeyPair takes the group member list and generates an
// ephemeral ECDH keypair for every other group member. Generated public
// ephemeral keys are broadcasted within the group.
//
// See Phase 1 of the protocol specification.
func (em *EphemeralKeyPairGeneratingMember) GenerateEphemeralKeyPair() (
	*EphemeralPublicKeyMessage,
	error,
) {
	ephemeralKeys := make(map[member.MemberIndex]*ephemeral.PublicKey)

	// Calculate ephemeral key pair for every other group member
	for _, member := range em.group.memberIDs {
		if member == em.ID {
			// don’t actually generate a key with ourselves
			continue
		}

		ephemeralKeyPair, err := ephemeral.GenerateKeyPair()
		if err != nil {
			return nil, err
		}

		// save the generated ephemeral key to our state
		em.ephemeralKeyPairs[member] = ephemeralKeyPair

		// store the public key to the map for the message
		ephemeralKeys[member] = ephemeralKeyPair.PublicKey
	}

	return &EphemeralPublicKeyMessage{
		senderID:            em.ID,
		ephemeralPublicKeys: ephemeralKeys,
	}, nil
}

// GenerateSymmetricKeys attempts to generate symmetric keys for all remote group
// members via ECDH. It generates this symmetric key for each remote group member
// by doing an ECDH between the ephemeral private key generated for a remote
// group member, and the public key for this member, generated and broadcasted by
// the remote group member.
//
// See Phase 2 of the protocol specification.
func (sm *SymmetricKeyGeneratingMember) GenerateSymmetricKeys(
	ephemeralPubKeyMessages []*EphemeralPublicKeyMessage,
) error {
	for _, ephemeralPubKeyMessage := range ephemeralPubKeyMessages {
		sm.evidenceLog.PutEphemeralMessage(ephemeralPubKeyMessage)

		otherMember := ephemeralPubKeyMessage.senderID

		// Find the ephemeral key pair generated by this group member for
		// the other group member.
		ephemeralKeyPair, ok := sm.ephemeralKeyPairs[otherMember]
		if !ok {
			return fmt.Errorf(
				"ephemeral key pair does not exist for member %v",
				otherMember,
			)
		}

		// Get the ephemeral private key generated by this group member for
		// the other group member.
		thisMemberEphemeralPrivateKey := ephemeralKeyPair.PrivateKey

		// Get the ephemeral public key broadcasted by the other group member,
		// which was intended for this group member.
		otherMemberEphemeralPublicKey := ephemeralPubKeyMessage.ephemeralPublicKeys[sm.ID]

		// Create symmetric key for the current group member and the other
		// group member by ECDH'ing the public and private key.
		symmetricKey := thisMemberEphemeralPrivateKey.Ecdh(
			otherMemberEphemeralPublicKey,
		)
		sm.symmetricKeys[otherMember] = symmetricKey
	}

	return nil
}

// CalculateMembersSharesAndCommitments starts with generating coefficients for
// two polynomials. It then calculates shares for all group member and packs
// them into a broadcast message. Individual shares inside the message are
// encrypted with the symmetric key of the indended share receiver.
// Additionally, it calculates commitments to `a` coefficients of first
// polynomial using second's polynomial `b` coefficients.
//
// If there are no symmetric keys established with all other group members,
// function yields an error.
//
// See Phase 3 of the protocol specification.
func (cm *CommittingMember) CalculateMembersSharesAndCommitments() (
	*PeerSharesMessage,
	*MemberCommitmentsMessage,
	error,
) {
	polynomialDegree := cm.group.dishonestThreshold
	coefficientsA, err := generatePolynomial(polynomialDegree)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"could not generate shares polynomial [%v]",
			err,
		)
	}
	coefficientsB, err := generatePolynomial(polynomialDegree)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"could not generate hiding polynomial [%v]",
			err,
		)
	}

	cm.secretCoefficients = coefficientsA

	// Calculate shares for other group members by evaluating polynomials
	// defined by coefficients `a_i` and `b_i`
	var sharesMessage = newPeerSharesMessage(cm.ID)
	for _, receiverID := range cm.group.MemberIDs() {
		// s_j = f_(j) mod q
		memberShareS := cm.evaluateMemberShare(receiverID, coefficientsA)
		// t_j = g_(j) mod q
		memberShareT := cm.evaluateMemberShare(receiverID, coefficientsB)

		// Check if calculated shares for the current member.
		// If so, store them without sharing in a message.
		if cm.ID == receiverID {
			cm.selfSecretShareS = memberShareS
			cm.selfSecretShareT = memberShareT
			continue
		}

		// If there is no symmetric key established with the receiver,
		// yield an error.
		symmetricKey, hasKey := cm.symmetricKeys[receiverID]
		if !hasKey {
			return nil, nil, fmt.Errorf(
				"no symmetric key for receiver %v", receiverID,
			)
		}

		err := sharesMessage.addShares(
			receiverID,
			memberShareS,
			memberShareT,
			symmetricKey,
		)
		if err != nil {
			return nil, nil, fmt.Errorf(
				"could not add shares for receiver %v [%v]",
				receiverID,
				err,
			)
		}
	}

	commitments := make([]*bn256.G1, len(coefficientsA))
	for k := range commitments {
		commitments[k] = cm.calculateCommitment(coefficientsA[k], coefficientsB[k])
	}
	commitmentsMessage := &MemberCommitmentsMessage{
		senderID:    cm.ID,
		commitments: commitments,
	}

	return sharesMessage, commitmentsMessage, nil
}

// calculateCommitment generates a Pedersen commitment to a secret value
// `secret` with a blinding factor `t`.
func (cm *CommittingMember) calculateCommitment(
	secret *big.Int,
	t *big.Int,
) *bn256.G1 {
	gs := new(bn256.G1).ScalarBaseMult(secret)                 // G * secret
	ht := new(bn256.G1).ScalarMult(cm.protocolParameters.H, t) // H * t

	return new(bn256.G1).Add(gs, ht) // G * secret + H * t
}

// generatePolynomial generates a random polynomial over `Z_q` of a given degree.
// This function will generate a slice of `degree + 1` coefficients. Each value
// will be a random `big.Int` in range `(0, q)` where `q` is cardinality of
// alt_bn128 elliptic curve.
func generatePolynomial(degree int) ([]*big.Int, error) {
	generateCoefficient := func() (c *big.Int, err error) {
		for {
			c, err = crand.Int(crand.Reader, bn256.Order)
			if c.Sign() > 0 || err != nil {
				return
			}
		}
	}

	coefficients := make([]*big.Int, degree+1)
	for i := range coefficients {
		coefficient, err := generateCoefficient()
		if err != nil {
			return nil, err
		}

		coefficients[i] = coefficient
	}

	return coefficients, nil
}

// evaluateMemberShare calculates a share for given memberID.
//
// It calculates `s_j = Σ a_k * j^k mod q`for k in [0..T], where:
// - `a_k` is k coefficient
// - `j` is memberID
// - `T` is threshold
// - `q` is the order of cyclic group formed over the alt_bn128 curve
func (cm *CommittingMember) evaluateMemberShare(
	memberID member.MemberIndex, coefficients []*big.Int,
) *big.Int {
	result := big.NewInt(0)
	for k, a := range coefficients {
		result = new(big.Int).Mod(
			new(big.Int).Add(
				result,
				new(big.Int).Mul(
					a,
					pow(memberID, k),
				),
			),
			bn256.Order,
		)
	}
	return result
}

// VerifyReceivedSharesAndCommitmentsMessages verifies shares and commitments
// received in messages from other group members.
// It returns accusation message with ID of members for which verification failed.
//
// If cannot match commitments message with shares message for given sender then
// error is returned. Also, error is returned if the member does not have
// a symmetric encryption key established with sender of a message.
//
// All the received PeerSharesMessage should be validated before they are passed
// to this function. It should never happen that the message can't be decrypted
// by this function.
//
// See Phase 4 of the protocol specification.
func (cvm *CommitmentsVerifyingMember) VerifyReceivedSharesAndCommitmentsMessages(
	sharesMessages []*PeerSharesMessage,
	commitmentsMessages []*MemberCommitmentsMessage,
) (*SecretSharesAccusationsMessage, error) {
	for _, sharesMessage := range sharesMessages {
		cvm.evidenceLog.PutPeerSharesMessage(sharesMessage)
	}

	accusedMembersKeys := make(map[member.MemberIndex]*ephemeral.PrivateKey)
	for _, commitmentsMessage := range commitmentsMessages {
		// Find share message sent by the same member who sent commitment message
		sharesMessageFound := false
		for _, sharesMessage := range sharesMessages {
			if sharesMessage.senderID == commitmentsMessage.senderID {
				sharesMessageFound = true

				// If there is no symmetric key established with a sender of
				// the message, error is returned.
				symmetricKey, hasKey := cvm.symmetricKeys[sharesMessage.senderID]
				if !hasKey {
					return nil, fmt.Errorf(
						"no symmetric key for sender %v",
						sharesMessage.senderID,
					)
				}

				// Decrypt shares using symmetric key established with sender.
				// Since all the messages are validated prior to passing to this
				// function, decryption error should never happen.
				shareS, err := sharesMessage.decryptShareS(cvm.ID, symmetricKey) // s_ji
				if err != nil {
					return nil, fmt.Errorf(
						"could not decrypt share S [%v]",
						err,
					)
				}
				shareT, err := sharesMessage.decryptShareT(cvm.ID, symmetricKey) // t_ji
				if err != nil {
					return nil, fmt.Errorf(
						"could not decrypt share T [%v]",
						err,
					)
				}

				if !cvm.areSharesValidAgainstCommitments(
					shareS, // s_ji
					shareT, // t_ji
					commitmentsMessage.commitments, // C_j
					cvm.ID, // i
				) {
					accusedMembersKeys[commitmentsMessage.senderID] =
						cvm.ephemeralKeyPairs[commitmentsMessage.senderID].PrivateKey
					break
				}
				cvm.receivedValidSharesS[commitmentsMessage.senderID] = shareS
				cvm.receivedValidSharesT[commitmentsMessage.senderID] = shareT
				cvm.receivedValidPeerCommitments[commitmentsMessage.senderID] =
					commitmentsMessage.commitments

				break
			}
		}
		if !sharesMessageFound {
			return nil, fmt.Errorf("cannot find shares message from member %v",
				commitmentsMessage.senderID,
			)
		}
	}

	return &SecretSharesAccusationsMessage{
		senderID:           cvm.ID,
		accusedMembersKeys: accusedMembersKeys,
	}, nil
}

// areSharesValidAgainstCommitments verifies if commitments are valid for passed
// shares.
//
// The `j` member generated a polynomial with `k` coefficients before. Then they
// calculated a commitments to the polynomial's coefficients `C_j` and individual
// shares `s_ji` and `t_ji` with a polynomial for a member `i`. In this function
// the verifier checks if the shares are valid against the commitments.
//
// The verifier checks whether [GJKR 99] 1.(b) holds:
// `(g ^ s_ji) * (h ^ t_ji) mod p == Π (C_j[k] ^ (i^k)) mod p` for `k` in `[0..T]`
//
// What, using elliptic curve, is the same as:
// `G * s_ji + H * t_ji == Σ (C_j[k] * (i^k))` for `k` in `[0..T]`
func (cm *CommittingMember) areSharesValidAgainstCommitments(
	shareS, shareT *big.Int, // s_ji, t_ji
	commitments []*bn256.G1, // C_j
	memberID member.MemberIndex, // i
) bool {
	var sum *bn256.G1                // Σ (C_j[k] * (i^k)) for k in [0..T]
	for k, ck := range commitments { // k, C_j[k]
		ci := new(bn256.G1).ScalarMult(ck, pow(memberID, k)) // C_j[k] * (i^k)
		if sum == nil {
			sum = ci
		} else {
			sum = new(bn256.G1).Add(sum, ci)
		}
	}

	commitment := cm.calculateCommitment(shareS, shareT) // G * s_ji + H * t_ji

	return commitment.String() == sum.String()
}

// ResolveSecretSharesAccusationsMessages resolves complaints received in
// secret shares accusations messages. The member calls this function to judge
// which party of the dispute is lying.
//
// The current member cannot be a part of a dispute. If the current member is
// either an accuser or is accused, the accusation is ignored. The accused
// party cannot be a judge in its own case. From the other hand, the accuser has
// already performed the calculation in the previous phase which resulted in the
// accusation and waits now for a judgment from other players.
//
// This function needs to decrypt shares sent previously by the accused member
// to the accuser in an encrypted form. To do that it needs to recover a symmetric
// key used for data encryption. It takes private key revealed by the accuser
// and public key broadcasted by the accused and performs Elliptic Curve Diffie-
// Hellman operation between them.
//
// It returns IDs of members who should be disqualified. It will be an accuser
// if the validation shows that shares and commitments are valid, so the accusation
// was unfounded. Else it confirms that accused member misbehaved and their ID is
// added to the list.
//
// See Phase 5 of the protocol specification.
func (sjm *SharesJustifyingMember) ResolveSecretSharesAccusationsMessages(
	messages []*SecretSharesAccusationsMessage,
) ([]member.MemberIndex, error) {
	disqualifiedMembers := make([]member.MemberIndex, 0)
	for _, message := range messages {
		accuserID := message.senderID
		for accusedID, revealedAccuserPrivateKey := range message.accusedMembersKeys {
			if sjm.ID == accuserID || sjm.ID == accusedID {
				// The member cannot resolve the dispute in which it's involved.
				continue
			}

			symmetricKey, err := recoverSymmetricKey(
				sjm.evidenceLog,
				accusedID,
				accuserID,
				revealedAccuserPrivateKey,
			)
			if err != nil {
				// TODO Should we disqualify accuser/accused member here?
				return nil, fmt.Errorf("could not recover symmetric key [%v]", err)
			}

			shareS, shareT, err := recoverShares(
				sjm.evidenceLog,
				accusedID,
				accuserID,
				symmetricKey,
			)
			if err != nil {
				// TODO Should we disqualify accuser/accused member here?
				return nil, fmt.Errorf("could not decrypt shares [%v]", err)
			}

			if sjm.areSharesValidAgainstCommitments(
				shareS, shareT, // s_mj, t_mj
				sjm.receivedValidPeerCommitments[accusedID], // C_m
				accuserID, // j
			) {
				disqualifiedMembers = append(disqualifiedMembers, accuserID)
			} else {
				disqualifiedMembers = append(disqualifiedMembers, accusedID)
			}
		}
	}
	return disqualifiedMembers, nil
}

// Recover ephemeral symmetric key used to encrypt communication between sender
// and receiver assuming that receiver revealed its private ephemeral key.
//
// Finds ephemeral public key sent by sender to the receiver. Performs ECDH
// operation between sender's public key and receiver's private key to recover
// the ephemeral symmetric key.
func recoverSymmetricKey(
	evidenceLog evidenceLog,
	senderID, receiverID member.MemberIndex, receiverPrivateKey *ephemeral.PrivateKey,
) (ephemeral.SymmetricKey, error) {
	ephemeralPublicKeyMessage := evidenceLog.ephemeralPublicKeyMessage(senderID)
	if ephemeralPublicKeyMessage == nil {
		return nil, fmt.Errorf(
			"no ephemeral public key message for sender %v",
			senderID,
		)
	}

	senderPublicKey, ok := ephemeralPublicKeyMessage.ephemeralPublicKeys[receiverID]
	if !ok {
		return nil, fmt.Errorf(
			"no ephemeral public key generated for receiver %v",
			receiverID,
		)
	}

	return receiverPrivateKey.Ecdh(senderPublicKey), nil
}

// Recovers from the evidence log share S and share T sent by sender to the
// receiver.
//
// First it finds in the evidence log the Peer Shares Message sent by the sender
// to the receiver. Then it decrypts the decrypted shares with provided symmetric
// key.
func recoverShares(
	evidenceLog evidenceLog,
	senderID, receiverID member.MemberIndex, symmetricKey ephemeral.SymmetricKey,
) (*big.Int, *big.Int, error) {
	peerSharesMessage := evidenceLog.peerSharesMessage(senderID)
	if peerSharesMessage == nil {
		return nil, nil, fmt.Errorf(
			"no peer shares message for sender %v",
			senderID,
		)
	}

	shareS, err := peerSharesMessage.decryptShareS(receiverID, symmetricKey) // s_mj
	if err != nil {
		// TODO Should we disqualify accuser/accused member here?
		return nil, nil, fmt.Errorf("cannot decrypt share S [%v]", err)
	}
	shareT, err := peerSharesMessage.decryptShareT(receiverID, symmetricKey) // t_mj
	if err != nil {
		// TODO Should we disqualify accuser/accused member here?
		return nil, nil, fmt.Errorf("cannot decrypt share T [%v]", err)
	}

	return shareS, shareT, nil
}

// CombineMemberShares sums up all `s` shares intended for this member.
// Combines secret shares calculated by current member `i` for itself `s_ii`
// with shares calculated by peer members `j` for this member `s_ji`.
//
// `x_i = Σ s_ji mod q` for `j` in a group of players who passed secret shares
// accusations stage. `q` is the order of cyclic group formed over the alt_bn128
// curve.
//
// See Phase 6 of the protocol specification.
func (qm *QualifiedMember) CombineMemberShares() {
	combinedSharesS := qm.selfSecretShareS // s_ii
	for _, s := range qm.receivedValidSharesS {
		combinedSharesS = new(big.Int).Mod(
			new(big.Int).Add(combinedSharesS, s),
			bn256.Order,
		)
	}

	qm.groupPrivateKeyShare = combinedSharesS
}

// CalculatePublicKeySharePoints calculates public values for member's
// coefficients.
//
// It calculates:
// `A_k = g^a_k` for `k` in `[0..T]`.
//
// What, using elliptic curve, is the same as:
// `A_k = G * a_k` for `k` in `[0..T]`.
// where `G` is curve's generator.
//
// See Phase 7 of the protocol specification.
func (sm *SharingMember) CalculatePublicKeySharePoints() *MemberPublicKeySharePointsMessage {
	sm.publicKeySharePoints = make([]*bn256.G2, len(sm.secretCoefficients))
	for i, a := range sm.secretCoefficients {
		sm.publicKeySharePoints[i] = new(bn256.G2).ScalarBaseMult(a)
	}

	return &MemberPublicKeySharePointsMessage{
		senderID:             sm.ID,
		publicKeySharePoints: sm.publicKeySharePoints,
	}
}

// VerifyPublicKeySharePoints validates public key share points received in
// messages from peer group members.
// It returns accusation message with ID of members for which the verification
// failed.
//
// See Phase 8 of the protocol specification.
func (sm *SharingMember) VerifyPublicKeySharePoints(
	messages []*MemberPublicKeySharePointsMessage,
) (*PointsAccusationsMessage, error) {
	accusedMembersKeys := make(map[member.MemberIndex]*ephemeral.PrivateKey)
	// `product = Π (A_j[k] ^ (i^k)) mod p` for k in [0..T],
	// where: j is sender's ID, i is current member ID, T is threshold.
	for _, message := range messages {
		if !sm.isShareValidAgainstPublicKeySharePoints(
			sm.ID,
			sm.receivedValidSharesS[message.senderID],
			message.publicKeySharePoints,
		) {
			accusedMembersKeys[message.senderID] = sm.ephemeralKeyPairs[message.senderID].PrivateKey
			continue
		}
		sm.receivedValidPeerPublicKeySharePoints[message.senderID] = message.publicKeySharePoints
	}

	return &PointsAccusationsMessage{
		senderID:           sm.ID,
		accusedMembersKeys: accusedMembersKeys,
	}, nil
}

// isShareValidAgainstPublicKeySharePoints verifies if public key share points
// are valid for passed share S.
//
// The `j` member calculated public key share points for their polynomial
// coefficients and share `s_ji` with a polynomial for a member `i`. In this
// function the verifier checks if the public key share points are valid against
// the share S.
//
// The verifier checks whether [GJKR 99] 4.(b) holds:
// `g^s_ji mod p == Π (A_j[k] ^ (i^k)) mod p` for `k` in `[0..T]`
//
// What, using elliptic curve, is the same as:
// G * s_ji == Σ ( A_j[k] * (i^k) ) for `k` in `[0..T]`
func (sm *SharingMember) isShareValidAgainstPublicKeySharePoints(
	senderID member.MemberIndex, shareS *big.Int,
	publicKeySharePoints []*bn256.G2,
) bool {
	var sum *bn256.G2 // Σ ( A_j[k] * (i^k) ) for `k` in `[0..T]`
	for k, a := range publicKeySharePoints {
		aj := new(bn256.G2).ScalarMult(a, pow(senderID, k)) // A_j[k] * (i^k)
		if sum == nil {
			sum = aj
		} else {
			sum = new(bn256.G2).Add(sum, aj)
		}
	}

	gs := new(bn256.G2).ScalarBaseMult(shareS) // G * s_ji

	return gs.String() == sum.String()
}

// ResolvePublicKeySharePointsAccusationsMessages resolves a complaint received
// in points accusations messages. The member calls this function to judge
// which party of the dispute is lying.
//
// The current member cannot be a part of a dispute. If the current member is
// either an accuser or is accused, the accusation is ignored. The accused
// party cannot be a judge in its own case. From the other hand, the accuser has
// already performed the calculation in the previous phase which resulted in the
// accusation and waits now for a judgment from other players.
//
// This function needs to decrypt shares sent previously by the accused member
// to the accuser in an encrypted form. To do that it needs to recover a symmetric
// key used for data encryption. It takes private key revealed by the accuser
// and public key broadcasted by the accused and performs Elliptic Curve Diffie-
// Hellman operation between them.
//
// It returns IDs of members who should be disqualified. It will be an accuser
// if the validation shows that coefficients are valid, so the accusation was
// unfounded. Else it confirms that accused member misbehaved and their ID is
// added to the list.
//
// See Phase 9 of the protocol specification.
func (pjm *PointsJustifyingMember) ResolvePublicKeySharePointsAccusationsMessages(
	messages []*PointsAccusationsMessage,
) ([]member.MemberIndex, error) {
	disqualifiedMembers := make([]member.MemberIndex, 0)
	for _, message := range messages {
		accuserID := message.senderID
		for accusedID, revealedAccuserPrivateKey := range message.accusedMembersKeys {
			if pjm.ID == message.senderID || pjm.ID == accusedID {
				// The member cannot resolve the dispute in which it's involved.
				continue
			}

			evidenceLog := pjm.evidenceLog

			recoveredSymmetricKey, err := recoverSymmetricKey(
				evidenceLog,
				accusedID,
				accuserID,
				revealedAccuserPrivateKey,
			)
			if err != nil {
				// TODO Should we disqualify accuser/accused member here?
				return nil, fmt.Errorf("could not recover symmetric key [%v]", err)
			}

			shareS, _, err := recoverShares(
				evidenceLog,
				accusedID,
				accuserID,
				recoveredSymmetricKey,
			)
			if err != nil {
				// TODO Should we disqualify accuser/accused member here?
				return nil, fmt.Errorf("could not decrypt share S [%v]", err)
			}

			if pjm.isShareValidAgainstPublicKeySharePoints(
				message.senderID,
				shareS,
				pjm.receivedValidPeerPublicKeySharePoints[accusedID],
			) {
				// TODO The accusation turned out to be unfounded. Should we add accused
				// member's individual public key to receivedValidPeerPublicKeySharePoints?
				disqualifiedMembers = append(disqualifiedMembers, message.senderID)
				continue
			}
			disqualifiedMembers = append(disqualifiedMembers, accusedID)
		}
	}
	return disqualifiedMembers, nil
}

// RevealDisqualifiedMembersKeys reveals ephemeral private keys used to create
// an ephemeral symmetric key with disqualified members who share a group private
// key. The function filters members who sent valid share S in Phase 3 but were
// disqualified in Phase 9. It returns a message containing a map of ephemeral
// private key for each disqualified member sharing a group private key.
//
// See Phase 10 of the protocol specification.
func (rm *RevealingMember) RevealDisqualifiedMembersKeys() (
	*DisqualifiedEphemeralKeysMessage,
	error,
) {
	privateKeys := make(map[member.MemberIndex]*ephemeral.PrivateKey)

	for _, disqualifiedMemberID := range rm.disqualifiedSharingMembers() {
		ephemeralKeyPair, ok := rm.ephemeralKeyPairs[disqualifiedMemberID]
		if !ok {
			return nil, fmt.Errorf(
				"no ephemeral key pair for disqualified member %v",
				disqualifiedMemberID,
			)
		}
		privateKeys[disqualifiedMemberID] = ephemeralKeyPair.PrivateKey
	}

	return &DisqualifiedEphemeralKeysMessage{
		senderID:    rm.ID,
		privateKeys: privateKeys,
	}, nil
}

func (rm *RevealingMember) disqualifiedSharingMembers() []member.MemberIndex {
	disqualifiedMembersIDs := rm.group.disqualifiedMemberIDs

	// From disqualified members list filter those who provided valid shares in
	// Phase 3 and are sharing the group private key.
	disqualifiedSharingMembers := make([]member.MemberIndex, 0)
	for _, disqualifiedMemberID := range disqualifiedMembersIDs {
		if _, ok := rm.receivedValidSharesS[disqualifiedMemberID]; ok {
			disqualifiedSharingMembers = append(
				disqualifiedSharingMembers,
				disqualifiedMemberID,
			)
		}
	}
	return disqualifiedSharingMembers
}

// ReconstructDisqualifiedIndividualKeys reconstructs individual private key `z_m` and  public
// key `y_m` of disqualified members `m`. To do that, it first needs to recover
// shares calculated by disqualified members `m` in Phase 3 for other members `k`.
// The shares were encrypted before broadcast, so ephemeral symmetric key needs
// to be recovered. This requires messages containing ephemeral private key
// revealed by member `k` used in communication with disqualified member `m`.
//
// See Phase 11 of the protocol specification.
func (rm *ReconstructingMember) ReconstructDisqualifiedIndividualKeys(
	messages []*DisqualifiedEphemeralKeysMessage,
) error {
	revealedDisqualifiedShares, err := rm.recoverDisqualifiedShares(messages)
	if err != nil {
		return fmt.Errorf("recovering disqualified shares failed [%v]", err)
	}
	rm.reconstructIndividualPrivateKeys(revealedDisqualifiedShares) // z_m
	rm.reconstructIndividualPublicKeys()                            // y_m
	return nil
}

// Recover shares `s_mk` calculated by members `m` disqualified in Phase 9.
// The shares were evaluated in Phase 3 by `m` for other members `k` and
// broadcasted in an encrypted fashion, hence reconstructing member has to
// recover a symmetric key to decode the shares messages. It returns a slice
// containing shares `s_mk` recovered for each member `m` whose ephemeral key
// was revealed in provided DisqualifiedMembersKeysMessage.
func (rm *ReconstructingMember) recoverDisqualifiedShares(
	messages []*DisqualifiedEphemeralKeysMessage,
) ([]*disqualifiedShares, error) {
	revealedDisqualifiedShares := make([]*disqualifiedShares, 0)

	// For disqualified member `m` add shares `s_mk` the member calculated for
	// other members `k` who revealed the ephemeral key.
	addShare := func(
		disqualifiedMemberID, revealingMemberID member.MemberIndex, // m, k
		shareS *big.Int, // s_mk
	) {
		// If a `disqualifiedShares` entry already exists in the slice for given
		// disqualified member add the share.
		for _, disqualifiedShares := range revealedDisqualifiedShares {
			if disqualifiedShares.disqualifiedMemberID == disqualifiedMemberID {
				disqualifiedShares.peerSharesS[revealingMemberID] = shareS
				return
			}
		}

		// When a `disqualifiedShares` entry doesn't exist yet in the slice for given
		// disqualified member initialize it with the share.
		newDisqualifiedShares := &disqualifiedShares{
			disqualifiedMemberID: disqualifiedMemberID,
			peerSharesS:          make(map[member.MemberIndex]*big.Int),
		}
		newDisqualifiedShares.peerSharesS[revealingMemberID] = shareS

		revealedDisqualifiedShares = append(
			revealedDisqualifiedShares,
			newDisqualifiedShares,
		)
	}

	for _, message := range messages {
		revealingMemberID := message.senderID
		for disqualifiedMemberID, revealedPrivateKey := range message.privateKeys {
			publicKey := rm.evidenceLog.ephemeralPublicKeyMessage(revealingMemberID).
				ephemeralPublicKeys[disqualifiedMemberID]
			if !publicKey.IsKeyMatching(revealedPrivateKey) {
				fmt.Printf("invalid private key for public key from member %v\n", revealingMemberID)
				rm.group.MarkMemberAsDisqualified(revealingMemberID)
				continue
			}

			recoveredSymmetricKey, err := recoverSymmetricKey(
				rm.evidenceLog,
				disqualifiedMemberID, // m
				revealingMemberID,    // k
				revealedPrivateKey,
			)
			if err != nil {
				fmt.Printf("cannot recover symmetric key [%v]", err)
				// TODO Disqualify the revealing member
				continue
			}

			shareS, _, err := recoverShares(
				rm.evidenceLog,
				disqualifiedMemberID,  // m
				revealingMemberID,     // k
				recoveredSymmetricKey, // s_mk
			)
			if err != nil {
				fmt.Printf("cannot decrypt share S [%v]", err)
				// TODO Disqualify the revealing member
				continue
			}

			addShare(disqualifiedMemberID, revealingMemberID, shareS)
		}
	}
	return revealedDisqualifiedShares, nil
}

// disqualifiedShares contains shares `s_mk` calculated by the disqualified
// member `m` for peer members `k`. The shares were revealed due to disqualification
// of the member `m` from the protocol execution.
type disqualifiedShares struct {
	disqualifiedMemberID member.MemberIndex              // m
	peerSharesS          map[member.MemberIndex]*big.Int // <k, s_mk>
}

// reconstructIndividualPrivateKeys reconstructs disqualified members' individual
// private keys `z_m` from provided revealed shares calculated by disqualified
// members for peer members.
//
// Function need to be executed for qualified members that presented valid shares
// and commitments and were approved for Phase 6 but were disqualified on public
// key shares validation stage (Phase 9).
//
// It stores a map of reconstructed individual private keys for each disqualified
// member in a current member's reconstructedIndividualPrivateKeys field:
// <disqualifiedMemberID, privateKeyShare>
func (rm *ReconstructingMember) reconstructIndividualPrivateKeys(
	revealedDisqualifiedShares []*disqualifiedShares,
) {
	rm.reconstructedIndividualPrivateKeys = make(map[member.MemberIndex]*big.Int, len(revealedDisqualifiedShares))

	for _, ds := range revealedDisqualifiedShares { // for each disqualified member
		// Reconstruct individual private key `z_m = Σ (s_mk * a_mk) mod q` where:
		// - `z_m` is disqualified member's individual private key
		// - `s_mk` is a share calculated by disqualified member `m` for peer member `k`
		// - `a_mk` is lagrange coefficient for peer member k (see below)
		individualPrivateKey := big.NewInt(0)
		// Get IDs of all peer members from disqualified shares.
		var peerIDs []member.MemberIndex
		for k := range ds.peerSharesS {
			peerIDs = append(peerIDs, k)
		}
		// For each peerID `k` and peerShareS `s_mk` calculate `s_mk * a_mk`
		for peerID, peerShareS := range ds.peerSharesS {
			// a_mk
			lagrangeCoefficient := rm.calculateLagrangeCoefficient(peerID, peerIDs)

			// Σ (s_mk * a_mk) mod q
			individualPrivateKey = new(big.Int).Mod(
				new(big.Int).Add(
					individualPrivateKey,
					// s_mk * a_mk
					new(big.Int).Mul(peerShareS, lagrangeCoefficient),
				),
				bn256.Order,
			)
		}
		// <m, z_m>
		rm.reconstructedIndividualPrivateKeys[ds.disqualifiedMemberID] =
			individualPrivateKey
	}
}

// Calculates Lagrange coefficient `a_mk` for member `k` in a group of members.
//
// `a_mk = Π (l / (l - k)) mod q` where:
// - `a_mk` is a lagrange coefficient for the member `k`,
// - `l` are IDs of members who provided shares,
// - `q` is an order of alt_bn128 elliptic curve
// and `l != k`.
func (rm *ReconstructingMember) calculateLagrangeCoefficient(memberID member.MemberIndex, groupMembersIDs []member.MemberIndex) *big.Int {
	lagrangeCoefficient := big.NewInt(1)
	// For each otherID `l` in groupMembersIDs:
	for _, otherID := range groupMembersIDs {
		if otherID != memberID { // l != k
			// l / (l - k)
			quotient := new(big.Int).Mod(
				new(big.Int).Mul(
					big.NewInt(int64(otherID)),
					new(big.Int).ModInverse(
						new(big.Int).Sub(
							otherID.Int(),
							memberID.Int(),
						),
						bn256.Order,
					),
				),
				bn256.Order,
			)

			// Π (l / (l - k)) mod q
			lagrangeCoefficient = new(big.Int).Mod(
				new(big.Int).Mul(
					lagrangeCoefficient, quotient,
				),
				bn256.Order,
			)
		}
	}
	return lagrangeCoefficient // a_mk
}

// reconstructIndividualPublicKeys calculates and stores individual public keys
// `y_m` from reconstructed individual private keys `z_m`.
//
// Public key is calculated as `g^privateKey mod p` what, using elliptic curve,
// is the same as `G * privateKey`.
//
// See Phase 11 of the protocol specification.
func (rm *ReconstructingMember) reconstructIndividualPublicKeys() {
	rm.reconstructedIndividualPublicKeys = make(
		map[member.MemberIndex]*bn256.G2,
		len(rm.reconstructedIndividualPrivateKeys),
	)
	for memberID, individualPrivateKey := range rm.reconstructedIndividualPrivateKeys {
		// y_m = G * z_m
		individualPublicKey := new(bn256.G2).ScalarBaseMult(individualPrivateKey)
		rm.reconstructedIndividualPublicKeys[memberID] = individualPublicKey
	}
}

func pow(id member.MemberIndex, y int) *big.Int {
	return new(big.Int).Exp(id.Int(), big.NewInt(int64(y)), nil)
}

// CombineGroupPublicKey calculates a group public key by combining individual
// public keys. Group public key is calculated as a product of individual public
// keys of all group members including member themself.
//
// `Y = Π y_j mod p` for `j`, where `y_j` is individual public key of each
// qualified group member. With elliptic curve, it is: `Y = Σ y_j`.
//
// This function combines individual public keys of all Qualified Members who were
// approved for Phase 6. Three categories of individual public keys are considered:
// 1. Current member's individual public key.
// 2. Peer members' individual public keys - for members who passed a public key
//    share points validation in Phase 8 and accusations resolution in Phase 9 and
//    are still active group members.
// 3. Disqualified members' individual public keys - for members who were disqualified
//    in Phase 9 and theirs individual private and public keys were reconstructed
//    in Phase 11.
//
// See Phase 12 of the protocol specification.
func (rm *CombiningMember) CombineGroupPublicKey() {
	// Current member's individual public key `A_i0`.
	groupPublicKey := rm.individualPublicKey()

	// Add received peer group members' individual public keys `A_j0`.
	for _, peerPublicKey := range rm.receivedValidPeerIndividualPublicKeys() {
		groupPublicKey = new(bn256.G2).Add(groupPublicKey, peerPublicKey)
	}

	// Add reconstructed disqualified members' individual public keys `G * z_m`.
	for _, peerPublicKey := range rm.reconstructedIndividualPublicKeys {
		groupPublicKey = new(bn256.G2).Add(groupPublicKey, peerPublicKey)

	}

	rm.groupPublicKey = groupPublicKey
}
