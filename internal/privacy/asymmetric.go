package privacy

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

// Use the data of the public key of the other party, only the other person's private key can be unbearable
func Encrypt(msg []byte, publicKey []byte) (cipherByte []byte, err error) {
	pubBlock, _ := pem.Decode(publicKey)
	cert, err := x509.ParseCertificate(pubBlock.Bytes)
	if err != nil {
		panic(err)
	}
	pubK := cert.PublicKey.(*rsa.PublicKey)

	cipherByte, err = HybridEncrypt(pubK, msg)

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

	plainText, err = HybridDecrypt(priKey, cipherByte)

	return
}

// https://www.programmerall.com/article/96002339528/
