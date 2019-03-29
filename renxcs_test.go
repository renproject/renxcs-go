package renxcs_test

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/renproject/libbtc-go"
	"github.com/renproject/libeth-go"
	. "github.com/renproject/renxcs-go"
	"github.com/renproject/renxcs-go/foreign/eth/zbtc"
	"github.com/renproject/renxcs-go/native/btc"
	"github.com/tyler-smith/go-bip39"
)

var _ = Describe("Bitcoin Native Binder", func() {

	loadMasterKey := func(network uint32) (*hdkeychain.ExtendedKey, error) {
		switch network {
		case 1:
			seed := bip39.NewSeed(os.Getenv("ETHEREUM_TESTNET_MNEMONIC"), os.Getenv("ETHEREUM_TESTNET_PASSPHRASE"))
			return hdkeychain.NewMaster(seed, &chaincfg.TestNet3Params)
		case 0:
			seed := bip39.NewSeed(os.Getenv("ETHEREUM_MNEMONIC"), os.Getenv("ETHEREUM_PASSPHRASE"))
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

	var renBTCAddress string
	var zbtcAddress common.Address

	renvmSign := func(txHash [32]byte, amount *big.Int) []byte {
		key, err := loadKey(44, 1, 0, 0, 0)
		if err != nil {
			panic(err)
		}
		sig, err := crypto.Sign(txHash[:], key)
		if err != nil {
			panic(err)
		}
		return sig
	}

	initInterop := func() (common.Address, string) {
		cl, err := libbtc.NewBlockchainInfoClient("testnet")
		Expect(err).Should(BeNil())
		key, err := loadKey(44, 1, 0, 0, 0)
		Expect(err).Should(BeNil())
		client, err := libeth.NewMercuryClient("kovan", "")
		Expect(err).Should(BeNil())
		acc, err := libeth.NewAccount(client, key)
		Expect(err).Should(BeNil())
		renETHAddress := acc.Address()
		address, _, err := zbtc.Deploy(acc, renETHAddress)
		Expect(err).Should(BeNil())
		zbtcAddress := address
		btcAcc := libbtc.NewAccount(cl, key, nil)
		btcAddr, err := btcAcc.Address()
		Expect(err).Should(BeNil())
		renBTCAddress := btcAddr
		return zbtcAddress, renBTCAddress.String()
	}

	claimToken := func(native NativeBinder, foreign ForeignBinder) [32]byte {
		amount := big.NewInt(20000)
		txHash, tx, err := native.Build(renBTCAddress, amount)
		if err != nil {
			panic(err)
		}
		txHashBytes, err := hex.DecodeString(txHash)
		if err != nil {
			panic(err)
		}
		txHashBytes32 := [32]byte{}
		copy(txHashBytes32[:], txHashBytes)
		if err := foreign.Claim(txHashBytes32, amount); err != nil {
			panic(err)
		}
		if err := native.Submit(tx); err != nil {
			panic(err)
		}
		return txHashBytes32
	}

	mintToken := func(foreign ForeignBinder, txHashBytes32 [32]byte, sig []byte) {
		if err := foreign.Mint(txHashBytes32, sig); err != nil {
			panic(err)
		}
	}

	BeforeSuite(func() {
		zbtcAddr, btcAddr := initInterop()
		zbtcAddress = zbtcAddr
		renBTCAddress = btcAddr
	})

	Context("when interacting honestly", func() {
		It("should successfully spend a btc address", func() {
			key, err := loadKey(44, 1, 2, 0, 0)
			Expect(err).Should(BeNil())
			cl, err := libbtc.NewBlockchainInfoClient("testnet")
			Expect(err).Should(BeNil())
			client, err := libeth.NewMercuryClient("kovan", "")
			Expect(err).Should(BeNil())
			acc, err := libeth.NewAccount(client, key)
			Expect(err).Should(BeNil())
			btcAcc := libbtc.NewAccount(cl, key, nil)
			native := btc.NewBitcoinBinder(btcAcc)
			foreign, err := zbtc.Connect(acc, zbtcAddress)
			Expect(err).Should(BeNil())
			txHash32 := claimToken(native, foreign)
			sig := renvmSign(txHash32, big.NewInt(20000))
			mintToken(foreign, txHash32, sig)
		})
	})
})
