package tbtc

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/keep-network/keep-core/pkg/bitcoin"

	"github.com/keep-network/keep-common/pkg/persistence"
)

// CalculateWalletIDFunc calculates the ECDSA wallet ID based on the provided
// wallet public key.
type CalculateWalletIdFunc func(walletPublicKey *ecdsa.PublicKey) ([32]byte, error)

// walletRegistry is the component that holds the data of the wallets managed
// by the given node. All functions of the registry are safe for concurrent use.
type walletRegistry struct {
	// mutex is a single struct-wide lock that ensures all functions
	// of the registry are thread-safe.
	mutex sync.Mutex

	// walletCache is a cache of maintained wallets. The cache's key is the
	// uncompressed public key of the given wallet.
	walletCache map[string]*walletCacheValue

	// walletStorage is the handle to the wallet storage responsible for
	// wallet persistence.
	walletStorage *walletStorage

	// calculateWalletIdFunc calculates the ECDSA wallet ID based on the
	// provided wallet public key.
	calculateWalletIdFunc CalculateWalletIdFunc
}

type walletCacheValue struct {
	// SHA-256+RIPEMD-160 hash computed over the compressed ECDSA public key of
	// the wallet.
	walletPublicKeyHash [20]byte
	// ECDSA wallet ID calculated as the keccak256 of the 64-byte-long
	// concatenation of the X and Y coordinates of the wallet's public key.
	walletID [32]byte
	// Array of wallet signers controlled by this node.
	signers []*signer
}

// newWalletRegistry creates a new instance of the walletRegistry.
func newWalletRegistry(
	persistence persistence.ProtectedHandle,
	calculateWalletIdFunc CalculateWalletIdFunc,
) (*walletRegistry, error) {
	walletStorage := newWalletStorage(persistence)

	// Pre-populate the wallet cache using the wallet storage.
	walletCache := make(map[string]*walletCacheValue)
	walletSigners := walletStorage.loadSigners()
	if len(walletSigners) > 0 {
		for walletStorageKey, signers := range walletSigners {
			// We need to extract the wallet from the signers array. The
			// walletStorage.loadSigners function guarantees there is always
			// at least one signer for the given walletStorageKey so, we
			// don't need to check len(signers). Then, we can just take the
			// wallet from the first signer as the wallet is same for all of
			// them.
			wallet := signers[0].wallet
			walletPublicKeyHash := bitcoin.PublicKeyHash(wallet.publicKey)
			walletID, err := calculateWalletIdFunc(wallet.publicKey)
			if err != nil {
				return nil, fmt.Errorf(
					"error while calculating wallet ID for wallet with public "+
						"key hash [0x%x]: [%v]",
					walletPublicKeyHash,
					err,
				)
			}

			walletCache[walletStorageKey] = &walletCacheValue{
				walletPublicKeyHash: walletPublicKeyHash,
				walletID:            walletID,
				signers:             signers,
			}

			logger.Infof(
				"wallet signing group [0x%v] loaded from storage "+
					"with [%v] members and wallet public key hash [0x%x]",
				walletStorageKey,
				len(signers),
				walletPublicKeyHash,
			)
		}
	} else {
		logger.Infof("no wallet signing groups found in the storage")
	}

	return &walletRegistry{
		walletCache:           walletCache,
		walletStorage:         walletStorage,
		calculateWalletIdFunc: calculateWalletIdFunc,
	}, nil
}

// getWalletsPublicKeys returns public keys of all registered wallets.
func (wr *walletRegistry) getWalletsPublicKeys() []*ecdsa.PublicKey {
	wr.mutex.Lock()
	defer wr.mutex.Unlock()

	keys := make([]*ecdsa.PublicKey, 0)
	for _, value := range wr.walletCache {
		// We can take the wallet from the first signer. All signers for the
		// given cache value belong to the same wallet.
		keys = append(keys, value.signers[0].wallet.publicKey)
	}

	return keys
}

// registerSigner registers the given signer using in the walletRegistry.
func (wr *walletRegistry) registerSigner(signer *signer) error {
	wr.mutex.Lock()
	defer wr.mutex.Unlock()

	err := wr.walletStorage.saveSigner(signer)
	if err != nil {
		return fmt.Errorf("cannot save signer in the storage: [%w]", err)
	}

	walletStorageKey := getWalletStorageKey(signer.wallet.publicKey)

	// If the wallet cache does not have the given entry yet, initialize
	// the value and compute the wallet ID and wallet public key hash. This way,
	// the hashes are computed only once. No need to initialize signers slice as
	// appending works with nil values.
	if _, ok := wr.walletCache[walletStorageKey]; !ok {
		walletID, err := wr.calculateWalletIdFunc(signer.wallet.publicKey)
		if err != nil {
			return fmt.Errorf("cannot calculate wallet ID: [%v]", err)
		}

		wr.walletCache[walletStorageKey] = &walletCacheValue{
			walletPublicKeyHash: bitcoin.PublicKeyHash(signer.wallet.publicKey),
			walletID:            walletID,
		}
	}

	wr.walletCache[walletStorageKey].signers = append(
		wr.walletCache[walletStorageKey].signers,
		signer,
	)

	return nil
}

// getSigners gets all signers for the given wallet held by the walletRegistry.
func (wr *walletRegistry) getSigners(
	walletPublicKey *ecdsa.PublicKey,
) []*signer {
	wr.mutex.Lock()
	defer wr.mutex.Unlock()

	if value, ok := wr.walletCache[getWalletStorageKey(walletPublicKey)]; ok {
		return value.signers
	}

	return nil
}

// getWalletByPublicKeyHash gets the given wallet by its 20-byte wallet
// public key hash. Second boolean return value denotes whether the wallet
// was found in the registry or not.
func (wr *walletRegistry) getWalletByPublicKeyHash(
	walletPublicKeyHash [20]byte,
) (wallet, bool) {
	wr.mutex.Lock()
	defer wr.mutex.Unlock()

	for _, value := range wr.walletCache {
		if value.walletPublicKeyHash == walletPublicKeyHash {
			// All signers belong to one wallet. Take that wallet from the
			// first signer.
			return value.signers[0].wallet, true
		}
	}

	return wallet{}, false
}

// getWalletByID gets the given wallet by its 32-byte wallet ID. Second boolean
// return value denotes whether the wallet was found in the registry or not.
func (wr *walletRegistry) getWalletByID(walletID [32]byte) (wallet, bool) {
	wr.mutex.Lock()
	defer wr.mutex.Unlock()

	for _, value := range wr.walletCache {
		if value.walletID == walletID {
			// All signers belong to one wallet. Take that wallet from the
			// first signer.
			return value.signers[0].wallet, true
		}
	}

	return wallet{}, false
}

// archiveWallet archives the wallet with the given public key hash. The wallet
// data is removed from the wallet cache and the entire wallet storage directory
// is moved to the archive directory.
func (wr *walletRegistry) archiveWallet(
	walletPublicKeyHash [20]byte,
) error {
	wr.mutex.Lock()
	defer wr.mutex.Unlock()

	var walletPublicKey *ecdsa.PublicKey

	for _, value := range wr.walletCache {
		if value.walletPublicKeyHash == walletPublicKeyHash {
			// All signers belong to one wallet. Take the wallet public key from
			//  the first signer.
			walletPublicKey = value.signers[0].wallet.publicKey
		}
	}

	if walletPublicKey == nil {
		return fmt.Errorf("wallet not found in the wallet cache")
	}

	walletStorageKey := getWalletStorageKey(walletPublicKey)

	// Archive the entire wallet storage.
	err := wr.walletStorage.archiveWallet(walletStorageKey)
	if err != nil {
		return fmt.Errorf("could not archive wallet: [%v]", err)
	}

	// Remove the wallet from the wallet cache.
	delete(wr.walletCache, walletStorageKey)

	return nil
}

// walletStorage is the component that persists data of the wallets managed
// by the given node using the underlying persistence layer. It should be
// used directly only by the walletRegistry.
type walletStorage struct {
	// persistence is the handle to the underlying persistence layer.
	persistence persistence.ProtectedHandle
}

// newWalletStorage creates a new instance of the walletStorage.
func newWalletStorage(persistence persistence.ProtectedHandle) *walletStorage {
	return &walletStorage{persistence}
}

// saveSigner saves the given signer using the underlying persistence layer
// of the walletStorage. It does not add the signer to any in-memory cache
// and should not be called from any other place than walletRegistry.
func (ws *walletStorage) saveSigner(signer *signer) error {
	signerBytes, err := signer.Marshal()
	if err != nil {
		return fmt.Errorf("could not marshal signer: [%w]", err)
	}

	err = ws.persistence.Save(
		signerBytes,
		getWalletStorageKey(signer.wallet.publicKey),
		fmt.Sprintf("/membership_%v", signer.signingGroupMemberIndex),
	)
	if err != nil {
		return fmt.Errorf(
			"could not save membership using the "+
				"underlying persistence layer: [%w]",
			err,
		)
	}

	return nil
}

// archiveWallet archives the given wallet data in the underlying persistence
// layer of the walletStorage.
func (ws *walletStorage) archiveWallet(walletStorageKey string) error {
	err := ws.persistence.Archive(walletStorageKey)
	if err != nil {
		return fmt.Errorf(
			"could not archive wallet storage using the "+
				"underlying persistence layer: [%w]",
			err,
		)
	}

	return nil
}

// loadSigners loads all signers stored using the underlying persistence layer.
// This function should not be called from any other place than walletRegistry.
func (ws *walletStorage) loadSigners() map[string][]*signer {
	signersByWallet := make(map[string][]*signer)

	descriptorsChan, errorsChan := ws.persistence.ReadAll()

	// Two goroutines read from descriptors and errors channels and either
	// add the signer to the result map or outputs a log error.
	// The reason for using two goroutines at the same time - one for
	// descriptors and one for errors - is that channels do not have to be
	// buffered, and we do not know in what order the information is written to
	// channels.
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		for descriptor := range descriptorsChan {
			content, err := descriptor.Content()
			if err != nil {
				logger.Errorf(
					"could not get content from file [%v] "+
						"in directory [%v]: [%v]",
					descriptor.Name(),
					descriptor.Directory(),
					err,
				)
				continue
			}

			signer := &signer{}
			if err := signer.Unmarshal(content); err != nil {
				logger.Errorf(
					"could not unmarshal signer from file [%v] "+
						"in directory [%v]: [%v]",
					descriptor.Name(),
					descriptor.Directory(),
					err,
				)
				continue
			}

			walletStorageKey := getWalletStorageKey(signer.wallet.publicKey)

			signersByWallet[walletStorageKey] = append(
				signersByWallet[walletStorageKey],
				signer,
			)
		}

		wg.Done()
	}()

	go func() {
		for err := range errorsChan {
			logger.Errorf(
				"could not load signer from disk: [%v]",
				err,
			)
		}

		wg.Done()
	}()

	wg.Wait()

	return signersByWallet
}

// getWalletStorageKey compute the wallet storage key that is used to identify
// the given wallet for caching and storage purposes.
func getWalletStorageKey(walletPublicKey *ecdsa.PublicKey) string {
	walletPublicKeyBytes := elliptic.Marshal(
		walletPublicKey.Curve,
		walletPublicKey.X,
		walletPublicKey.Y,
	)

	// Strip the 04 prefix to limit the key length to 128 characters in order
	// to make it usable as a directory name.
	return hex.EncodeToString(walletPublicKeyBytes)[2:]
}
