package wbtc

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/renproject/libeth-go"
	"github.com/renproject/renxcs-go"
	renCrypto "github.com/renproject/renxcs-go/crypto"
)

type eth struct {
	wbtc    *WrappedBTC
	account libeth.Account
}

func Deploy(account libeth.Account, ethAddrs, btcAddrs []common.Address) (common.Address, error) {
	var contractAddr common.Address
	_, err := account.Transact(
		context.Background(),
		libeth.Fast,
		nil,
		func(tops *bind.TransactOpts) (*types.Transaction, error) {
			addr, tx, _, err := DeployWrappedBTC(tops, bind.ContractBackend(account.EthClient()), ethAddrs, btcAddrs)
			if err != nil {
				return nil, err
			}
			contractAddr = addr
			return tx, nil
		},
		nil,
		0,
	)
	return contractAddr, err
}

func NewWBTCBinder(account libeth.Account) (renxcs.Bridge, error) {
	address, err := account.ReadAddress("REN.WBTC")
	if err != nil {
		return nil, err
	}
	wbtc, err := NewWrappedBTC(address, bind.ContractBackend(account.EthClient()))
	if err != nil {
		return nil, err
	}
	return &eth{
		wbtc:    wbtc,
		account: account,
	}, nil
}

// Claim a bitcoin P2PKH address on Ethereum
func (ethereum *eth) Claim(rsaPubKey rsa.PublicKey, btcAddr common.Address) error {
	encryptor, err := renCrypto.NewRSAEncrypter(rsaPubKey)
	if err != nil {
		return err
	}
	hash, err := encryptor.Hash(sha256.New())
	if err != nil {
		return err
	}
	hash32 := [32]byte{}
	copy(hash32[:], hash)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	_, err = ethereum.account.Transact(
		ctx,
		libeth.Fast,
		nil,
		func(tops *bind.TransactOpts) (*types.Transaction, error) {
			return ethereum.wbtc.Claim(tops, btcAddr, hash32)
		},
		nil,
		0,
	)
	return err
}

// Burn a bitcoin P2PKH address on Ethereum
func (ethereum *eth) Burn(rsaPubKey rsa.PublicKey, btcAddr common.Address) error {
	encryptor, err := renCrypto.NewRSAEncrypter(rsaPubKey)
	if err != nil {
		return err
	}
	hash, err := encryptor.Hash(sha256.New())
	if err != nil {
		return err
	}
	hash32 := [32]byte{}
	copy(hash32[:], hash)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	_, err = ethereum.account.Transact(
		ctx,
		libeth.Fast,
		nil,
		func(tops *bind.TransactOpts) (*types.Transaction, error) {
			return ethereum.wbtc.Burn(tops, btcAddr, hash32)
		},
		nil,
		0,
	)
	return err
}

func (ethereum *eth) UnusedAddress() (common.Address, error) {
	return ethereum.wbtc.UnusedAddress(&bind.CallOpts{})
}

func (ethereum *eth) Verify(rsaPubKey rsa.PublicKey, address common.Address) error {
	encryptor, err := renCrypto.NewRSAEncrypter(rsaPubKey)
	if err != nil {
		return err
	}
	hash, err := encryptor.Hash(sha256.New())
	if err != nil {
		return err
	}

	h32, err := ethereum.wbtc.RSAPubKeys(&bind.CallOpts{}, address)
	if !bytes.Equal(hash, h32[:]) {
		return fmt.Errorf("incorrect rsa pulic key")
	}
	return nil
}

func (ethereum *eth) OwnerOf(address common.Address) (common.Address, error) {
	return ethereum.wbtc.OwnerOf(&bind.CallOpts{}, new(big.Int).SetBytes(address.Bytes()))
}
