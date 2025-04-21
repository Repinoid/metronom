package privacy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
)

// Use the data of the public key of the other party, only the other person's private key can be unbearable
func Encrypt(msg []byte, publicKey string) (cipherByte []byte, err error) {
	//	msg := []byte(plain)
	// Decode public key
	pubBlock, _ := pem.Decode([]byte(publicKey))
	// read the public key
	pubKeyValue, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		panic(err)
	}
	pub := pubKeyValue.(*rsa.PublicKey)
	// Encrypted data method: Do not use EncryptPKCS1V15 method to encrypt, INCRYPTOAEP is recommended in the source code, so use security method to encrypt
	encryptOAEP, err := rsa.EncryptOAEP(sha1.New(), rand.Reader, pub, msg, nil)
	if err != nil {
		panic(err)
	}
	cipherByte = encryptOAEP
	return
}

// Decrypt the key encrypted by the private key
func Decrypt(cipherByte []byte, privateKey []byte) (plainText []byte, err error) {
	// Analyze the private key
	priBlock, _ := pem.Decode(privateKey)
	priKey, err := x509.ParsePKCS1PrivateKey(priBlock.Bytes)
	if err != nil {
		panic(err)
	}
	// Decrypt the contents of the RSA-OAEP mode encrypted
	decryptOAEP, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, priKey, cipherByte, nil)
	if err != nil {
		panic(err)
	}
	plainText = decryptOAEP
	return
}

// https://www.programmerall.com/article/96002339528/
