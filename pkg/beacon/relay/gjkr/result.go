package gjkr

import (
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
	"github.com/keep-network/keep-core/pkg/beacon/relay/member"
)

// Result of distributed key generation protocol.
type Result struct {
	// Group public key generated by protocol execution.
	GroupPublicKey *bn256.G2
	// Share of the group private key. It is used for signing and should never
	// be revealed publicly.
	GroupPrivateKeyShare *big.Int
	// Disqualified members IDs.
	Disqualified []member.MemberIndex
	// Inactive members IDs.
	Inactive []member.MemberIndex
}

// Equals checks if two results are equal.
// TODO: Check if we still need this function. If not remove.
func (r *Result) Equals(r2 *Result) bool {
	if r == nil || r2 == nil {
		return r == r2
	}

	if !publicKeysEqual(r.GroupPublicKey, r2.GroupPublicKey) {
		return false
	}

	if !memberIDSlicesEqual(r.Disqualified, r2.Disqualified) {
		return false
	}
	if !memberIDSlicesEqual(r.Inactive, r2.Inactive) {
		return false
	}

	return true
}

// publicKeysEqual checks if two public keys are equal.
func publicKeysEqual(expectedKey *bn256.G2, actualKey *bn256.G2) bool {
	if expectedKey != nil && actualKey != nil {
		return expectedKey.String() == actualKey.String()
	}
	return expectedKey == actualKey
}

// memberIDSlicesEqual checks if two slices of MemberIDs are equal. Slices need
// to have the same length and have the same order of entries.
func memberIDSlicesEqual(expectedSlice []member.MemberIndex, actualSlice []member.MemberIndex) bool {
	if len(expectedSlice) != len(actualSlice) {
		return false
	}

	for i := range expectedSlice {
		if expectedSlice[i] != actualSlice[i] {
			return false
		}
	}
	return true
}
