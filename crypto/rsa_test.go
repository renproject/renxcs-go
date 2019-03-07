package crypto_test

import (
	"crypto/rand"
	"crypto/rsa"
	"reflect"
	"testing/quick"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/renproject/renxcs-go/crypto"
)

var _ = Describe("RSA", func() {
	Context("when marshalling and unmarshalling the encryptor", func() {
		It("should not change", func() {
			Expect(quick.Check(func() bool {
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				Expect(err).Should(BeNil())
				encryptor, err := NewRSAEncryptor(key)
				Expect(err).Should(BeNil())
				data, err := encryptor.Marshal()
				Expect(err).Should(BeNil())
				encryptor2, err := NewRSAEncryptor(data)
				Expect(err).Should(BeNil())
				data2, err := encryptor2.Marshal()
				Expect(err).Should(BeNil())
				return reflect.DeepEqual(data, data2)
			}, &quick.Config{
				MaxCount: 8,
			})).Should(BeNil())
		})

		It("should be able to encrypt a message", func() {
			Expect(quick.Check(func() bool {
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				Expect(err).Should(BeNil())
				encryptor, err := NewRSAEncryptor(key)
				Expect(err).Should(BeNil())
				data, err := encryptor.Marshal()
				Expect(err).Should(BeNil())
				_, err = encryptor.Encrypt([]byte("Secret"))
				Expect(err).Should(BeNil())
				encryptor2, err := NewRSAEncryptor(data)
				Expect(err).Should(BeNil())
				_, err = encryptor2.Encrypt([]byte("Secret"))
				Expect(err).Should(BeNil())
				return true
			}, &quick.Config{
				MaxCount: 8,
			})).Should(BeNil())
		})
	})

	Context("when marshalling and unmarshalling the decryptor", func() {
		It("should not change", func() {
			Expect(quick.Check(func() bool {
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				Expect(err).Should(BeNil())
				decryptor, err := NewRSADecryptor(key)
				Expect(err).Should(BeNil())
				data, err := decryptor.Marshal()
				Expect(err).Should(BeNil())
				decryptor2, err := NewRSADecryptor(data)
				Expect(err).Should(BeNil())
				data2, err := decryptor2.Marshal()
				Expect(err).Should(BeNil())
				return reflect.DeepEqual(data, data2)
			}, &quick.Config{
				MaxCount: 8,
			})).Should(BeNil())
		})

		It("should be able to decrypt a message", func() {
			Expect(quick.Check(func() bool {
				key, err := rsa.GenerateKey(rand.Reader, 2048)
				Expect(err).Should(BeNil())
				decryptor, err := NewRSADecryptor(key)
				Expect(err).Should(BeNil())
				cipherText, err := decryptor.Encryptor().Encrypt([]byte("Secret"))
				Expect(err).Should(BeNil())
				data, err := decryptor.Marshal()
				Expect(err).Should(BeNil())
				data1, err := decryptor.Decrypt(cipherText)
				Expect(err).Should(BeNil())
				decryptor2, err := NewRSADecryptor(data)
				Expect(err).Should(BeNil())
				data2, err := decryptor2.Decrypt(cipherText)
				Expect(err).Should(BeNil())
				return reflect.DeepEqual(data1, data2)
			}, &quick.Config{
				MaxCount: 8,
			})).Should(BeNil())
		})
	})
})
