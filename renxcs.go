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
	Mint(value *big.Int, sig []byte) error
	Burn(to []byte, value *big.Int, sig []byte) error
}
