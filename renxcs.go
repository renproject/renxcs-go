package renxcs

import (
	"math/big"
)

const MinMintValue = 100000

type NativeBinder interface {
	Build(address string, value *big.Int) (string, []byte, error)
	Submit(tx []byte) error
}

type ForeignBinder interface {
	Claim(txhash [32]byte, value *big.Int) error
	Mint(txHash [32]byte, sigR, sigS *big.Int) error
}
