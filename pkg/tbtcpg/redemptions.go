package tbtcpg

import (
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/ipfs/go-log/v2"
	"go.uber.org/zap"

	"github.com/keep-network/keep-core/internal/hexutils"
	"github.com/keep-network/keep-core/pkg/bitcoin"
	"github.com/keep-network/keep-core/pkg/tbtc"
)

// RedemptionTask is a task that may produce a redemption proposal.
type RedemptionTask struct {
	chain    Chain
	btcChain bitcoin.Chain
}

func NewRedemptionTask(
	chain Chain,
	btcChain bitcoin.Chain,
) *RedemptionTask {
	return &RedemptionTask{
		chain:    chain,
		btcChain: btcChain,
	}
}

func (rt *RedemptionTask) Run(request *tbtc.CoordinationProposalRequest) (
	tbtc.CoordinationProposal,
	bool,
	error,
) {
	walletPublicKeyHash := request.WalletPublicKeyHash

	taskLogger := logger.With(
		zap.String("task", rt.ActionType().String()),
		zap.String("walletPKH", fmt.Sprintf("0x%x", walletPublicKeyHash)),
	)

	redemptionMaxSize, err := rt.chain.GetRedemptionMaxSize()
	if err != nil {
		return nil, false, fmt.Errorf(
			"failed to get redemption max size: [%w]",
			err,
		)
	}

	redeemersOutputScripts, err := rt.FindPendingRedemptions(
		taskLogger,
		walletPublicKeyHash,
		redemptionMaxSize,
	)
	if err != nil {
		return nil, false, fmt.Errorf(
			"cannot find pending redemption requests: [%w]",
			err,
		)
	}

	if len(redeemersOutputScripts) == 0 {
		taskLogger.Info("no pending redemption requests")
		return nil, false, nil
	}

	proposal, err := rt.ProposeRedemption(
		taskLogger,
		walletPublicKeyHash,
		redeemersOutputScripts,
		0,
	)
	if err != nil {
		return nil, false, fmt.Errorf(
			"cannot prepare redemption proposal: [%w]",
			err,
		)
	}

	return proposal, true, nil
}

func (rt *RedemptionTask) ActionType() tbtc.WalletActionType {
	return tbtc.ActionRedemption
}

// RedemptionRequest represents a redemption request.
type RedemptionRequest struct {
	WalletPublicKeyHash  [20]byte
	RedemptionKey        string
	RedeemerOutputScript bitcoin.Script
	RequestedAt          time.Time
	RequestedAmount      uint64
}

// FindPendingRedemptions finds pending redemptions requests for the
// provided wallet. The returned value is a list of redeemers output
// scripts that come from detected pending requests targeting this wallet.
// The maxNumberOfRequests parameter is used as a ceiling for the number of
// requests in the result. If number of discovered requests meets the
// maxNumberOfRequests the function will stop fetching more requests.
func (rt *RedemptionTask) FindPendingRedemptions(
	taskLogger log.StandardLogger,
	walletPublicKeyHash [20]byte,
	maxNumberOfRequests uint16,
) ([]bitcoin.Script, error) {
	if walletPublicKeyHash == [20]byte{} {
		return nil, fmt.Errorf("wallet public key hash is required")
	}

	taskLogger.Infof(
		"fetching max [%d] redemption requests",
		maxNumberOfRequests,
	)

	blockCounter, err := rt.chain.BlockCounter()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get block counter: [%w]",
			err,
		)
	}

	currentBlockNumber, err := blockCounter.CurrentBlock()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get current block number: [%w]",
			err,
		)
	}

	requestMinAge, err := rt.chain.GetRedemptionRequestMinAge()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get redemption request minimum age: [%w]",
			err,
		)
	}

	_, _, _, _, requestTimeout, _, _, err := rt.chain.GetRedemptionParameters()
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get redemption parameters: [%w]",
			err,
		)
	}

	taskLogger.Infof("fetching pending redemption requests")

	pendingRedemptions, err := findPendingRedemptions(
		taskLogger,
		rt.chain,
		walletPublicKeyHash,
		currentBlockNumber,
		maxNumberOfRequests,
		requestTimeout,
		requestMinAge,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get pending redemptions: [%w]", err)
	}

	taskLogger.Infof("found [%d] redemption requests", len(pendingRedemptions))

	result := make([]bitcoin.Script, 0)

	for _, pendingRedemption := range pendingRedemptions {
		taskLogger.Infof(
			"redemption request [%s] - requested at: [%s]",
			pendingRedemption.RedemptionKey,
			pendingRedemption.RequestedAt,
		)

		result = append(result, pendingRedemption.RedeemerOutputScript)
	}

	return result, nil
}

// ProposeRedemption returns a redemption proposal.
func (rt *RedemptionTask) ProposeRedemption(
	taskLogger log.StandardLogger,
	walletPublicKeyHash [20]byte,
	redeemersOutputScripts []bitcoin.Script,
	fee int64,
) (*tbtc.RedemptionProposal, error) {
	if len(redeemersOutputScripts) == 0 {
		return nil, fmt.Errorf("redemptions list is empty")
	}

	taskLogger.Infof("preparing a redemption proposal")

	// Estimate fee if it's missing. Do not check the estimated fee against
	// the maximum total and per-request fees allowed by the Bridge. This
	// is done during the on-chain validation of the proposal so there is no
	// need to do it here.
	if fee <= 0 {
		taskLogger.Infof("estimating redemption transaction fee")

		estimatedFee, err := EstimateRedemptionFee(
			rt.btcChain,
			redeemersOutputScripts,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"cannot estimate redemption transaction fee: [%w]",
				err,
			)
		}

		fee = estimatedFee
	}

	taskLogger.Infof("redemption transaction fee: [%d]", fee)

	proposal := &tbtc.RedemptionProposal{
		RedeemersOutputScripts: redeemersOutputScripts,
		RedemptionTxFee:        big.NewInt(fee),
	}

	taskLogger.Infof("validating the redemption proposal")

	if _, err := tbtc.ValidateRedemptionProposal(
		taskLogger,
		walletPublicKeyHash,
		proposal,
		rt.chain,
	); err != nil {
		return nil, fmt.Errorf("failed to verify redemption proposal: %v", err)
	}

	return proposal, nil
}

func findPendingRedemptions(
	fnLogger log.StandardLogger,
	chain Chain,
	walletPublicKeyHash [20]byte,
	currentBlockNumber uint64,
	requestsLimit uint16,
	requestTimeout uint32,
	requestMinAge uint32,
) ([]*RedemptionRequest, error) {
	// We are interested with `RedemptionRequested` events that are not
	// timed out yet. That means there is no sense to look for events that
	// occurred earlier than `now - requestTimeout`. However,
	// the event filter expects a block range while `requestTimeout`
	// is in seconds. To overcome that problem, we estimate the redemption
	// request timeout in blocks, using the average block time of the host chain.
	// Note that this estimation is not 100% accurate as the actual block time
	// may differ from the assumed one.
	requestTimeoutBlocks :=
		uint64(requestTimeout) / uint64(chain.AverageBlockTime().Seconds())
	// Then, we set the start block of the filter using the estimated redemption
	// request timeout in blocks. Note that if the actual average block time is
	// lesser than the assumed one, some events being on the edge of the block
	// range may be omitted. To avoid that, we make the block range a little
	// wider by using a constant factor of 1000 blocks.
	filterStartBlock := uint64(0)
	if filterLookbackBlocks := requestTimeoutBlocks + 1000; currentBlockNumber > filterLookbackBlocks {
		filterStartBlock = currentBlockNumber - filterLookbackBlocks
	}

	filter := &tbtc.RedemptionRequestedEventFilter{
		StartBlock: filterStartBlock,
	}
	if walletPublicKeyHash != [20]byte{} {
		filter.WalletPublicKeyHash = [][20]byte{walletPublicKeyHash}
	}

	events, err := chain.PastRedemptionRequestedEvents(filter)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get past redemption requested events: [%w]",
			err,
		)
	}

	// Take the oldest first.
	sort.SliceStable(
		events, func(i, j int) bool {
			return events[i].BlockNumber < events[j].BlockNumber
		},
	)

	// There may be multiple events targeting the same redemption key
	// (i.e. the same wallet and output script pair). The Bridge contract
	// allows only for one pending request with the given redemption key
	// at the same time. That means we need to deduplicate the events list
	// and take only the latest event for the given redemption key.
	eventsSet := make(map[string]*tbtc.RedemptionRequestedEvent)
	for _, event := range events {
		redemptionKey, err := chain.BuildRedemptionKey(
			event.WalletPublicKeyHash,
			event.RedeemerOutputScript,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to build redemption key: [%v]", err)
		}

		// Events are sorted from the oldest to the newest so there is no
		// need to check for existence. We can just overwrite.
		eventsSet[hexutils.Encode(redemptionKey.Bytes())] = event
	}

	fnLogger.Infof("found [%d] RedemptionRequested events", len(eventsSet))

	fnLogger.Infof("checking pending redemptions details")

	pendingRedemptions := make([]*RedemptionRequest, 0)

	eventIndex := 0
redemptionRequestedLoop:
	for redemptionKey, event := range eventsSet {
		eventIndex++

		fnLogger.Debugf(
			"getting pending redemption details [%s]",
			redemptionKey,
		)

		// Check if there is still a pending redemption for the given redemption
		// requested event.
		pendingRedemption, found, err := chain.GetPendingRedemptionRequest(
			event.WalletPublicKeyHash,
			event.RedeemerOutputScript,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to get pending redemption request: [%w]",
				err,
			)
		}
		if !found {
			fnLogger.Infof(
				"redemption request [%s] is no longer pending",
				redemptionKey,
			)

			continue redemptionRequestedLoop
		}

		pendingRedemptions = append(
			pendingRedemptions, &RedemptionRequest{
				WalletPublicKeyHash:  event.WalletPublicKeyHash,
				RedemptionKey:        redemptionKey,
				RedeemerOutputScript: event.RedeemerOutputScript,
				RequestedAt:          pendingRedemption.RequestedAt,
				RequestedAmount:      pendingRedemption.RequestedAmount,
			},
		)
	}

	resultSliceCapacity := len(pendingRedemptions)
	if requestsLimit > 0 {
		resultSliceCapacity = int(requestsLimit)
	}

	// Sort the pending redemptions from oldest to newest.
	sort.SliceStable(
		pendingRedemptions, func(i, j int) bool {
			return pendingRedemptions[i].RequestedAt.Before(pendingRedemptions[j].RequestedAt)
		},
	)

	// Only redemption requests in range:
	// [now - requestTimeout, now - requestMinAge]
	// should be taken into consideration.
	redemptionRequestsRangeStartTimestamp := time.Now().Add(
		-time.Duration(requestTimeout) * time.Second,
	)
	redemptionRequestsRangeEndTimestamp := time.Now().Add(
		-time.Duration(requestMinAge) * time.Second,
	)

	result := make([]*RedemptionRequest, 0, resultSliceCapacity)
	for _, pendingRedemption := range pendingRedemptions {
		if len(result) == cap(result) {
			break
		}

		// Check if timeout passed for the redemption request.
		if pendingRedemption.RequestedAt.Before(redemptionRequestsRangeStartTimestamp) {
			fnLogger.Infof(
				"redemption request [%s] has already timed out",
				pendingRedemption.RedemptionKey,
			)
			continue
		}

		// Check if enough time elapsed since the redemption request.
		if pendingRedemption.RequestedAt.After(redemptionRequestsRangeEndTimestamp) {
			fnLogger.Infof(
				"redemption request [%s] is not old enough",
				pendingRedemption.RedemptionKey,
			)
			continue
		}

		result = append(result, pendingRedemption)
	}

	return result, nil
}

// EstimateRedemptionFee estimates fee for the redemption transaction that pays
// the provided redeemers output scripts.
func EstimateRedemptionFee(
	btcChain bitcoin.Chain,
	redeemersOutputScripts []bitcoin.Script,
) (int64, error) {
	sizeEstimator := bitcoin.NewTransactionSizeEstimator().
		// 1 P2WPKH main UTXO input.
		AddPublicKeyHashInputs(1, true).
		// 1 P2WPKH change output.
		AddPublicKeyHashOutputs(1, true)

	for _, script := range redeemersOutputScripts {
		switch bitcoin.GetScriptType(script) {
		case bitcoin.P2PKHScript:
			sizeEstimator.AddPublicKeyHashOutputs(1, false)
		case bitcoin.P2WPKHScript:
			sizeEstimator.AddPublicKeyHashOutputs(1, true)
		case bitcoin.P2SHScript:
			sizeEstimator.AddScriptHashOutputs(1, false)
		case bitcoin.P2WSHScript:
			sizeEstimator.AddScriptHashOutputs(1, true)
		default:
			return 0, fmt.Errorf("non-standard redeemer output script type")
		}
	}

	transactionSize, err := sizeEstimator.VirtualSize()
	if err != nil {
		return 0, fmt.Errorf("cannot estimate transaction virtual size: [%v]", err)
	}

	feeEstimator := bitcoin.NewTransactionFeeEstimator(btcChain)

	totalFee, err := feeEstimator.EstimateFee(transactionSize)
	if err != nil {
		return 0, fmt.Errorf("cannot estimate transaction fee: [%v]", err)
	}

	return totalFee, nil
}
