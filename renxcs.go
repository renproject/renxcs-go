package renxcs

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const MinMintValue = 100000

type NativeBinder interface {
	Build(address string, value *big.Int) (string, []byte, error)
	Submit(tx []byte) error
}

type ForeignBinder interface {
	Mint(*big.Int, [32]byte, *big.Int, *big.Int) error
	Burn(to []byte, value *big.Int) error
	Commitment(to common.Address, value *big.Int, hash [32]byte) ([32]byte, error)
	VerifySignature(to common.Address, value *big.Int, hash [32]byte, r, s *big.Int, v byte) (bool, error)
}
