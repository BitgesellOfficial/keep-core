package chain

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

// DKGResult is a result of distributed key generation protocol.
//
// Success means that the protocol execution finished with acceptable number of
// disqualified or inactive members. The group of remaining members should be
// added to the signing groups for the threshold relay.
//
// Failure means that the group creation could not finish, due to either the number
// of inactive or disqualified participants, or the presented results being
// disputed in a way where the correct outcome cannot be ascertained.
type DKGResult struct {
	// Result type of the protocol execution. True if success, false if failure.
	Success bool
	// Group public key generated by protocol execution, empty if the protocol failed.
	GroupPublicKey []byte
	// Disqualified members are represented as a slice of bytes for optimizing
	// on-chain storage. The length of the slice, and ordering of the members is
	// the same as the members group. Disqualified members are marked as 0x01,
	// non-disqualified members as 0x00.
	Disqualified []byte
	// Inactive members are represented as a slice of bytes for optimizing
	// on-chain storage. The length of the slice, and ordering of the members is
	// the same as the members group. Inactive members are marked as 0x01,
	// active members as 0x00.
	Inactive []byte
}

// DKGResultHash is a Keccak-256 hash of DKG Result.
type DKGResultHash [32]byte

// DKGResultsVotes is a map of votes for each DKG Result.
type DKGResultsVotes map[DKGResultHash]int

// Equals checks if two DKG results are equal.
func (r *DKGResult) Equals(r2 *DKGResult) bool {
	if r == nil || r2 == nil {
		return r == r2
	}
	if r.Success != r2.Success {
		return false
	}
	if !bytes.Equal(r.GroupPublicKey, r2.GroupPublicKey) {
		return false
	}
	if !bytes.Equal(r.Disqualified, r2.Disqualified) {
		return false
	}
	if !bytes.Equal(r.Inactive, r2.Inactive) {
		return false
	}
	return true
}

// Hash returns Keccak-256 hash of the DKG result.
func (r *DKGResult) Hash() (dkgResultHash DKGResultHash, err error) {
	encodedDKGResult, err := r.encode()

	h := sha3.NewKeccak256()
	h.Write(encodedDKGResult)
	h.Sum(dkgResultHash[:0])

	return
}

// encode returns DKG result encoded to the format described by Solidity Contract
// Application Binary Interface (ABI).
func (r *DKGResult) encode() ([]byte, error) {
	boolType, err := abi.NewType("bool")
	if err != nil {
		return nil, fmt.Errorf("bool type creation failed [%v]", err)
	}
	bytesType, err := abi.NewType("bytes")
	if err != nil {
		return nil, fmt.Errorf("bytes type creation failed [%v]", err)
	}

	arguments := abi.Arguments{
		{Type: boolType},
		{Type: bytesType},
		{Type: bytesType},
		{Type: bytesType},
	}

	bytes, err := arguments.Pack(r.Success, r.GroupPublicKey, r.Disqualified, r.Inactive)
	if err != nil {
		return nil, fmt.Errorf("encoding failed [%v]", err)
	}

	return bytes, nil
}
