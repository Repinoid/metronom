package privacy

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
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
	pub := cert.PublicKey.(*rsa.PublicKey)

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
	//pK := priKey.(*rsa.PrivateKey)
	decryptOAEP, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, priKey, cipherByte, nil)
	if err != nil {
		panic(err)
	}
	plainText = decryptOAEP
	return
}

// https://www.programmerall.com/article/96002339528/
