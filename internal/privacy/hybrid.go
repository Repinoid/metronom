package privacy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
)

func HybridEncrypt(pub *rsa.PublicKey, data []byte) ([]byte, error) {
	// Generate a random symmetric key
	symKey := make([]byte, 32) // 256-bit AES key
	if _, err := rand.Read(symKey); err != nil {
		return nil, err
	}

	// Encrypt the symmetric key with RSA
	encKey, err := rsa.EncryptPKCS1v15(rand.Reader, pub, symKey)
	if err != nil {
		return nil, err
	}

	// Encrypt the data with AES
	block, err := aes.NewCipher(symKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, data, nil)

	// Combine the encrypted key, nonce, and ciphertext
	result := append(encKey, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

func HybridDecrypt(priv *rsa.PrivateKey, data []byte) ([]byte, error) {
	keySize := priv.PublicKey.Size()
	if len(data) < keySize {
		return nil, fmt.Errorf("invalid ciphertext length")
	}

	// Split into encrypted key and ciphertext
	encKey := data[:keySize]
	ciphertext := data[keySize:]

	// Decrypt the symmetric key
	symKey, err := rsa.DecryptPKCS1v15(rand.Reader, priv, encKey)
	if err != nil {
		return nil, err
	}

	// Decrypt the data with AES
	block, err := aes.NewCipher(symKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("invalid ciphertext length")
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	return gcm.Open(nil, nonce, ciphertext, nil)
}
