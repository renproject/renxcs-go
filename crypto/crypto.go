package crypto

type Decryptor interface {
	Marshal() ([]byte, error)
	Decrypt(msg []byte) ([]byte, error)

	Encryptor() Encryptor
}

type Encryptor interface {
	Marshal() ([]byte, error)
	Encrypt(msg []byte) ([]byte, error)
}
