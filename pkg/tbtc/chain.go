package tbtc

import (
	"math/big"

	"github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-core/pkg/ecdsa/dkg"
	"github.com/keep-network/keep-core/pkg/operator"
	"github.com/keep-network/keep-core/pkg/protocol/group"
	"github.com/keep-network/keep-core/pkg/sortition"
	"github.com/keep-network/keep-core/pkg/subscription"
)

// GroupSelectionChain defines the subset of the TBTC chain interface that
// pertains to the group selection activities.
type GroupSelectionChain interface {
	// SelectGroup returns the group members for the group generated by
	// the given seed. This function can return an error if the beacon chain's
	// state does not allow for group selection at the moment.
	SelectGroup(seed *big.Int) ([]chain.Address, error)
}

// DistributedKeyGenerationChain defines the subset of the TBTC chain
// interface that pertains specifically to group formation's distributed key
// generation process.
type DistributedKeyGenerationChain interface {
	// OnDKGStarted registers a callback that is invoked when an on-chain
	// notification of the DKG process start is seen.
	OnDKGStarted(
		func(event *dkg.DKGStartedEvent),
	) subscription.EventSubscription
	// SubmitDKGResult sends DKG result to a chain, along with signatures over
	// result hash from group participants supporting the result.
	// Signatures over DKG result hash are collected in a map keyed by signer's
	// member index.
	SubmitDKGResult(
		participantIndex group.MemberIndex,
		dkgResult *dkg.Result,
		signatures map[group.MemberIndex][]byte,
	) error
	// OnDKGResultSubmitted registers a callback that is invoked when an on-chain
	// notification of a new, valid submitted result is seen.
	OnDKGResultSubmitted(
		func(event *dkg.DKGResultSubmissionEvent),
	) subscription.EventSubscription
	// CalculateDKGResultHash calculates 256-bit hash of DKG result in standard
	// specific for the chain. Operation is performed off-chain.
	CalculateDKGResultHash(result *dkg.Result) (dkg.ResultHash, error)
}

// GroupSelectionInterface defines the subset of the beacon chain interface that
// pertains to the group selection activities.
type GroupSelectionInterface interface {
	// SelectGroup returns the group members for the group generated by
	// the given seed. This function can return an error if the beacon chain's
	// state does not allow for group selection at the moment.
	SelectGroup(seed *big.Int) ([]chain.Address, error)
}

// GroupRegistrationInterface defines the subset of the beacon chain interface
// that pertains to the group registration activities.
type GroupRegistrationInterface interface {
	// OnGroupRegistered is a callback that is invoked when an on-chain
	// notification of a new, valid group being registered is seen.
	OnGroupRegistered(
		func(groupRegistration *GroupRegistrationEvent),
	) subscription.EventSubscription
	// IsGroupRegistered checks if group with the given public key is registered
	// on-chain.
	IsGroupRegistered(groupPublicKey []byte) (bool, error)
	// IsStaleGroup checks if a group with the given public key is considered
	// as stale on-chain. Group is considered as stale if it is expired and when
	// its expiration time and potentially executed operation timeout are both
	// in the past. Stale group is never selected by the chain to any new
	// operation.
	IsStaleGroup(groupPublicKey []byte) (bool, error)
}

// GroupInterface defines the subset of the beacon chain interface that pertains
// specifically to the group management.
type GroupInterface interface {
	GroupSelectionInterface
	GroupRegistrationInterface
}

// GroupRegistrationEvent represents an event of registering a new group with the
// given public key.
// TODO: Adjust to the v2 RandomBeacon contract and rename to GroupRegistered.
type GroupRegistrationEvent struct {
	GroupPublicKey []byte

	BlockNumber uint64
}

// Chain represents the interface that the TBTC module expects to interact
// with the anchoring blockchain on.
type Chain interface {
	// GetConfig returns the expected configuration of the TBTC module.
	GetConfig() *ChainConfig
	// BlockCounter returns the chain's block counter.
	BlockCounter() (chain.BlockCounter, error)
	// Signing returns the chain's signer.
	Signing() chain.Signing
	// OperatorKeyPair returns the key pair of the operator assigned to this
	// chain handle.
	OperatorKeyPair() (*operator.PrivateKey, *operator.PublicKey, error)

	sortition.Chain
	GroupInterface
	GroupSelectionChain
	DistributedKeyGenerationChain
}

// ChainConfig contains the config data needed for the TBTC to operate.
type ChainConfig struct {
	// GroupSize is the size of a group in TBTC.
	GroupSize int
	// HonestThreshold is the minimum number of active participants behaving
	// according to the protocol needed to generate a signature.
	HonestThreshold int
	// ResultPublicationBlockStep is the duration (in blocks) that has to pass
	// before group member with the given index is eligible to submit the
	// result.
	// Nth player becomes eligible to submit the result after
	// T_dkg + (N-1) * T_step
	// where T_dkg is time for phases 1-12 to complete and T_step is the result
	// publication block step.
	ResultPublicationBlockStep uint64
}

// DishonestThreshold is the maximum number of misbehaving participants for
// which it is still possible to generate a signature.
// Misbehaviour is any misconduct to the protocol, including inactivity.
func (cc *ChainConfig) DishonestThreshold() int {
	return cc.GroupSize - cc.HonestThreshold
}
