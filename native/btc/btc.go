package btc

import (
	"bytes"
	"context"
	"math/big"
	"time"

	"github.com/btcsuite/btcd/wire"
	"github.com/renproject/libbtc-go"
)

type bitcoin struct {
	libbtc.Account
}

func NewBitcoinBinder(account libbtc.Account) *bitcoin {
	return &bitcoin{
		Account: account,
	}
}

func (bitcoin *bitcoin) Build(address string, value *big.Int) (string, []byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	return bitcoin.Account.BuildTransfer(ctx, address, value.Int64(), libbtc.Fast, false)
}

func (bitcoin *bitcoin) Submit(tx []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	msgTx := wire.NewMsgTx(2)
	msgTx.Deserialize(bytes.NewBuffer(tx))
	return bitcoin.PublishTransaction(ctx, msgTx)
}
