package tbtc

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"reflect"
	"sync"

	"github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-core/pkg/chain/local_v1"
	"github.com/keep-network/keep-core/pkg/operator"
	"github.com/keep-network/keep-core/pkg/protocol/group"
	"github.com/keep-network/keep-core/pkg/subscription"
	"github.com/keep-network/keep-core/pkg/tecdsa/dkg"
	"golang.org/x/crypto/sha3"
)

const localChainOperatorID = chain.OperatorID(1)

var errNilDKGResult = fmt.Errorf("nil DKG result")

type localChain struct {
	dkgResultSubmissionHandlersMutex sync.Mutex
	dkgResultSubmissionHandlers      map[int]func(submission *DKGResultSubmittedEvent)

	dkgResultApprovalHandlersMutex sync.Mutex
	dkgResultApprovalHandlers      map[int]func(submission *DKGResultApprovedEvent)

	dkgMutex       sync.Mutex
	dkgState       DKGState
	dkgResult      *DKGChainResult
	dkgResultValid bool

	blockCounter       chain.BlockCounter
	chainConfig        *ChainConfig
	operatorPrivateKey *operator.PrivateKey

	operatorsIDs map[chain.Address]chain.OperatorID
}

func (lc *localChain) GetConfig() *ChainConfig {
	return lc.chainConfig
}

func (lc *localChain) BlockCounter() (chain.BlockCounter, error) {
	return lc.blockCounter, nil
}

func (lc *localChain) Signing() chain.Signing {
	return local_v1.NewSigner(lc.operatorPrivateKey)
}

func (lc *localChain) OperatorKeyPair() (
	*operator.PrivateKey,
	*operator.PublicKey,
	error,
) {
	return lc.operatorPrivateKey, &lc.operatorPrivateKey.PublicKey, nil
}

func (lc *localChain) OperatorToStakingProvider() (chain.Address, bool, error) {
	panic("unsupported")
}

func (lc *localChain) EligibleStake(stakingProvider chain.Address) (*big.Int, error) {
	panic("unsupported")
}

func (lc *localChain) IsPoolLocked() (bool, error) {
	panic("unsupported")
}

func (lc *localChain) IsOperatorInPool() (bool, error) {
	panic("unsupported")
}

func (lc *localChain) IsOperatorUpToDate() (bool, error) {
	panic("unsupported")
}

func (lc *localChain) JoinSortitionPool() error {
	panic("unsupported")
}

func (lc *localChain) UpdateOperatorStatus() error {
	panic("unsupported")
}

func (lc *localChain) IsEligibleForRewards() (bool, error) {
	panic("unsupported")
}

func (lc *localChain) CanRestoreRewardEligibility() (bool, error) {
	panic("unsupported")
}

func (lc *localChain) RestoreRewardEligibility() error {
	panic("unsupported")
}

func (lc *localChain) IsChaosnetActive() (bool, error) {
	panic("unsupported")
}

func (lc *localChain) IsBetaOperator() (bool, error) {
	panic("unsupported")
}

func (lc *localChain) GetOperatorID(
	operatorAddress chain.Address,
) (chain.OperatorID, error) {
	return lc.operatorsIDs[operatorAddress], nil
}

func (lc *localChain) SelectGroup() (*GroupSelectionResult, error) {
	panic("not implemented")
}

func (lc *localChain) OnDKGStarted(
	handler func(event *DKGStartedEvent),
) subscription.EventSubscription {
	panic("unsupported")
}

func (lc *localChain) OnDKGResultSubmitted(
	handler func(event *DKGResultSubmittedEvent),
) subscription.EventSubscription {
	lc.dkgResultSubmissionHandlersMutex.Lock()
	defer lc.dkgResultSubmissionHandlersMutex.Unlock()

	handlerID := local_v1.GenerateHandlerID()
	lc.dkgResultSubmissionHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lc.dkgResultSubmissionHandlersMutex.Lock()
		defer lc.dkgResultSubmissionHandlersMutex.Unlock()

		delete(lc.dkgResultSubmissionHandlers, handlerID)
	})
}

func (lc *localChain) OnDKGResultApproved(
	handler func(event *DKGResultApprovedEvent),
) subscription.EventSubscription {
	lc.dkgResultApprovalHandlersMutex.Lock()
	defer lc.dkgResultApprovalHandlersMutex.Unlock()

	handlerID := local_v1.GenerateHandlerID()
	lc.dkgResultApprovalHandlers[handlerID] = handler

	return subscription.NewEventSubscription(func() {
		lc.dkgResultApprovalHandlersMutex.Lock()
		defer lc.dkgResultApprovalHandlersMutex.Unlock()

		delete(lc.dkgResultApprovalHandlers, handlerID)
	})
}

func (lc *localChain) startDKG() error {
	lc.dkgMutex.Lock()
	defer lc.dkgMutex.Unlock()

	if lc.dkgState != Idle {
		return fmt.Errorf("DKG not idle")
	}

	lc.dkgState = AwaitingResult

	return nil
}

func (lc *localChain) SubmitDKGResult(
	memberIndex group.MemberIndex,
	tecdsaDkgResult *dkg.Result,
	signatures map[group.MemberIndex][]byte,
	groupSelectionResult *GroupSelectionResult,
) error {
	lc.dkgResultSubmissionHandlersMutex.Lock()
	defer lc.dkgResultSubmissionHandlersMutex.Unlock()

	lc.dkgMutex.Lock()
	defer lc.dkgMutex.Unlock()

	if lc.dkgState != AwaitingResult {
		return fmt.Errorf("not awaiting DKG result")
	}

	blockNumber, err := lc.blockCounter.CurrentBlock()
	if err != nil {
		return fmt.Errorf("failed to get the current block")
	}

	groupPublicKeyBytes, err := tecdsaDkgResult.GroupPublicKeyBytes()
	if err != nil {
		return fmt.Errorf(
			"failed to extract group public key bytes from the result [%v]",
			err,
		)
	}

	signingMembersIndexes := make([]group.MemberIndex, 0)
	signaturesConcatenation := make([]byte, 0)
	for memberIndex, signature := range signatures {
		signingMembersIndexes = append(signingMembersIndexes, memberIndex)
		signaturesConcatenation = append(signaturesConcatenation, signature...)
	}

	operatingMembersIndexes := tecdsaDkgResult.Group.OperatingMemberIndexes()
	operatingOperatorsIDsBytes := make([]byte, 0)
	for _, operatingMemberID := range operatingMembersIndexes {
		operatorIDBytes := make([]byte, 4)
		operatorID := groupSelectionResult.OperatorsIDs[operatingMemberID-1]
		binary.BigEndian.PutUint32(operatorIDBytes, operatorID)

		operatingOperatorsIDsBytes = append(
			operatingOperatorsIDsBytes,
			operatorIDBytes...,
		)
	}

	result := &DKGChainResult{
		SubmitterMemberIndex:     memberIndex,
		GroupPublicKey:           groupPublicKeyBytes,
		MisbehavedMembersIndexes: tecdsaDkgResult.MisbehavedMembersIndexes(),
		Signatures:               signaturesConcatenation,
		SigningMembersIndexes:    signingMembersIndexes,
		Members:                  groupSelectionResult.OperatorsIDs,
		MembersHash:              sha3.Sum256(operatingOperatorsIDsBytes),
	}

	resultHash := computeTestDkgResultHash(result)

	for _, handler := range lc.dkgResultSubmissionHandlers {
		handler(&DKGResultSubmittedEvent{
			Seed:        nil,
			ResultHash:  resultHash,
			Result:      result,
			BlockNumber: blockNumber,
		})
	}

	lc.dkgState = Challenge
	lc.dkgResult = result
	lc.dkgResultValid = true

	return nil
}

func (lc *localChain) GetDKGState() (DKGState, error) {
	lc.dkgMutex.Lock()
	defer lc.dkgMutex.Unlock()

	return lc.dkgState, nil
}

func (lc *localChain) CalculateDKGResultHash(
	startBlock uint64,
	tecdsaDkgResult *dkg.Result,
) (dkg.ResultHash, error) {
	if tecdsaDkgResult == nil {
		return dkg.ResultHash{}, errNilDKGResult
	}

	encodedDKGResult := fmt.Sprint(tecdsaDkgResult)
	dkgResultHash := dkg.ResultHash(
		sha3.Sum256([]byte(encodedDKGResult)),
	)
	return dkgResultHash, nil
}

func (lc *localChain) IsDKGResultValid(dkgResult *DKGChainResult) (bool, error) {
	lc.dkgMutex.Lock()
	defer lc.dkgMutex.Unlock()

	if lc.dkgState != Challenge {
		return false, fmt.Errorf("not in DKG result challenge period")
	}

	if !reflect.DeepEqual(dkgResult, lc.dkgResult) {
		return false, fmt.Errorf("result does not match the submitted one")
	}

	return lc.dkgResultValid, nil
}

func (lc *localChain) invalidateDKGResult(dkgResult *DKGChainResult) error {
	lc.dkgMutex.Lock()
	defer lc.dkgMutex.Unlock()

	if lc.dkgState != Challenge {
		return fmt.Errorf("not in DKG result challenge period")
	}

	if !reflect.DeepEqual(dkgResult, lc.dkgResult) {
		return fmt.Errorf("result does not match the submitted one")
	}

	lc.dkgResultValid = false

	return nil
}

func (lc *localChain) ChallengeDKGResult(dkgResult *DKGChainResult) error {
	lc.dkgMutex.Lock()
	defer lc.dkgMutex.Unlock()

	if lc.dkgState != Challenge {
		return fmt.Errorf("not in DKG result challenge period")
	}

	if !reflect.DeepEqual(dkgResult, lc.dkgResult) {
		return fmt.Errorf("result does not match the submitted one")
	}

	if lc.dkgResultValid {
		return fmt.Errorf("submitted result is valid")
	}

	lc.dkgState = AwaitingResult
	lc.dkgResult = nil
	lc.dkgResultValid = false

	return nil
}

func (lc *localChain) ApproveDKGResult(dkgResult *DKGChainResult) error {
	lc.dkgResultApprovalHandlersMutex.Lock()
	defer lc.dkgResultApprovalHandlersMutex.Unlock()

	lc.dkgMutex.Lock()
	defer lc.dkgMutex.Unlock()

	if lc.dkgState != Challenge {
		return fmt.Errorf("not in DKG result challenge period")
	}

	if !reflect.DeepEqual(dkgResult, lc.dkgResult) {
		return fmt.Errorf("result does not match the submitted one")
	}

	if !lc.dkgResultValid {
		return fmt.Errorf("submitted result is invalid")
	}

	blockNumber, err := lc.blockCounter.CurrentBlock()
	if err != nil {
		return fmt.Errorf("failed to get the current block")
	}

	for _, handler := range lc.dkgResultApprovalHandlers {
		handler(&DKGResultApprovedEvent{
			ResultHash:  computeTestDkgResultHash(dkgResult),
			Approver:    "",
			BlockNumber: blockNumber,
		})
	}

	lc.dkgState = Idle
	lc.dkgResult = nil
	lc.dkgResultValid = false

	return nil
}

func (lc *localChain) DKGParameters() (*DKGParameters, error) {
	return &DKGParameters{
		SubmissionTimeoutBlocks:       10,
		ChallengePeriodBlocks:         20,
		ApprovePrecedencePeriodBlocks: 5,
	}, nil
}

func (lc *localChain) OnHeartbeatRequested(
	handler func(event *HeartbeatRequestedEvent),
) subscription.EventSubscription {
	panic("unsupported")
}

func (lc *localChain) operator() (chain.OperatorID, chain.Address, error) {
	_, operatorPublicKey, err := lc.OperatorKeyPair()
	if err != nil {
		return 0, "", err
	}

	operatorAddress, err := lc.Signing().PublicKeyToAddress(operatorPublicKey)
	if err != nil {
		return 0, "", err
	}

	return localChainOperatorID, operatorAddress, nil
}

// Connect sets up the local chain.
func Connect(
	groupSize int,
	groupQuorum int,
	honestThreshold int,
) *localChain {
	operatorPrivateKey, _, err := operator.GenerateKeyPair(local_v1.DefaultCurve)
	if err != nil {
		panic(err)
	}

	return ConnectWithKey(
		groupSize,
		groupQuorum,
		honestThreshold,
		operatorPrivateKey,
	)
}

// ConnectWithKey sets up the local chain using the provided operator private
// key.
func ConnectWithKey(
	groupSize int,
	groupQuorum int,
	honestThreshold int,
	operatorPrivateKey *operator.PrivateKey,
) *localChain {
	blockCounter, _ := local_v1.BlockCounter()

	chainConfig := &ChainConfig{
		GroupSize:       groupSize,
		GroupQuorum:     groupQuorum,
		HonestThreshold: honestThreshold,
	}

	localChain := &localChain{
		dkgResultSubmissionHandlers: make(
			map[int]func(submission *DKGResultSubmittedEvent),
		),
		dkgResultApprovalHandlers: make(
			map[int]func(submission *DKGResultApprovedEvent),
		),
		blockCounter:       blockCounter,
		chainConfig:        chainConfig,
		operatorPrivateKey: operatorPrivateKey,
		operatorsIDs:       make(map[chain.Address]chain.OperatorID),
	}

	operatorID, operatorAddress, _ := localChain.operator()
	localChain.operatorsIDs[operatorAddress] = operatorID

	return localChain
}

func computeTestDkgResultHash(result *DKGChainResult) [32]byte {
	return sha3.Sum256(result.GroupPublicKey)
}
