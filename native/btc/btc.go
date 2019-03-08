package btc

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/renproject/libbtc-go"
	"github.com/renproject/renxcs-go"
)

type bitcoin struct {
	libbtc.Account
}

func NewBitcoinBinder(account libbtc.Account) renxcs.NativeBinder {
	return &bitcoin{
		Account: account,
	}
}

func (bitcoin *bitcoin) Lock(address string, value *big.Int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	txHash, _, err := bitcoin.Transfer(ctx, address, value.Int64(), libbtc.Fast, false)
	if err != nil {
		return err
	}
	fmt.Println(bitcoin.FormatTransactionView("Successfully locked funds on bitcoin", txHash))
	return nil
}

func (bitcoin *bitcoin) Unlock(address string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	txHash, _, err := bitcoin.Transfer(ctx, address, renxcs.MinMintValue, libbtc.Fast, true)
	if err != nil {
		return err
	}
	fmt.Println(bitcoin.FormatTransactionView("Successfully unlocked funds on bitcoin", txHash))
	return nil
}
