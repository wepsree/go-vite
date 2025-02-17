package api

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/vitelabs/go-vite"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/interfaces"
	ledger "github.com/vitelabs/go-vite/interfaces/core"
	"github.com/vitelabs/go-vite/ledger/chain"
	"github.com/vitelabs/go-vite/ledger/generator"
	"github.com/vitelabs/go-vite/ledger/pool"
	"github.com/vitelabs/go-vite/vm/contracts/dex"
	"github.com/vitelabs/go-vite/wallet"
	"github.com/vitelabs/go-vite/wallet/entropystore"
)

type HexSignedTuple struct {
	Message    string `json:"message"`
	SignedData string `json:"signedData"`
	Pubkey     string `json:"pubkey"`
}

type NewStoreResponse struct {
	Mnemonic    string        `json:"mnemonic"`
	PrimaryAddr types.Address `json:"primaryAddr"`
	Filename    string        `json:"filename"`
}

type FindAddrResult struct {
	EntropyStoreFile string `json:"entropyStoreFile"`
	Index            uint32 `json:"index"`
}

type DeriveResult struct {
	Bip44Path  string        `json:"bip44Path"`
	Address    types.Address `json:"address"`
	PrivateKey []byte        `json:"privateKey"`
}

type CreateTransferTxParms struct {
	EntropystoreFile *string           `json:"entropystoreFile,omitempty"`
	SelfAddr         types.Address     `json:"selfAddr"`
	ToAddr           types.Address     `json:"toAddr"`
	TokenTypeId      types.TokenTypeId `json:"tokenTypeId"`
	Passphrase       string            `json:"passphrase"`
	Amount           string            `json:"amount"`
	Data             []byte            `json:"data,omitempty"`
	Difficulty       *string           `json:"difficulty,omitempty"`
}

type IsMayValidKeystoreFileResponse struct {
	Maybe      bool
	MayAddress types.Address
}

func NewWalletApi(vite *vite.Vite) *WalletApi {
	return &WalletApi{
		wallet:    vite.WalletManager(),
		chain:     vite.Chain(),
		pool:      vite.Pool(),
		consensus: vite.Consensus(),
	}
}

type WalletApi struct {
	wallet    *wallet.Manager
	chain     chain.Chain
	pool      pool.Writer
	consensus generator.Consensus
}

func (m WalletApi) String() string {
	return "WalletApi"
}

func (m WalletApi) ListAllEntropyFiles() []string {
	return m.wallet.ListAllEntropyFiles()
}

func (m WalletApi) ListEntropyFilesInStandardDir() ([]string, error) {
	return m.wallet.ListEntropyFilesInStandardDir()
}

func (m WalletApi) ListEntropyStoreAddresses(entropyStore string, from, to uint32) ([]types.Address, error) {
	if from > to {
		return nil, errors.New("from value > to")
	}

	manager, e := m.wallet.GetEntropyStoreManager(entropyStore)
	if e != nil {
		return nil, e
	}
	return manager.ListAddress(from, to)
}

func (m WalletApi) NewMnemonicAndEntropyStore(passphrase string) (*NewStoreResponse, error) {
	mnemonic, em, err := m.wallet.NewMnemonicAndEntropyStore(passphrase)
	if err != nil {
		return nil, err
	}

	return &NewStoreResponse{
		Mnemonic:    mnemonic,
		PrimaryAddr: em.GetPrimaryAddr(),
		Filename:    em.GetEntropyStoreFile(),
	}, nil
}
func (m WalletApi) DeriveByFullPath(entropyStore string, fullpath string) (*DeriveResult, error) {
	manager, e := m.wallet.GetEntropyStoreManager(entropyStore)
	if e != nil {
		return nil, e
	}
	_, key, e := manager.DeriveForFullPath(fullpath)
	if e != nil {
		return nil, e
	}

	address, err := key.Address()
	if err != nil {
		return nil, err
	}

	privateKey, err := key.PrivateKey()
	if err != nil {
		return nil, err
	}

	return &DeriveResult{
		Bip44Path:  fullpath,
		Address:    *address,
		PrivateKey: privateKey,
	}, nil
}

func (m WalletApi) DeriveByIndex(entropyStore string, index uint32) (*DeriveResult, error) {
	manager, e := m.wallet.GetEntropyStoreManager(entropyStore)
	if e != nil {
		return nil, e
	}
	path, key, e := manager.DeriveForIndexPath(index)
	if e != nil {
		return nil, e
	}

	address, err := key.Address()
	if err != nil {
		return nil, err
	}

	privateKey, err := key.PrivateKey()
	if err != nil {
		return nil, err
	}

	return &DeriveResult{
		Bip44Path:  path,
		Address:    *address,
		PrivateKey: privateKey,
	}, nil
}

func (m WalletApi) RecoverEntropyStoreFromMnemonic(mnemonic string, newPassphrase string) (*NewStoreResponse, error) {
	em, e := m.wallet.RecoverEntropyStoreFromMnemonic(mnemonic, newPassphrase)
	if e != nil {
		return nil, e
	}
	return &NewStoreResponse{
		Mnemonic:    mnemonic,
		PrimaryAddr: em.GetPrimaryAddr(),
		Filename:    em.GetEntropyStoreFile(),
	}, nil
}

func (m WalletApi) GlobalCheckAddrUnlocked(addr types.Address) bool {
	return m.wallet.GlobalCheckAddrUnlock(addr)
}

func (m WalletApi) IsAddrUnlocked(entropyStore string, addr types.Address) bool {
	manager, e := m.wallet.GetEntropyStoreManager(entropyStore)
	if e != nil {
		return false
	}
	return manager.IsAddrUnlocked(addr)
}

func (m WalletApi) RefreshCache() {
	m.wallet.RefreshCache()
}

func (m WalletApi) ExtractMnemonic(entropyStore string, passphrase string) (string, error) {
	return m.wallet.ExtractMnemonic(entropyStore, passphrase)
}

func (m WalletApi) FindAddrWithPassphrase(entropyStore string, passphrase string, addr types.Address) (findResult *FindAddrResult, e error) {
	manager, e := m.wallet.GetEntropyStoreManager(entropyStore)
	if e != nil {
		return nil, e
	}
	_, index, e := manager.FindAddrWithPassphrase(passphrase, addr)
	return &FindAddrResult{
		EntropyStoreFile: manager.GetEntropyStoreFile(),
		Index:            index,
	}, nil

}

func (m WalletApi) FindAddr(entropyStore string, addr types.Address) (findResult *FindAddrResult, e error) {
	manager, e := m.wallet.GetEntropyStoreManager(entropyStore)
	if e != nil {
		return nil, e
	}
	_, index, e := manager.FindAddr(addr)
	return &FindAddrResult{
		EntropyStoreFile: manager.GetEntropyStoreFile(),
		Index:            index,
	}, nil
}

func (m WalletApi) GlobalFindAddr(addr types.Address) (findResult *FindAddrResult, e error) {
	path, _, index, e := m.wallet.GlobalFindAddr(addr)
	if e != nil {
		return nil, e
	}
	return &FindAddrResult{
		EntropyStoreFile: path,
		Index:            index,
	}, nil
}

// Deprecated
func (m WalletApi) GlobalFindAddrWithPassphrase(addr types.Address, passphrase string) (findResult *FindAddrResult, e error) {
	path, _, index, e := m.wallet.GlobalFindAddrWithPassphrase(addr, passphrase)
	if e != nil {
		return nil, e
	}
	return &FindAddrResult{
		EntropyStoreFile: path,
		Index:            index,
	}, nil
}

func (m WalletApi) AddEntropyStore(filename string) error {
	return m.wallet.AddEntropyStore(filename)
}

func (m WalletApi) SignData(addr types.Address, hexMsg string) (*HexSignedTuple, error) {
	hash, err := types.HexToHash(hexMsg)
	if err != nil {
		return nil, err
	}
	account, e := m.wallet.Account(addr)
	if e != nil {
		return nil, e
	}

	signedData, pubkey, err := account.Sign(hash.Bytes())
	if err != nil {
		return nil, err
	}

	t := HexSignedTuple{
		Message:    hexMsg,
		Pubkey:     hex.EncodeToString(pubkey),
		SignedData: hex.EncodeToString(signedData),
	}

	return &t, nil
}

func (m WalletApi) CreateTxWithPassphrase(params CreateTransferTxParms) (*types.Hash, error) {
	if !checkTxToAddressAvailable(params.ToAddr) {
		return nil, errors.New("ToAddress is invalid")
	}
	if params.ToAddr == types.AddressDexFund && !dex.VerifyNewOrderPriceForRpc(params.Data) {
		return nil, dex.InvalidOrderPriceErr
	}
	amount, ok := new(big.Int).SetString(params.Amount, 10)
	if !ok {
		return nil, ErrStrToBigInt
	}
	if err := checkTokenIdValid(m.chain, &params.TokenTypeId); err != nil {
		return nil, err
	}
	var difficulty *big.Int = nil
	if params.Difficulty != nil {
		difficulty, ok = new(big.Int).SetString(*params.Difficulty, 10)
		if !ok {
			return nil, ErrStrToBigInt
		}
	}

	msg := &interfaces.IncomingMessage{
		BlockType:      ledger.BlockTypeSendCall,
		AccountAddress: params.SelfAddr,
		ToAddress:      &params.ToAddr,
		TokenId:        &params.TokenTypeId,
		Amount:         amount,
		Fee:            nil,
		Difficulty:     difficulty,
		Data:           params.Data,
	}

	addrState, err := generator.GetAddressStateForGenerator(m.chain, &msg.AccountAddress)
	if err != nil || addrState == nil {
		return nil, fmt.Errorf("failed to get addr state for generator, err:%v", err)
	}
	g, e := generator.NewGenerator(m.chain, m.consensus, msg.AccountAddress, addrState.LatestSnapshotHash, addrState.LatestAccountHash)
	if e != nil {
		return nil, e
	}

	account, err := m.wallet.AccountSearch(params.EntropystoreFile, msg.AccountAddress, params.Passphrase)
	if err != nil {
		return nil, err
	}
	result, e := g.GenerateWithMessage(msg, &msg.AccountAddress, account.Sign)

	if e != nil {
		return nil, e
	}
	if result.Err != nil {
		return nil, result.Err
	}
	if result.VMBlock != nil {
		return &result.VMBlock.AccountBlock.Hash, m.pool.AddDirectAccountBlock(params.SelfAddr, result.VMBlock)
	} else {
		return nil, errors.New("generator gen an empty block")
	}

}

func (m WalletApi) SignDataWithPassphrase(addr types.Address, hexMsg string, passphrase string) (*HexSignedTuple, error) {
	hash, err := types.HexToHash(hexMsg)
	if err != nil {
		return nil, err
	}
	account, err := m.wallet.AccountSearch(nil, addr, passphrase)
	if err != nil {
		return nil, err
	}
	signedData, pubkey, err := account.Sign(hash.Bytes())
	if err != nil {
		return nil, err
	}

	t := HexSignedTuple{
		Message:    hexMsg,
		Pubkey:     hex.EncodeToString(pubkey),
		SignedData: hex.EncodeToString(signedData),
	}

	return &t, nil
}

func (m WalletApi) IsMayValidKeystoreFile(path string) IsMayValidKeystoreFileResponse {
	b, addr, _ := entropystore.IsMayValidEntropystoreFile(path)
	if b && addr != nil {
		return IsMayValidKeystoreFileResponse{
			true, *addr,
		}
	}
	return IsMayValidKeystoreFileResponse{
		false, types.Address{},
	}
}

func (m WalletApi) GetDataDir() string {
	return m.wallet.GetDataDir()
}

func (m WalletApi) GetPrivateKey(entropyStore string, passphrase string) (*string, error) {
	manager, e := m.wallet.GetEntropyStoreManager(entropyStore)
	if e != nil {
		return nil, e
	}
	err := manager.Unlock(passphrase)
	if err != nil {
		return nil, err
	}
	pk, err := manager.GetPrivateKey(manager.GetPrimaryAddr())
	if err != nil {
		return nil, err
	}
	pkStr := hex.EncodeToString(pk)
	return &pkStr, nil
}
