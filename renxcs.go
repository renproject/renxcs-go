package renxcs

import (
	"crypto/rsa"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

const MinMintValue = 100000

type NativeBinder interface {
	Lock(address string, value *big.Int) error
	Unlock(address string) error
}

type Bridge interface {
	Claim(rsaPubKey rsa.PublicKey, btcAddr common.Address) error
	Burn(rsaPubKey rsa.PublicKey, btcAddr common.Address) error
	UnusedAddress() (common.Address, error)
	Verify(rsaPubKey rsa.PublicKey, address common.Address) error
	OwnerOf(address common.Address) (common.Address, error)
}
