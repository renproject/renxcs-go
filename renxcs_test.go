package renxcs_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/renproject/libbtc-go"
	"github.com/renproject/libbtc-go/clients"
	"github.com/renproject/libeth-go"
	. "github.com/renproject/renxcs-go"
	renCrypto "github.com/renproject/renxcs-go/crypto"
	"github.com/renproject/renxcs-go/foreign/eth/wbtc"
	"github.com/renproject/renxcs-go/native/btc"
	"github.com/sirupsen/logrus"
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

	ethKeys := map[string]*ecdsa.PrivateKey{}
	btcKeys := map[string]*ecdsa.PrivateKey{}
	var wbtcAddress common.Address

	claimEthPrivateKey := func(binder Bridge, pubKey rsa.PublicKey, btcAddress string) ([]byte, error) {
		if err := binder.Verify(pubKey, common.HexToAddress(btcAddress)); err != nil {
			return nil, err
		}
		addr, err := binder.OwnerOf(common.HexToAddress(btcAddress))
		if err != nil {
			return nil, err
		}
		encrypter, err := renCrypto.NewRSAEncrypter(pubKey)
		if err != nil {
			return nil, err
		}
		signer, err := renCrypto.NewECDSASigner(ethKeys[addr.String()])
		if err != nil {
			return nil, err
		}
		privKeyBytes, err := signer.Marshal()
		if err != nil {
			return nil, err
		}
		return encrypter.Encrypt(privKeyBytes)
	}

	claimBtcPrivateKey := func(binder Bridge, pubKey rsa.PublicKey, btcAddress common.Address) ([]byte, error) {
		if err := binder.Verify(pubKey, btcAddress); err != nil {
			return nil, err
		}
		encrypter, err := renCrypto.NewRSAEncrypter(pubKey)
		if err != nil {
			return nil, err
		}

		privKey, ok := btcKeys[hex.EncodeToString(btcAddress.Bytes())]
		if !ok {
			return nil, fmt.Errorf("cannot find the private key corresponding to: %s", btcAddress)
		}

		signer, err := renCrypto.NewECDSASigner(privKey)
		if err != nil {
			return nil, err
		}
		privKeyBytes, err := signer.Marshal()
		if err != nil {
			return nil, err
		}
		fmt.Println("Priv Key", hex.EncodeToString(privKeyBytes))
		return encrypter.Encrypt(privKeyBytes)
	}

	// burnTokenReq := func(binder Bridge, pubKey rsa.PublicKey, address string) ([]byte, error) {
	// 	if err := binder.Verify(pubKey, address); err != nil {
	// 		return nil, err
	// 	}

	// }

	BeforeSuite(func() {
		ethAddrs := []common.Address{}
		btcAddrs := []common.Address{}

		for i := uint32(0); i < 10; i++ {
			btcKey, err := loadKey(44, 1, 1, 0, i)
			Expect(err).Should(BeNil())
			ethKey, err := loadKey(44, 1, 2, 0, i)
			Expect(err).Should(BeNil())
			cl, err := clients.NewBlockchainInfoClient("testnet")
			Expect(err).Should(BeNil())
			ethAddress := crypto.PubkeyToAddress(ethKey.PublicKey)
			pubKey, err := cl.SerializePublicKey((*btcec.PublicKey)(&btcKey.PublicKey))
			Expect(err).Should(BeNil())
			addr, err := cl.PublicKeyToAddress(pubKey)
			Expect(err).Should(BeNil())
			pubKeyHash := addr.(*btcutil.AddressPubKeyHash)
			addrString := hex.EncodeToString(pubKeyHash.Hash160()[:])
			ethKeys[ethAddress.String()] = ethKey
			btcKeys[addrString] = btcKey
			ethAddrs = append(ethAddrs, ethAddress)
			btcAddrs = append(btcAddrs, common.HexToAddress(addrString))
		}

		key, err := loadKey(44, 1, 0, 0, 0)
		Expect(err).Should(BeNil())
		acc, err := libeth.NewAccount("https://kovan.infura.io", key)
		Expect(err).Should(BeNil())

		address, err := wbtc.Deploy(acc, ethAddrs, btcAddrs)
		Expect(err).Should(BeNil())
		wbtcAddress = address
		fmt.Println(address.String())
	})

	Context("when interacting honestly", func() {
		var btcAddress common.Address
		var rsakey *rsa.PrivateKey

		It("should successfuly claim the nft", func() {
			privKey, err := loadKey(44, 1, 0, 0, 0)
			Expect(err).Should(BeNil())
			acc, err := libeth.NewAccount("https://kovan.infura.io", privKey)
			Expect(err).Should(BeNil())
			acc.WriteAddress("REN.WBTC", wbtcAddress)
			binder, err := wbtc.NewWBTCBinder(acc)
			Expect(err).Should(BeNil())
			btcAddress, err = binder.UnusedAddress()
			Expect(err).Should(BeNil())
			rsakey, err = rsa.GenerateKey(rand.Reader, 2048)
			Expect(err).Should(BeNil())
			Expect(binder.Claim(rsakey.PublicKey, btcAddress)).Should(BeNil())
		})

		It("should successfully fund a btc address", func() {
			privKey, err := loadKey(44, 1, 0, 0, 0)
			Expect(err).Should(BeNil())
			client, err := clients.NewBitcoinFNClient(os.Getenv("RPC_URL"), os.Getenv("RPC_USER"), os.Getenv("RPC_PASSWORD"))
			Expect(err).Should(BeNil())
			acc := libbtc.NewAccount(client, privKey, logrus.StandardLogger())
			binder := btc.NewBitcoinBinder(acc)
			addr, err := btcutil.NewAddressPubKeyHash(btcAddress.Bytes(), acc.NetworkParams())
			Expect(binder.Lock(addr.EncodeAddress(), big.NewInt(50000))).Should(BeNil())
		})

		It("should successfully burn the nft", func() {
			privKey, err := loadKey(44, 1, 0, 0, 0)
			Expect(err).Should(BeNil())
			acc, err := libeth.NewAccount("https://kovan.infura.io", privKey)
			Expect(err).Should(BeNil())
			acc.WriteAddress("REN.WBTC", wbtcAddress)
			binder, err := wbtc.NewWBTCBinder(acc)
			Expect(err).Should(BeNil())
			encryptedPrivKey, err := claimEthPrivateKey(binder, rsakey.PublicKey, btcAddress.String())
			Expect(err).Should(BeNil())
			decryptor, err := renCrypto.NewRSADecrypter(rsakey)
			Expect(err).Should(BeNil())
			privKeyBytes, err := decryptor.Decrypt(encryptedPrivKey)
			Expect(err).Should(BeNil())
			signer, err := renCrypto.NewECDSASigner(privKeyBytes)
			Expect(err).Should(BeNil())
			tokenOwner, err := libeth.NewAccount("https://kovan.infura.io", signer.PrivKey().(*ecdsa.PrivateKey))
			Expect(err).Should(BeNil())
			tokenOwnerBinder, err := wbtc.NewWBTCBinder(tokenOwner)
			Expect(err).Should(BeNil())
			_, err = acc.Transfer(context.Background(), tokenOwner.Address(), big.NewInt(1000000000000000), libeth.Fast, 0, false)
			Expect(err).Should(BeNil())
			Expect(tokenOwnerBinder.Burn(rsakey.PublicKey, btcAddress)).Should(BeNil())
		})

		It("should successfully spend a btc address", func() {
			privKey, err := loadKey(44, 1, 0, 0, 0)
			Expect(err).Should(BeNil())
			acc, err := libeth.NewAccount("https://kovan.infura.io", privKey)
			Expect(err).Should(BeNil())
			acc.WriteAddress("REN.WBTC", wbtcAddress)
			binder, err := wbtc.NewWBTCBinder(acc)
			Expect(err).Should(BeNil())
			encryptedPrivKey, err := claimBtcPrivateKey(binder, rsakey.PublicKey, btcAddress)
			Expect(err).Should(BeNil())
			decryptor, err := renCrypto.NewRSADecrypter(rsakey)
			Expect(err).Should(BeNil())
			privKeyBytes, err := decryptor.Decrypt(encryptedPrivKey)
			Expect(err).Should(BeNil())
			signer, err := renCrypto.NewECDSASigner(privKeyBytes)
			Expect(err).Should(BeNil())
			client, err := clients.NewBitcoinFNClient(os.Getenv("RPC_URL"), os.Getenv("RPC_USER"), os.Getenv("RPC_PASSWORD"))
			Expect(err).Should(BeNil())
			pubKeyBytes, err := client.SerializePublicKey((*btcec.PublicKey)(&privKey.PublicKey))
			Expect(err).Should(BeNil())
			addr, err := client.PublicKeyToAddress(pubKeyBytes)
			Expect(err).Should(BeNil())
			btcAcc := libbtc.NewAccount(client, signer.PrivKey().(*ecdsa.PrivateKey), logrus.StandardLogger())
			btcBinder := btc.NewBitcoinBinder(btcAcc)
			Expect(btcBinder.Unlock(addr.EncodeAddress())).Should(BeNil())
		})
	})
})
