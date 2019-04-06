package zbtc_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/renproject/libeth-go"
	. "github.com/renproject/renxcs-go/foreign/eth/zbtc"
	"github.com/tyler-smith/go-bip39"
)

var _ = Describe("ZBTC Foreign Binder", func() {

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
		It("should be able to mint zbtc", func() {
			client, err := libeth.NewMercuryClient("kovan", "")
			Expect(err).Should(BeNil())
			privKey, err := loadKey(44, 1, 1, 0, 0)
			Expect(err).Should(BeNil())
			account, err := libeth.NewAccount(client, privKey)
			Expect(err).Should(BeNil())
			addr, zbtc, err := Deploy(account, common.HexToAddress("0x7e93d041fbf29a66760740274cf6107424b5b1e9"))
			Expect(err).Should(BeNil())
			fmt.Println("Address: ", addr.String())
			renKey, err := loadKey(44, 1, 0, 0, 0)
			Expect(err).Should(BeNil())
			hash32 := [32]byte{}
			rand.Read(hash32[:])
			hash, err := zbtc.Commitment(account.Address(), big.NewInt(20000), hash32)
			Expect(err).Should(BeNil())
			sig, err := crypto.Sign(hash[:], renKey)
			r := new(big.Int).SetBytes(sig[:32])
			s := new(big.Int).SetBytes(sig[32:64])
			Expect(err).Should(BeNil())
			err = zbtc.Mint(big.NewInt(20000), hash32, r, s)
			Expect(err).Should(BeNil())
		})

		// It("should be able to unlock btc", func() {
		// 	renKey, err := loadKey(44, 1, 1, 0, 0)
		// 	privKey, err := loadKey(44, 1, 0, 0, 0)
		// 	Expect(err).Should(BeNil())
		// 	client, err := clients.NewBitcoinFNClient(os.Getenv("RPC_URL"), os.Getenv("RPC_USER"), os.Getenv("RPC_PASSWORD"))
		// 	Expect(err).Should(BeNil())
		// 	pubKeyBytes, err := client.SerializePublicKey((*btcec.PublicKey)(&privKey.PublicKey))
		// 	Expect(err).Should(BeNil())
		// 	addr, err := client.PublicKeyToAddress(pubKeyBytes)
		// 	Expect(err).Should(BeNil())
		// 	acc := libbtc.NewAccount(client, renKey, logrus.StandardLogger())
		// 	binder := NewBitcoinBinder(acc)
		// 	Expect(binder.Unlock(addr.EncodeAddress())).Should(BeNil())
		// })
	})
})
