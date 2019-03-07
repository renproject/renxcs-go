package crypto

import "crypto"

type Decrypter interface {
	Marshal() ([]byte, error)
	Decrypt(msg []byte) ([]byte, error)
	PrivKey() crypto.PrivateKey

	Encrypter() Encrypter
}

type Encrypter interface {
	Marshal() ([]byte, error)
	Encrypt(msg []byte) ([]byte, error)
	PubKey() crypto.PublicKey
}

type Signer interface {
	Marshal() ([]byte, error)
	Sign(msgHash [32]byte) ([]byte, error)
	PrivKey() crypto.PrivateKey

	Verifier() Verifier
}

type Verifier interface {
	Marshal() ([]byte, error)
	Verify(sig []byte, msgHash [32]byte) error
	PubKey() crypto.PublicKey
}
