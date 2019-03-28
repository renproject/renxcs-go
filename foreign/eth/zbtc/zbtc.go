package zbtc

import (
	"math/big"

	"github.com/renproject/libeth-go"
)

type zbtc struct {
	account libeth.Account
}

func NewZBTCBinder() *zbtc {
	return &zbtc{}
}

func Claim(txhash [32]byte, value *big.Int) error {
	return nil
}

func Mint(txHash [32]byte, sigR, sigS *big.Int) error {
	return nil
}
