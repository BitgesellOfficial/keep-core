package tbtc

import (
	"context"
	"fmt"
	"github.com/keep-network/keep-core/pkg/generator"
	"github.com/keep-network/keep-core/pkg/net"
	"github.com/keep-network/keep-core/pkg/protocol/announcer"
	"github.com/keep-network/keep-core/pkg/protocol/group"
	"github.com/keep-network/keep-core/pkg/tecdsa"
	"github.com/keep-network/keep-core/pkg/tecdsa/signing"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/semaphore"
	"math/big"
	"reflect"
	"strings"
)

const (
	// signingBatchInterludeBlocks determines the block duration of the
	// interlude preserved between subsequent signings in a signing batch.
	signingBatchInterludeBlocks = 2
)

// signingExecutor is a component responsible for executing signing related to
// a specific wallet whose part is controlled by this node.
//
// TODO: Add busy check.
type signingExecutor struct {
	lock *semaphore.Weighted

	signers             []*signer
	broadcastChannel    net.BroadcastChannel
	membershipValidator *group.MembershipValidator
	chainConfig         *ChainConfig
	protocolLatch       *generator.ProtocolLatch

	// currentBlockFn is a function used to get the current block.
	currentBlockFn func() (uint64, error)
	// waitForBlockFn is a function used to wait for the given block.
	waitForBlockFn waitForBlockFn

	// signingAttemptsLimit determines the maximum attempts count that will
	// be made by a single signer for the given message. Once the attempts
	// limit is hit the signer gives up.
	signingAttemptsLimit uint
}

func newSigningExecutor(
	signers []*signer,
	broadcastChannel net.BroadcastChannel,
	membershipValidator *group.MembershipValidator,
	chainConfig *ChainConfig,
	protocolLatch *generator.ProtocolLatch,
	currentBlockFn func() (uint64, error),
	waitForBlockFn waitForBlockFn,
	signingAttemptsLimit uint,
) *signingExecutor {
	return &signingExecutor{
		lock:                 semaphore.NewWeighted(1),
		signers:              signers,
		broadcastChannel:     broadcastChannel,
		membershipValidator:  membershipValidator,
		chainConfig:          chainConfig,
		protocolLatch:        protocolLatch,
		currentBlockFn:       currentBlockFn,
		waitForBlockFn:       waitForBlockFn,
		signingAttemptsLimit: signingAttemptsLimit,
	}
}

// signBatch performs the signing process for each message from the given
// messages batch, one after another. If at least one message cannot be signed,
// this function returns an error. If all messages were signed successfully,
// a slice of signatures is returned. Order of the returned signatures matches
// the order of the messages in the batch, i.e. the first signature corresponds
// to the first message, and so on.
func (se *signingExecutor) signBatch(
	ctx context.Context,
	messages []*big.Int,
	startBlock uint64,
) ([]*tecdsa.Signature, error) {
	wallet := se.wallet()

	walletPublicKeyBytes, err := marshalPublicKey(wallet.publicKey)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal wallet public key: [%v]", err)
	}

	messagesDigests := make([]string, len(messages))
	for i, message := range messages {
		bytes := message.Bytes()
		messagesDigests[i] = fmt.Sprintf("0x%x...%x", bytes[:2], bytes[len(bytes)-2:])
	}

	signingBatchLogger := logger.With(
		zap.String("wallet", fmt.Sprintf("0x%x", walletPublicKeyBytes)),
		zap.String("messages", strings.Join(messagesDigests, ", ")),
	)

	signingStartBlock := startBlock // start block for the first signing
	signatures := make([]*tecdsa.Signature, len(messages))
	endBlocks := make([]uint64, len(messages))

	for i, message := range messages {
		signingBatchMessageLogger := signingBatchLogger.With(
			zap.String("message", fmt.Sprintf("0x%x", message)),
			zap.String("index", fmt.Sprintf("%v/%v", i+1, len(messages))),
		)

		signingBatchMessageLogger.Infof("generating signature for message")

		if i > 0 {
			signingStartBlock = endBlocks[i-1] + signingBatchInterludeBlocks
		}

		signature, endBlock, err := se.sign(ctx, message, signingStartBlock)
		if err != nil {
			return nil, err
		}

		signingBatchMessageLogger.Infof(
			"generated signature [%v] for message at block [%v]",
			signature,
			endBlock,
		)

		signatures[i] = signature
		endBlocks[i] = endBlock
	}

	return signatures, nil
}

// sign performs the signing process for the given message. The process is
// triggered according to the given start block. If the message cannot be signed
// within a limited time window, an error is returned. If the message was
// signed successfully, this function returns the signature along with the
// block at which the signature was calculated. This end block is common for
// all wallet signers so can be used as a synchronization point.
func (se *signingExecutor) sign(
	ctx context.Context,
	message *big.Int,
	startBlock uint64,
) (*tecdsa.Signature, uint64, error) {
	if lockAcquired := se.lock.TryAcquire(1); !lockAcquired {
		return nil, 0, fmt.Errorf("signing executor is busy")
	}
	defer se.lock.Release(1)

	wallet := se.wallet()
	// Actual wallet signing group size may be different from the
	// `GroupSize` parameter of the chain config.
	signingGroupSize := len(wallet.signingGroupOperators)
	// The dishonest threshold for the wallet signing group must be
	// also calculated using the actual wallet signing group size.
	signingGroupDishonestThreshold := signingGroupSize -
		se.chainConfig.HonestThreshold

	walletPublicKeyBytes, err := marshalPublicKey(wallet.publicKey)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot marshal wallet public key: [%v]", err)
	}

	loopTimeoutBlock := startBlock +
		uint64(se.signingAttemptsLimit*signingAttemptMaximumBlocks())

	signingLogger := logger.With(
		zap.String("wallet", fmt.Sprintf("0x%x", walletPublicKeyBytes)),
		zap.String("message", fmt.Sprintf("0x%x", message)),
		zap.Uint64("signingStartBlock", startBlock),
		zap.Uint64("signingTimeoutBlock", loopTimeoutBlock),
	)

	type signerOutcome struct {
		signature *tecdsa.Signature
		endBlock  uint64
		err       error
	}

	signerOutcomesChan := make(chan *signerOutcome, len(se.signers))

	for _, currentSigner := range se.signers {
		go func(signer *signer) {
			se.protocolLatch.Lock()
			defer se.protocolLatch.Unlock()

			announcer := announcer.New(
				fmt.Sprintf("%v-%v", ProtocolName, "signing"),
				se.chainConfig.GroupSize,
				se.broadcastChannel,
				se.membershipValidator,
			)

			syncer := newSigningSyncer(
				se.chainConfig.GroupSize,
				se.broadcastChannel,
				se.membershipValidator,
			)

			retryLoop := newSigningRetryLoop(
				signingLogger,
				message,
				startBlock,
				signer.signingGroupMemberIndex,
				wallet.signingGroupOperators,
				se.chainConfig,
				announcer,
				syncer,
			)

			// Set up the loop timeout signal.
			loopCtx, cancelLoopCtx := withCancelOnBlock(
				ctx,
				loopTimeoutBlock,
				se.waitForBlockFn,
			)

			loopResult, err := retryLoop.start(
				loopCtx,
				se.waitForBlockFn,
				func(attempt *signingAttemptParams) (*signing.Result, uint64, error) {
					signingAttemptLogger := signingLogger.With(
						zap.Uint("attemptNumber", attempt.number),
						zap.Uint64("attemptStartBlock", attempt.startBlock),
						zap.Uint64("attemptTimeoutBlock", attempt.timeoutBlock),
					)

					signingAttemptLogger.Infof(
						"[member:%v] starting signing protocol "+
							"with [%v] group members (excluded: [%v])",
						signer.signingGroupMemberIndex,
						signingGroupSize-len(attempt.excludedMembersIndexes),
						attempt.excludedMembersIndexes,
					)

					// Set up the attempt timeout signal.
					attemptCtx, _ := withCancelOnBlock(
						loopCtx,
						attempt.timeoutBlock,
						se.waitForBlockFn,
					)

					sessionID := fmt.Sprintf(
						"%v-%v",
						message.Text(16),
						attempt.number,
					)

					result, err := signing.Execute(
						attemptCtx,
						signingAttemptLogger,
						message,
						sessionID,
						signer.signingGroupMemberIndex,
						signer.privateKeyShare,
						signingGroupSize,
						signingGroupDishonestThreshold,
						attempt.excludedMembersIndexes,
						se.broadcastChannel,
						se.membershipValidator,
					)
					if err != nil {
						return nil, 0, err
					}

					endBlock, err := se.currentBlockFn()
					if err != nil {
						return nil, 0, err
					}

					return result, endBlock, nil
				},
			)
			if err != nil {
				// Signer failed so there is no point to hold the loopCtx.
				// Cancel it regardless of their timeout.
				cancelLoopCtx()

				signingLogger.Errorf(
					"[member:%v] all retries for the signing failed; "+
						"giving up: [%v]",
					signer.signingGroupMemberIndex,
					err,
				)

				signerOutcomesChan <- &signerOutcome{
					signature: nil,
					endBlock:  0,
					err:       err,
				}

				return
			}

			// Do not cancel the loopCtx upon function exit immediately and
			// continue to broadcast sync messages until the successful attempt
			// timeout. This way we maximize the chance that other members,
			// especially the ones not participating in the successful signature
			// attempt get synced as well.
			go func() {
				defer cancelLoopCtx()

				err := se.waitForBlockFn(
					loopCtx,
					loopResult.attemptTimeoutBlock,
				)
				if err != nil {
					signingLogger.Warnf(
						"[member:%v] failed waiting for signing "+
							"loop stop signal: [%v]",
						signer.signingGroupMemberIndex,
						err,
					)
				}
			}()

			signingLogger.Infof(
				"[member:%v] generated signature [%v] at block [%v]",
				signer.signingGroupMemberIndex,
				loopResult.result.Signature,
				loopResult.latestEndBlock,
			)

			signerOutcomesChan <- &signerOutcome{
				signature: loopResult.result.Signature,
				endBlock:  loopResult.latestEndBlock,
				err:       nil,
			}
		}(currentSigner)
	}

	signerOutcomes := make([]*signerOutcome, 0)

	for {
		select {
		case outcome := <-signerOutcomesChan:
			signerOutcomes = append(signerOutcomes, outcome)

			if len(signerOutcomes) == len(se.signers) {
				outcomes := slices.CompactFunc(
					signerOutcomes,
					func(o1, o2 *signerOutcome) bool {
						return o1.signature.Equals(o2.signature) &&
							o1.endBlock == o2.endBlock &&
							reflect.DeepEqual(o1.err, o2.err)
					},
				)

				if len(outcomes) != 1 {
					return nil, 0, fmt.Errorf("signers came to different outcomes")
				}

				commonOutcome := outcomes[0]

				if err := commonOutcome.err; err != nil {
					return nil, 0, fmt.Errorf("signers failed: [%v]", err)
				}

				return commonOutcome.signature, commonOutcome.endBlock, nil
			}
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		}
	}
}

func (se *signingExecutor) wallet() wallet {
	// All signers belong to one wallet. Take that wallet from the
	// first signer.
	return se.signers[0].wallet
}
