package zbtc

import (
	"context"
	"encoding/hex"
	"fmt"
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

func Connect(account libeth.Account, zbtcAddress common.Address) (renxcs.ForeignBinder, error) {
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

func (zbtc *zbtc) Mint(value *big.Int, hash [32]byte, r, s *big.Int) error {
	// r := [32]byte{}
	// s := [32]byte{}
	// v := byte(0x00)
	// copy(r[:], sigR.Bytes())
	// copy(s[:], sigS.Bytes())
	// sig := append([]byte{v}, append(r[:], s[:]...)...)
	sig := append(r.Bytes(), append(s.Bytes(), 0x0)...)
	sig1 := append(r.Bytes(), append(s.Bytes(), 0x1)...)

	ok, err := zbtc.bindings.VerifySig(&bind.CallOpts{}, zbtc.account.Address(), value,
		hash, sig)
	if err != nil {
		return err
	}
	if !ok {
		ok, err := zbtc.bindings.VerifySig(&bind.CallOpts{}, zbtc.account.Address(), value,
			hash, sig1)
		if !ok || err != nil {
			return fmt.Errorf("invalid signature")
		}
		sig = sig1
	}

	_, err = zbtc.account.Transact(
		context.Background(),
		libeth.Fast,
		nil,
		func(tops *bind.TransactOpts) (*types.Transaction, error) {
			fmt.Println(hex.EncodeToString(sig))
			return zbtc.bindings.Mint(tops, zbtc.account.Address(), value, hash, sig)
		},
		nil,
		0,
	)
	return err
}

func (zbtc *zbtc) Burn(to []byte, value *big.Int) error {
	_, err := zbtc.account.Transact(
		context.Background(),
		libeth.Fast,
		nil,
		func(tops *bind.TransactOpts) (*types.Transaction, error) {
			return zbtc.bindings.Burn(tops, to, value)
		},
		nil,
		0,
	)
	return err
}

func (zbtc *zbtc) Commitment(to common.Address, value *big.Int, hash [32]byte) ([32]byte, error) {
	return zbtc.bindings.Commitment(&bind.CallOpts{}, to, value, hash)
}

func (zbtc *zbtc) VerifySignature(to common.Address, value *big.Int, hash [32]byte, r, s *big.Int, v byte) (bool, error) {
	return zbtc.bindings.VerifySig(&bind.CallOpts{}, to, value, hash, append(r.Bytes(), append(s.Bytes(), v)...))
}
