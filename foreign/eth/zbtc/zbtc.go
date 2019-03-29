package zbtc

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/renproject/libeth-go"
	"github.com/renproject/renxcs-go"
)

type zbtc struct {
	account  libeth.Account
	bindings *ZBTC
}

func Connect(account libeth.Account, zbtcAddress common.Address) (*zbtc, error) {
	bindings, err := NewZBTC(zbtcAddress, bind.ContractBackend(account.EthClient()))
	if err != nil {
		return nil, err
	}
	return &zbtc{
		account:  account,
		bindings: bindings,
	}, nil
}

func Deploy(account libeth.Account, owner common.Address) (common.Address, renxcs.ForeignBinder, error) {
	var bindings *ZBTC
	var zbtcAddr common.Address
	_, err := account.Transact(
		context.Background(),
		libeth.Fast,
		nil,
		func(tops *bind.TransactOpts) (*types.Transaction, error) {
			addr, tx, binder, err := DeployZBTC(tops, bind.ContractBackend(account.EthClient()), owner)
			if err != nil {
				return tx, err
			}
			bindings = binder
			zbtcAddr = addr
			return tx, nil
		},
		nil,
		0,
	)
	return zbtcAddr, &zbtc{
		account:  account,
		bindings: bindings,
	}, err
}

func (zbtc *zbtc) Claim(txhash [32]byte, value *big.Int) error {
	_, err := zbtc.account.Transact(
		context.Background(),
		libeth.Fast,
		nil,
		func(tops *bind.TransactOpts) (*types.Transaction, error) {
			return zbtc.bindings.Claim(tops, value, zbtc.account.Address(), txhash)
		},
		nil,
		0,
	)
	return err
}

func (zbtc *zbtc) Mint(txHash [32]byte, sig []byte) error {
	// r := [32]byte{}
	// s := [32]byte{}
	// v := byte(0x00)
	// copy(r[:], sigR.Bytes())
	// copy(s[:], sigS.Bytes())
	// sig := append([]byte{v}, append(r[:], s[:]...)...)

	_, err := zbtc.account.Transact(
		context.Background(),
		libeth.Fast,
		nil,
		func(tops *bind.TransactOpts) (*types.Transaction, error) {
			return zbtc.bindings.Mint(tops, txHash, sig)
		},
		nil,
		0,
	)
	return err
}
