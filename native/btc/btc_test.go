package btc_test

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"

	// "github.com/renproject/renxcs-go"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/renproject/libbtc-go"
	"github.com/renproject/libbtc-go/clients"
	. "github.com/renproject/renxcs-go/native/btc"
	"github.com/sirupsen/logrus"
	"github.com/tyler-smith/go-bip39"
)

var _ = Describe("Bitcoin Native Binder", func() {

	loadMasterKey := func(network uint32) (*hdkeychain.ExtendedKey, error) {
		switch network {
		case 1:
			seed := bip39.NewSeed(os.Getenv("BITCOIN_TESTNET_MNEMONIC"), os.Getenv("BITCOIN_TESTNET_PASSPHRASE"))
			return hdkeychain.NewMaster(seed, &chaincfg.TestNet3Params)
		case 0:
			seed := bip39.NewSeed(os.Getenv("BITCOIN_MNEMONIC"), os.Getenv("BITCOIN_PASSPHRASE"))
			return hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
		default:
			return nil, fmt.Errorf("network id: %d", network)
		}
	}

	loadKey := func(path ...uint32) (*ecdsa.PrivateKey, error) {
		key, err := loadMasterKey(path[1])
		if err != nil {
			return nil, err
		}
		for _, val := range path {
			key, err = key.Child(val)
			if err != nil {
				return nil, err
			}
		}
		privKey, err := key.ECPrivKey()
		if err != nil {
			return nil, err
		}
		return privKey.ToECDSA(), nil
	}

	Context("when honestly locking and unlocking btc", func() {
		It("should be able to lock btc", func() {
			renKey, err := loadKey(44, 1, 1, 0, 0)
			privKey, err := loadKey(44, 1, 0, 0, 0)
			Expect(err).Should(BeNil())
			client, err := clients.NewBitcoinFNClient(os.Getenv("RPC_URL"), os.Getenv("RPC_USER"), os.Getenv("RPC_PASSWORD"))
			Expect(err).Should(BeNil())
			acc := libbtc.NewAccount(client, privKey, logrus.StandardLogger())
			binder := NewBitcoinBinder(acc)
			pubKey, err := client.SerializePublicKey((*btcec.PublicKey)(&renKey.PublicKey))
			Expect(err).Should(BeNil())
			addr, err := client.PublicKeyToAddress(pubKey)
			Expect(err).Should(BeNil())
			Expect(binder.Lock(addr.EncodeAddress(), big.NewInt(50000))).Should(BeNil())
		})

		It("should be able to unlock btc", func() {
			renKey, err := loadKey(44, 1, 1, 0, 0)
			privKey, err := loadKey(44, 1, 0, 0, 0)
			Expect(err).Should(BeNil())
			client, err := clients.NewBitcoinFNClient(os.Getenv("RPC_URL"), os.Getenv("RPC_USER"), os.Getenv("RPC_PASSWORD"))
			Expect(err).Should(BeNil())
			pubKeyBytes, err := client.SerializePublicKey((*btcec.PublicKey)(&privKey.PublicKey))
			Expect(err).Should(BeNil())
			addr, err := client.PublicKeyToAddress(pubKeyBytes)
			Expect(err).Should(BeNil())
			acc := libbtc.NewAccount(client, renKey, logrus.StandardLogger())
			binder := NewBitcoinBinder(acc)
			Expect(binder.Unlock(addr.EncodeAddress())).Should(BeNil())
		})
	})
})
