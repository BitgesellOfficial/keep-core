package tbtcpg

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/keep-network/keep-core/pkg/bitcoin"
	"github.com/keep-network/keep-core/pkg/chain"
	"github.com/keep-network/keep-core/pkg/tbtc"
)

type depositParameters = struct {
	dustThreshold      uint64
	treasuryFeeDivisor uint64
	txMaxFee           uint64
	revealAheadPeriod  uint32
}

type redemptionParameters = struct {
	dustThreshold                   uint64
	treasuryFeeDivisor              uint64
	txMaxFee                        uint64
	txMaxTotalFee                   uint64
	timeout                         uint32
	timeoutSlashingAmount           *big.Int
	timeoutNotifierRewardMultiplier uint32
}

type walletParameters = struct {
	creationPeriod        uint32
	creationMinBtcBalance uint64
	creationMaxBtcBalance uint64
	closureMinBtcBalance  uint64
	maxAge                uint32
	maxBtcTransfer        uint64
	closingPeriod         uint32
}

type movingFundsParameters = struct {
	txMaxTotalFee                        uint64
	dustThreshold                        uint64
	timeoutResetDelay                    uint32
	timeout                              uint32
	timeoutSlashingAmount                *big.Int
	timeoutNotifierRewardMultiplier      uint32
	commitmentGasOffset                  uint16
	sweepTxMaxTotalFee                   uint64
	sweepTimeout                         uint32
	sweepTimeoutSlashingAmount           *big.Int
	sweepTimeoutNotifierRewardMultiplier uint32
}

type movingFundsCommitmentSubmission struct {
	WalletPublicKeyHash [20]byte
	WalletMainUtxo      *bitcoin.UnspentTransactionOutput
	WalletMembersIDs    []uint32
	WalletMemberIndex   uint32
	TargetWallets       [][20]byte
}

type LocalChain struct {
	mutex sync.Mutex

	depositRequests                          map[[32]byte]*tbtc.DepositChainRequest
	pastDepositRevealedEvents                map[[32]byte][]*tbtc.DepositRevealedEvent
	pastNewWalletRegisteredEvents            map[[32]byte][]*tbtc.NewWalletRegisteredEvent
	depositParameters                        depositParameters
	depositSweepProposalValidations          map[[32]byte]bool
	redemptionParameters                     redemptionParameters
	redemptionRequestMinAge                  uint32
	walletParameters                         walletParameters
	walletChainData                          map[[20]byte]*tbtc.WalletChainData
	blockCounter                             chain.BlockCounter
	pastRedemptionRequestedEvents            map[[32]byte][]*tbtc.RedemptionRequestedEvent
	averageBlockTime                         time.Duration
	pendingRedemptionRequests                map[[32]byte]*tbtc.RedemptionRequest
	redemptionProposalValidations            map[[32]byte]bool
	heartbeatProposalValidations             map[[16]byte]bool
	movingFundsParameters                    movingFundsParameters
	pastMovingFundsCommitmentSubmittedEvents map[[32]byte][]*tbtc.MovingFundsCommitmentSubmittedEvent
	movingFundsProposalValidations           map[[32]byte]bool
	movingFundsCommitmentSubmissions         []*movingFundsCommitmentSubmission
	operatorIDs                              map[chain.Address]uint32
	redemptionDelays						 map[[32]byte]time.Duration
}

func NewLocalChain() *LocalChain {
	return &LocalChain{
		depositRequests:                          make(map[[32]byte]*tbtc.DepositChainRequest),
		pastDepositRevealedEvents:                make(map[[32]byte][]*tbtc.DepositRevealedEvent),
		pastNewWalletRegisteredEvents:            make(map[[32]byte][]*tbtc.NewWalletRegisteredEvent),
		depositSweepProposalValidations:          make(map[[32]byte]bool),
		pastRedemptionRequestedEvents:            make(map[[32]byte][]*tbtc.RedemptionRequestedEvent),
		walletChainData:                          make(map[[20]byte]*tbtc.WalletChainData),
		pendingRedemptionRequests:                make(map[[32]byte]*tbtc.RedemptionRequest),
		redemptionProposalValidations:            make(map[[32]byte]bool),
		heartbeatProposalValidations:             make(map[[16]byte]bool),
		pastMovingFundsCommitmentSubmittedEvents: make(map[[32]byte][]*tbtc.MovingFundsCommitmentSubmittedEvent),
		movingFundsProposalValidations:           make(map[[32]byte]bool),
		movingFundsCommitmentSubmissions:         make([]*movingFundsCommitmentSubmission, 0),
		operatorIDs:                              make(map[chain.Address]uint32),
		redemptionDelays:						 make(map[[32]byte]time.Duration),
	}
}

func (lc *LocalChain) PastDepositRevealedEvents(
	filter *tbtc.DepositRevealedEventFilter,
) ([]*tbtc.DepositRevealedEvent, error) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	eventsKey, err := buildPastDepositRevealedEventsKey(filter)
	if err != nil {
		return nil, err
	}

	events, ok := lc.pastDepositRevealedEvents[eventsKey]
	if !ok {
		return nil, fmt.Errorf("no events for given filter")
	}

	return events, nil
}

func (lc *LocalChain) AddPastDepositRevealedEvent(
	filter *tbtc.DepositRevealedEventFilter,
	event *tbtc.DepositRevealedEvent,
) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	eventsKey, err := buildPastDepositRevealedEventsKey(filter)
	if err != nil {
		return err
	}

	lc.pastDepositRevealedEvents[eventsKey] = append(
		lc.pastDepositRevealedEvents[eventsKey],
		event,
	)

	return nil
}

func buildPastDepositRevealedEventsKey(
	filter *tbtc.DepositRevealedEventFilter,
) ([32]byte, error) {
	if filter == nil {
		return [32]byte{}, nil
	}

	var buffer bytes.Buffer

	startBlock := make([]byte, 8)
	binary.BigEndian.PutUint64(startBlock, filter.StartBlock)
	buffer.Write(startBlock)

	if filter.EndBlock != nil {
		endBlock := make([]byte, 8)
		binary.BigEndian.PutUint64(startBlock, *filter.EndBlock)
		buffer.Write(endBlock)
	}

	for _, depositor := range filter.Depositor {
		depositorBytes, err := hex.DecodeString(depositor.String())
		if err != nil {
			return [32]byte{}, err
		}

		buffer.Write(depositorBytes)
	}

	for _, walletPublicKeyHash := range filter.WalletPublicKeyHash {
		buffer.Write(walletPublicKeyHash[:])
	}

	return sha256.Sum256(buffer.Bytes()), nil
}

func (lc *LocalChain) GetDepositRequest(
	fundingTxHash bitcoin.Hash,
	fundingOutputIndex uint32,
) (*tbtc.DepositChainRequest, bool, error) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	requestKey := buildDepositRequestKey(fundingTxHash, fundingOutputIndex)

	request, ok := lc.depositRequests[requestKey]
	if !ok {
		return nil, false, nil
	}

	return request, true, nil
}

func (lc *LocalChain) SetDepositRequest(
	fundingTxHash bitcoin.Hash,
	fundingOutputIndex uint32,
	request *tbtc.DepositChainRequest,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	requestKey := buildDepositRequestKey(fundingTxHash, fundingOutputIndex)

	lc.depositRequests[requestKey] = request
}

func (lc *LocalChain) PastNewWalletRegisteredEvents(
	filter *tbtc.NewWalletRegisteredEventFilter,
) ([]*tbtc.NewWalletRegisteredEvent, error) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	eventsKey, err := buildPastNewWalletRegisteredEventsKey(filter)
	if err != nil {
		return nil, err
	}

	events, ok := lc.pastNewWalletRegisteredEvents[eventsKey]
	if !ok {
		return nil, fmt.Errorf("no events for given filter")
	}

	return events, nil
}

func (lc *LocalChain) AddPastNewWalletRegisteredEvent(
	filter *tbtc.NewWalletRegisteredEventFilter,
	event *tbtc.NewWalletRegisteredEvent,
) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	eventsKey, err := buildPastNewWalletRegisteredEventsKey(filter)
	if err != nil {
		return err
	}

	if _, ok := lc.pastNewWalletRegisteredEvents[eventsKey]; !ok {
		lc.pastNewWalletRegisteredEvents[eventsKey] = []*tbtc.NewWalletRegisteredEvent{}
	}

	lc.pastNewWalletRegisteredEvents[eventsKey] = append(
		lc.pastNewWalletRegisteredEvents[eventsKey],
		event,
	)

	return nil
}

func buildPastNewWalletRegisteredEventsKey(
	filter *tbtc.NewWalletRegisteredEventFilter,
) ([32]byte, error) {
	if filter == nil {
		return [32]byte{}, nil
	}

	var buffer bytes.Buffer

	startBlock := make([]byte, 8)
	binary.BigEndian.PutUint64(startBlock, filter.StartBlock)
	buffer.Write(startBlock)

	if filter.EndBlock != nil {
		endBlock := make([]byte, 8)
		binary.BigEndian.PutUint64(startBlock, *filter.EndBlock)
		buffer.Write(endBlock)
	}

	for _, ecdsaWalletID := range filter.EcdsaWalletID {
		buffer.Write(ecdsaWalletID[:])
	}

	for _, walletPublicKeyHash := range filter.WalletPublicKeyHash {
		buffer.Write(walletPublicKeyHash[:])
	}

	return sha256.Sum256(buffer.Bytes()), nil
}

func (lc *LocalChain) PastRedemptionRequestedEvents(
	filter *tbtc.RedemptionRequestedEventFilter,
) ([]*tbtc.RedemptionRequestedEvent, error) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	eventsKey, err := buildPastRedemptionRequestedEventsKey(filter)
	if err != nil {
		return nil, err
	}

	events, ok := lc.pastRedemptionRequestedEvents[eventsKey]
	if !ok {
		return nil, fmt.Errorf("no events for given filter")
	}

	return events, nil
}

func (lc *LocalChain) AddPastRedemptionRequestedEvent(
	filter *tbtc.RedemptionRequestedEventFilter,
	event *tbtc.RedemptionRequestedEvent,
) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	eventsKey, err := buildPastRedemptionRequestedEventsKey(filter)
	if err != nil {
		return err
	}

	lc.pastRedemptionRequestedEvents[eventsKey] = append(
		lc.pastRedemptionRequestedEvents[eventsKey],
		event,
	)

	return nil
}

func buildPastRedemptionRequestedEventsKey(
	filter *tbtc.RedemptionRequestedEventFilter,
) ([32]byte, error) {
	if filter == nil {
		return [32]byte{}, nil
	}

	var buffer bytes.Buffer

	startBlock := make([]byte, 8)
	binary.BigEndian.PutUint64(startBlock, filter.StartBlock)
	buffer.Write(startBlock)

	if filter.EndBlock != nil {
		endBlock := make([]byte, 8)
		binary.BigEndian.PutUint64(startBlock, *filter.EndBlock)
		buffer.Write(endBlock)
	}

	for _, walletPublicKeyHash := range filter.WalletPublicKeyHash {
		buffer.Write(walletPublicKeyHash[:])
	}

	for _, redeemer := range filter.Redeemer {
		redeemerHex, err := hex.DecodeString(redeemer.String())
		if err != nil {
			return [32]byte{}, err
		}

		buffer.Write(redeemerHex)
	}

	return sha256.Sum256(buffer.Bytes()), nil
}

func buildPastMovingFundsCommitmentSubmittedEventsKey(
	filter *tbtc.MovingFundsCommitmentSubmittedEventFilter,
) ([32]byte, error) {
	if filter == nil {
		return [32]byte{}, nil
	}

	var buffer bytes.Buffer

	startBlock := make([]byte, 8)
	binary.BigEndian.PutUint64(startBlock, filter.StartBlock)
	buffer.Write(startBlock)

	if filter.EndBlock != nil {
		endBlock := make([]byte, 8)
		binary.BigEndian.PutUint64(startBlock, *filter.EndBlock)
		buffer.Write(endBlock)
	}

	for _, walletPublicKeyHash := range filter.WalletPublicKeyHash {
		buffer.Write(walletPublicKeyHash[:])
	}

	return sha256.Sum256(buffer.Bytes()), nil
}

func (lc *LocalChain) BuildDepositKey(fundingTxHash bitcoin.Hash, fundingOutputIndex uint32) *big.Int {
	depositKeyBytes := buildDepositRequestKey(fundingTxHash, fundingOutputIndex)

	return new(big.Int).SetBytes(depositKeyBytes[:])
}

func buildDepositRequestKey(
	fundingTxHash bitcoin.Hash,
	fundingOutputIndex uint32,
) [32]byte {
	fundingOutputIndexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(fundingOutputIndexBytes, fundingOutputIndex)

	return sha256.Sum256(append(fundingTxHash[:], fundingOutputIndexBytes...))
}

func (lc *LocalChain) BuildRedemptionKey(
	walletPublicKeyHash [20]byte,
	redeemerOutputScript bitcoin.Script,
) (*big.Int, error) {
	redemptionKeyBytes := buildRedemptionRequestKey(
		walletPublicKeyHash,
		redeemerOutputScript,
	)

	return new(big.Int).SetBytes(redemptionKeyBytes[:]), nil
}

func buildRedemptionRequestKey(
	walletPublicKeyHash [20]byte,
	redeemerOutputScript bitcoin.Script,
) [32]byte {
	return sha256.Sum256(append(walletPublicKeyHash[:], redeemerOutputScript...))
}

func (lc *LocalChain) GetDepositParameters() (
	dustThreshold uint64,
	treasuryFeeDivisor uint64,
	txMaxFee uint64,
	revealAheadPeriod uint32,
	err error,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	return lc.depositParameters.dustThreshold,
		lc.depositParameters.treasuryFeeDivisor,
		lc.depositParameters.txMaxFee,
		lc.depositParameters.revealAheadPeriod,
		nil
}

func (lc *LocalChain) GetPendingRedemptionRequest(
	walletPublicKeyHash [20]byte,
	redeemerOutputScript bitcoin.Script,
) (*tbtc.RedemptionRequest, bool, error) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	requestKey := buildRedemptionRequestKey(walletPublicKeyHash, redeemerOutputScript)

	request, ok := lc.pendingRedemptionRequests[requestKey]
	if !ok {
		return nil, false, nil
	}

	return request, true, nil
}

func (lc *LocalChain) SetPendingRedemptionRequest(
	walletPublicKeyHash [20]byte,
	request *tbtc.RedemptionRequest,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	requestKey := buildRedemptionRequestKey(
		walletPublicKeyHash,
		request.RedeemerOutputScript,
	)

	lc.pendingRedemptionRequests[requestKey] = request
}

func (lc *LocalChain) SetDepositParameters(
	dustThreshold uint64,
	treasuryFeeDivisor uint64,
	txMaxFee uint64,
	revealAheadPeriod uint32,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	lc.depositParameters = depositParameters{
		dustThreshold:      dustThreshold,
		treasuryFeeDivisor: treasuryFeeDivisor,
		txMaxFee:           txMaxFee,
		revealAheadPeriod:  revealAheadPeriod,
	}
}

func (lc *LocalChain) GetRedemptionParameters() (
	dustThreshold uint64,
	treasuryFeeDivisor uint64,
	txMaxFee uint64,
	txMaxTotalFee uint64,
	timeout uint32,
	timeoutSlashingAmount *big.Int,
	timeoutNotifierRewardMultiplier uint32,
	err error,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	return lc.redemptionParameters.dustThreshold,
		lc.redemptionParameters.treasuryFeeDivisor,
		lc.redemptionParameters.txMaxFee,
		lc.redemptionParameters.txMaxTotalFee,
		lc.redemptionParameters.timeout,
		lc.redemptionParameters.timeoutSlashingAmount,
		lc.redemptionParameters.timeoutNotifierRewardMultiplier,
		nil
}

func (lc *LocalChain) SetRedemptionParameters(
	dustThreshold uint64,
	treasuryFeeDivisor uint64,
	txMaxFee uint64,
	txMaxTotalFee uint64,
	timeout uint32,
	timeoutSlashingAmount *big.Int,
	timeoutNotifierRewardMultiplier uint32,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	lc.redemptionParameters = redemptionParameters{
		dustThreshold:                   dustThreshold,
		treasuryFeeDivisor:              treasuryFeeDivisor,
		txMaxFee:                        txMaxFee,
		txMaxTotalFee:                   txMaxTotalFee,
		timeout:                         timeout,
		timeoutSlashingAmount:           timeoutSlashingAmount,
		timeoutNotifierRewardMultiplier: timeoutNotifierRewardMultiplier,
	}
}

func (lc *LocalChain) ValidateDepositSweepProposal(
	walletPublicKeyHash [20]byte,
	proposal *tbtc.DepositSweepProposal,
	depositsExtraInfo []struct {
		*tbtc.Deposit
		FundingTx *bitcoin.Transaction
	},
) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	key, err := buildDepositSweepProposalValidationKey(
		walletPublicKeyHash,
		proposal,
	)
	if err != nil {
		return err
	}

	result, ok := lc.depositSweepProposalValidations[key]
	if !ok {
		return fmt.Errorf("validation result unknown")
	}

	if !result {
		return fmt.Errorf("validation failed")
	}

	return nil
}

func (lc *LocalChain) SetDepositSweepProposalValidationResult(
	walletPublicKeyHash [20]byte,
	proposal *tbtc.DepositSweepProposal,
	depositsExtraInfo []struct {
		*tbtc.Deposit
		FundingTx *bitcoin.Transaction
	},
	result bool,
) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	key, err := buildDepositSweepProposalValidationKey(
		walletPublicKeyHash,
		proposal,
	)
	if err != nil {
		return err
	}

	lc.depositSweepProposalValidations[key] = result

	return nil
}

func buildDepositSweepProposalValidationKey(
	walletPublicKeyHash [20]byte,
	proposal *tbtc.DepositSweepProposal,
) ([32]byte, error) {
	var buffer bytes.Buffer

	buffer.Write(walletPublicKeyHash[:])

	for _, deposit := range proposal.DepositsKeys {
		buffer.Write(deposit.FundingTxHash[:])

		fundingOutputIndex := make([]byte, 4)
		binary.BigEndian.PutUint32(fundingOutputIndex, deposit.FundingOutputIndex)
		buffer.Write(fundingOutputIndex)
	}

	buffer.Write(proposal.SweepTxFee.Bytes())

	return sha256.Sum256(buffer.Bytes()), nil
}

func (lc *LocalChain) ValidateRedemptionProposal(
	walletPublicKeyHash [20]byte,
	proposal *tbtc.RedemptionProposal,
) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	key, err := buildRedemptionProposalValidationKey(
		walletPublicKeyHash,
		proposal,
	)
	if err != nil {
		return err
	}

	result, ok := lc.redemptionProposalValidations[key]
	if !ok {
		return fmt.Errorf("validation result unknown")
	}

	if !result {
		return fmt.Errorf("validation failed")
	}

	return nil
}

func (lc *LocalChain) SetRedemptionProposalValidationResult(
	walletPublicKeyHash [20]byte,
	proposal *tbtc.RedemptionProposal,
	result bool,
) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	key, err := buildRedemptionProposalValidationKey(
		walletPublicKeyHash,
		proposal,
	)
	if err != nil {
		return err
	}

	lc.redemptionProposalValidations[key] = result

	return nil
}

func (lc *LocalChain) ValidateHeartbeatProposal(
	walletPublicKeyHash [20]byte,
	proposal *tbtc.HeartbeatProposal,
) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	result, ok := lc.heartbeatProposalValidations[proposal.Message]
	if !ok {
		return fmt.Errorf("validation result unknown")
	}

	if !result {
		return fmt.Errorf("validation failed")
	}

	return nil
}

func (lc *LocalChain) SetHeartbeatProposalValidationResult(
	proposal *tbtc.HeartbeatProposal,
	result bool,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	lc.heartbeatProposalValidations[proposal.Message] = result
}

func (lc *LocalChain) GetMovingFundsParameters() (
	txMaxTotalFee uint64,
	dustThreshold uint64,
	timeoutResetDelay uint32,
	timeout uint32,
	timeoutSlashingAmount *big.Int,
	timeoutNotifierRewardMultiplier uint32,
	commitmentGasOffset uint16,
	sweepTxMaxTotalFee uint64,
	sweepTimeout uint32,
	sweepTimeoutSlashingAmount *big.Int,
	sweepTimeoutNotifierRewardMultiplier uint32,
	err error,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	return lc.movingFundsParameters.txMaxTotalFee,
		lc.movingFundsParameters.dustThreshold,
		lc.movingFundsParameters.timeoutResetDelay,
		lc.movingFundsParameters.timeout,
		lc.movingFundsParameters.timeoutSlashingAmount,
		lc.movingFundsParameters.timeoutNotifierRewardMultiplier,
		lc.movingFundsParameters.commitmentGasOffset,
		lc.movingFundsParameters.sweepTxMaxTotalFee,
		lc.movingFundsParameters.sweepTimeout,
		lc.movingFundsParameters.sweepTimeoutSlashingAmount,
		lc.movingFundsParameters.sweepTimeoutNotifierRewardMultiplier,
		nil
}

func (lc *LocalChain) SetMovingFundsParameters(
	txMaxTotalFee uint64,
	dustThreshold uint64,
	timeoutResetDelay uint32,
	timeout uint32,
	timeoutSlashingAmount *big.Int,
	timeoutNotifierRewardMultiplier uint32,
	commitmentGasOffset uint16,
	sweepTxMaxTotalFee uint64,
	sweepTimeout uint32,
	sweepTimeoutSlashingAmount *big.Int,
	sweepTimeoutNotifierRewardMultiplier uint32,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	lc.movingFundsParameters = movingFundsParameters{
		txMaxTotalFee:                        txMaxTotalFee,
		dustThreshold:                        dustThreshold,
		timeoutResetDelay:                    timeoutResetDelay,
		timeout:                              timeout,
		timeoutSlashingAmount:                timeoutSlashingAmount,
		timeoutNotifierRewardMultiplier:      timeoutNotifierRewardMultiplier,
		commitmentGasOffset:                  commitmentGasOffset,
		sweepTxMaxTotalFee:                   sweepTxMaxTotalFee,
		sweepTimeout:                         sweepTimeout,
		sweepTimeoutSlashingAmount:           sweepTimeoutSlashingAmount,
		sweepTimeoutNotifierRewardMultiplier: sweepTimeoutNotifierRewardMultiplier,
	}
}

func buildMovingFundsProposalValidationKey(
	walletPublicKeyHash [20]byte,
	mainUTXO *bitcoin.UnspentTransactionOutput,
	proposal *tbtc.MovingFundsProposal,
) ([32]byte, error) {
	var buffer bytes.Buffer

	buffer.Write(walletPublicKeyHash[:])

	buffer.Write(mainUTXO.Outpoint.TransactionHash[:])
	binary.Write(&buffer, binary.BigEndian, mainUTXO.Outpoint.OutputIndex)
	binary.Write(&buffer, binary.BigEndian, mainUTXO.Value)

	for _, wallet := range proposal.TargetWallets {
		buffer.Write(wallet[:])
	}

	buffer.Write(proposal.MovingFundsTxFee.Bytes())

	return sha256.Sum256(buffer.Bytes()), nil
}

func (lc *LocalChain) ValidateMovingFundsProposal(
	walletPublicKeyHash [20]byte,
	mainUTXO *bitcoin.UnspentTransactionOutput,
	proposal *tbtc.MovingFundsProposal,
) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	key, err := buildMovingFundsProposalValidationKey(
		walletPublicKeyHash,
		mainUTXO,
		proposal,
	)
	if err != nil {
		return err
	}

	result, ok := lc.movingFundsProposalValidations[key]
	if !ok {
		return fmt.Errorf("validation result unknown")
	}

	if !result {
		return fmt.Errorf("validation failed")
	}

	return nil
}

func (lc *LocalChain) SetMovingFundsProposalValidationResult(
	walletPublicKeyHash [20]byte,
	mainUTXO *bitcoin.UnspentTransactionOutput,
	proposal *tbtc.MovingFundsProposal,
	result bool,
) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	key, err := buildMovingFundsProposalValidationKey(
		walletPublicKeyHash,
		mainUTXO,
		proposal,
	)
	if err != nil {
		return err
	}

	lc.movingFundsProposalValidations[key] = result

	return nil
}

func buildRedemptionProposalValidationKey(
	walletPublicKeyHash [20]byte,
	proposal *tbtc.RedemptionProposal,
) ([32]byte, error) {
	var buffer bytes.Buffer

	buffer.Write(walletPublicKeyHash[:])

	for _, script := range proposal.RedeemersOutputScripts {
		buffer.Write(script)
	}

	buffer.Write(proposal.RedemptionTxFee.Bytes())

	return sha256.Sum256(buffer.Bytes()), nil
}

func (lc *LocalChain) GetRedemptionMaxSize() (uint16, error) {
	panic("unsupported")
}

func (lc *LocalChain) GetRedemptionRequestMinAge() (uint32, error) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	return lc.redemptionRequestMinAge, nil
}

func (lc *LocalChain) SetRedemptionRequestMinAge(redemptionRequestMinAge uint32) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	lc.redemptionRequestMinAge = redemptionRequestMinAge
}

func (lc *LocalChain) GetDepositSweepMaxSize() (uint16, error) {
	panic("unsupported")
}

func (lc *LocalChain) BlockCounter() (chain.BlockCounter, error) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	return lc.blockCounter, nil
}

func (lc *LocalChain) SetBlockCounter(blockCounter chain.BlockCounter) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	lc.blockCounter = blockCounter
}

func (lc *LocalChain) AverageBlockTime() time.Duration {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	return lc.averageBlockTime
}

func (lc *LocalChain) SetOperatorID(
	operatorAddress chain.Address,
	operatorID chain.OperatorID,
) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	_, ok := lc.operatorIDs[operatorAddress]
	if !ok {
		lc.operatorIDs[operatorAddress] = operatorID
	}

	return nil
}

func (lc *LocalChain) GetOperatorID(
	operatorAddress chain.Address,
) (chain.OperatorID, error) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	operatorID, ok := lc.operatorIDs[operatorAddress]
	if !ok {
		return 0, fmt.Errorf("operator not found")
	}

	return operatorID, nil
}

func (lc *LocalChain) SetAverageBlockTime(averageBlockTime time.Duration) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	lc.averageBlockTime = averageBlockTime
}

func (lc *LocalChain) GetWallet(walletPublicKeyHash [20]byte) (
	*tbtc.WalletChainData,
	error,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	data, ok := lc.walletChainData[walletPublicKeyHash]
	if !ok {
		fmt.Println("Not found")
		return nil, fmt.Errorf("wallet chain data not found")
	}

	return data, nil
}

func (lc *LocalChain) SetWallet(
	walletPublicKeyHash [20]byte,
	data *tbtc.WalletChainData,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	lc.walletChainData[walletPublicKeyHash] = data
}

func (lc *LocalChain) GetWalletParameters() (
	creationPeriod uint32,
	creationMinBtcBalance uint64,
	creationMaxBtcBalance uint64,
	closureMinBtcBalance uint64,
	maxAge uint32,
	maxBtcTransfer uint64,
	closingPeriod uint32,
	err error,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	return lc.walletParameters.creationPeriod,
		lc.walletParameters.creationMinBtcBalance,
		lc.walletParameters.creationMaxBtcBalance,
		lc.walletParameters.closureMinBtcBalance,
		lc.walletParameters.maxAge,
		lc.walletParameters.maxBtcTransfer,
		lc.walletParameters.closingPeriod,
		nil
}

func (lc *LocalChain) SetWalletParameters(
	creationPeriod uint32,
	creationMinBtcBalance uint64,
	creationMaxBtcBalance uint64,
	closureMinBtcBalance uint64,
	maxAge uint32,
	maxBtcTransfer uint64,
	closingPeriod uint32,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	lc.walletParameters = walletParameters{
		creationPeriod:        creationPeriod,
		creationMinBtcBalance: creationMinBtcBalance,
		creationMaxBtcBalance: creationMaxBtcBalance,
		closureMinBtcBalance:  closureMinBtcBalance,
		maxAge:                maxAge,
		maxBtcTransfer:        maxBtcTransfer,
		closingPeriod:         closingPeriod,
	}
}

func (lc *LocalChain) GetLiveWalletsCount() (uint32, error) {
	panic("unsupported")
}

func (lc *LocalChain) ComputeMainUtxoHash(mainUtxo *bitcoin.UnspentTransactionOutput) [32]byte {
	panic("unsupported")
}

func (lc *LocalChain) ComputeMovingFundsCommitmentHash(targetWallets [][20]byte) [32]byte {
	packedWallets := []byte{}

	for _, wallet := range targetWallets {
		packedWallets = append(packedWallets, wallet[:]...)
		// Each wallet hash must be padded with 12 zero bytes following the
		// actual hash.
		packedWallets = append(packedWallets, make([]byte, 12)...)
	}

	return crypto.Keccak256Hash(packedWallets)
}

func (lc *LocalChain) AddPastMovingFundsCommitmentSubmittedEvent(
	filter *tbtc.MovingFundsCommitmentSubmittedEventFilter,
	event *tbtc.MovingFundsCommitmentSubmittedEvent,
) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	eventsKey, err := buildPastMovingFundsCommitmentSubmittedEventsKey(filter)
	if err != nil {
		return err
	}

	if _, ok := lc.pastMovingFundsCommitmentSubmittedEvents[eventsKey]; !ok {
		lc.pastMovingFundsCommitmentSubmittedEvents[eventsKey] = []*tbtc.MovingFundsCommitmentSubmittedEvent{}
	}

	lc.pastMovingFundsCommitmentSubmittedEvents[eventsKey] = append(
		lc.pastMovingFundsCommitmentSubmittedEvents[eventsKey],
		event,
	)

	return nil
}

func (lc *LocalChain) PastMovingFundsCommitmentSubmittedEvents(
	filter *tbtc.MovingFundsCommitmentSubmittedEventFilter,
) ([]*tbtc.MovingFundsCommitmentSubmittedEvent, error) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	eventsKey, err := buildPastMovingFundsCommitmentSubmittedEventsKey(filter)
	if err != nil {
		return nil, err
	}

	events, ok := lc.pastMovingFundsCommitmentSubmittedEvents[eventsKey]
	if !ok {
		return nil, fmt.Errorf("no events for given filter")
	}

	return events, nil
}

func (lc *LocalChain) SubmitMovingFundsCommitment(
	walletPublicKeyHash [20]byte,
	walletMainUtxo bitcoin.UnspentTransactionOutput,
	walletMembersIDs []uint32,
	walletMemberIndex uint32,
	targetWallets [][20]byte,
) error {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	lc.movingFundsCommitmentSubmissions = append(
		lc.movingFundsCommitmentSubmissions,
		&movingFundsCommitmentSubmission{
			WalletPublicKeyHash: walletPublicKeyHash,
			WalletMainUtxo:      &walletMainUtxo,
			WalletMembersIDs:    walletMembersIDs,
			WalletMemberIndex:   walletMemberIndex,
			TargetWallets:       targetWallets,
		},
	)

	return nil
}

func (lc *LocalChain) GetMovingFundsSubmissions() []*movingFundsCommitmentSubmission {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	return lc.movingFundsCommitmentSubmissions
}

func (lc *LocalChain) GetRedemptionDelay(
	walletPublicKeyHash [20]byte,
	redeemerOutputScript bitcoin.Script,
) (time.Duration, error) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	key := buildRedemptionRequestKey(walletPublicKeyHash, redeemerOutputScript)

	delay, ok := lc.redemptionDelays[key]
	if !ok {
		return 0, fmt.Errorf("redemption delay not found")
	}

	return delay, nil
}

func (lc *LocalChain) SetRedemptionDelay(
	walletPublicKeyHash [20]byte,
	redeemerOutputScript bitcoin.Script,
	delay time.Duration,
) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()

	key := buildRedemptionRequestKey(walletPublicKeyHash, redeemerOutputScript)

	lc.redemptionDelays[key] = delay
}

type MockBlockCounter struct {
	mutex        sync.Mutex
	currentBlock uint64
}

func NewMockBlockCounter() *MockBlockCounter {
	return &MockBlockCounter{}
}

func (mbc *MockBlockCounter) WaitForBlockHeight(blockNumber uint64) error {
	return nil
}

func (mbc *MockBlockCounter) BlockHeightWaiter(blockNumber uint64) (
	<-chan uint64,
	error,
) {
	panic("unsupported")
}

func (mbc *MockBlockCounter) CurrentBlock() (uint64, error) {
	mbc.mutex.Lock()
	defer mbc.mutex.Unlock()

	return mbc.currentBlock, nil
}

func (mbc *MockBlockCounter) SetCurrentBlock(block uint64) {
	mbc.mutex.Lock()
	defer mbc.mutex.Unlock()

	mbc.currentBlock = block
}

func (mbc *MockBlockCounter) WatchBlocks(ctx context.Context) <-chan uint64 {
	panic("unsupported")
}
