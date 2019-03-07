package crypto

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/binary"
	"math/big"
)

type rsaEncryptor struct {
	*rsa.PublicKey
}

func NewRSAEncryptor(key interface{}) (Encryptor, error) {
	var pubKey *rsa.PublicKey
	switch key := key.(type) {
	case *rsa.PrivateKey:
		pubKey = &key.PublicKey
	case rsa.PrivateKey:
		pubKey = &key.PublicKey
	case *rsa.PublicKey:
		pubKey = key
	case rsa.PublicKey:
		pubKey = &key
	case []byte:
		publicKey, err := unmarshalRSAPubKey(key)
		if err != nil {
			return nil, err
		}
		pubKey = &publicKey
	}
	return &rsaEncryptor{
		PublicKey: pubKey,
	}, nil
}

func (encryptor *rsaEncryptor) Marshal() ([]byte, error) {
	return marshalRSAPubKey(encryptor.PublicKey)
}

func (encryptor *rsaEncryptor) Encrypt(data []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, encryptor.PublicKey, data, nil)
}

func marshalRSAPubKey(publicKey *rsa.PublicKey) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, int64(publicKey.E)); err != nil {
		return []byte{}, err
	}
	if err := binary.Write(buf, binary.BigEndian, publicKey.N.Bytes()); err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

func unmarshalRSAPubKey(pubKeyBytes []byte) (rsa.PublicKey, error) {
	var E int64
	if err := binary.Read(bytes.NewReader(pubKeyBytes[:8]), binary.BigEndian, &E); err != nil {
		return rsa.PublicKey{}, err
	}
	N := new(big.Int).SetBytes(pubKeyBytes[8:])
	return rsa.PublicKey{
		N: N,
		E: int(E),
	}, nil
}

type rsaDecryptor struct {
	*rsa.PrivateKey
}

func NewRSADecryptor(key interface{}) (Decryptor, error) {
	var privKey *rsa.PrivateKey
	switch key := key.(type) {
	case *rsa.PrivateKey:
		privKey = key
	case rsa.PrivateKey:
		privKey = &key
	case []byte:
		privateKey, err := unmarshalRSAPrivKey(key)
		if err != nil {
			return nil, err
		}
		privKey = &privateKey
	}
	return &rsaDecryptor{
		PrivateKey: privKey,
	}, nil
}

func (decryptor *rsaDecryptor) Marshal() ([]byte, error) {
	return marshalRSAPrivKey(decryptor.PrivateKey)
}

func (decryptor *rsaDecryptor) Decrypt(data []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, decryptor.PrivateKey, data, nil)
}

func (decryptor *rsaDecryptor) Encryptor() Encryptor {
	return &rsaEncryptor{
		PublicKey: &decryptor.PublicKey,
	}
}

func marshalRSAPrivKey(privateKey *rsa.PrivateKey) ([]byte, error) {
	buf := new(bytes.Buffer)
	primeCount := int64(len(privateKey.Primes))
	if err := binary.Write(buf, binary.BigEndian, primeCount); err != nil {
		return nil, err
	}
	for _, prime := range privateKey.Primes {
		primeBytes := prime.Bytes()
		primeSize := int64(len(primeBytes))
		if err := binary.Write(buf, binary.BigEndian, primeSize); err != nil {
			return []byte{}, err
		}
		if err := binary.Write(buf, binary.BigEndian, primeBytes); err != nil {
			return []byte{}, err
		}
	}
	dBytes := privateKey.D.Bytes()
	dSize := int64(len(dBytes))
	if err := binary.Write(buf, binary.BigEndian, dSize); err != nil {
		return []byte{}, err
	}
	if err := binary.Write(buf, binary.BigEndian, dBytes); err != nil {
		return []byte{}, err
	}
	pubKeyBytes, err := marshalRSAPubKey(&privateKey.PublicKey)
	if err != nil {
		return []byte{}, err
	}
	if err := binary.Write(buf, binary.BigEndian, pubKeyBytes); err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

func unmarshalRSAPrivKey(privKeyBytes []byte) (rsa.PrivateKey, error) {
	var primeCount int64
	if err := binary.Read(bytes.NewReader(privKeyBytes[:8]), binary.BigEndian, &primeCount); err != nil {
		return rsa.PrivateKey{}, err
	}
	privKeyBytes = privKeyBytes[8:]
	primes := []*big.Int{}
	for i := int64(0); i < primeCount; i++ {
		var primeSize int64
		if err := binary.Read(bytes.NewReader(privKeyBytes[:8]), binary.BigEndian, &primeSize); err != nil {
			return rsa.PrivateKey{}, err
		}
		prime := new(big.Int).SetBytes(privKeyBytes[8 : primeSize+8])
		primes = append(primes, prime)
		privKeyBytes = privKeyBytes[primeSize+8:]
	}
	var dSize int64
	if err := binary.Read(bytes.NewReader(privKeyBytes[:8]), binary.BigEndian, &dSize); err != nil {
		return rsa.PrivateKey{}, err
	}
	D := new(big.Int).SetBytes(privKeyBytes[8 : dSize+8])
	pubKey, err := unmarshalRSAPubKey(privKeyBytes[dSize+8:])
	if err != nil {
		return rsa.PrivateKey{}, err
	}
	return rsa.PrivateKey{
		PublicKey: pubKey,
		D:         D,
		Primes:    primes,
	}, nil
}
