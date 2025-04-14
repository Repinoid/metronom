// пакет криптографии
package privacy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
)

// Шифруем байты в байты.
func EncryptB2B(bytesToEncrypt, key []byte) (encrypted []byte, err error) {

	// Базовый интерфейс симметричного шифрования — cipher.Block из пакета  https://pkg.go.dev/crypto/cipher
	// Зашифруем помощью алгоритма AES (Advanced Encryption Standard). Это блочный алгоритм, размер блока — 16 байт.
	//
	// NewCipher создает и возвращает новый cipher.Block.
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// NewGCM returns the given 128-bit, block cipher wrapped in Galois Counter Mode with the standard nonce length.
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	// создаём вектор инициализации
	nonce, _ := RandBytes(aesGCM.NonceSize())
	// зашифровываем
	ciphertext := aesGCM.Seal(nonce, nonce, bytesToEncrypt, nil)
	return ciphertext, nil
}

// Расшифровываем байты в байты.
func DecryptB2B(encrypted, key []byte) (decrypted []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	// расшифровываем
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func RandBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// подписываем алгоритмом HMAC, используя SHA-256
func MakeHash(prior, data, keyB []byte) []byte {
	h := hmac.New(sha256.New, keyB) // New returns a new HMAC hash using the given hash.Hash type and key.
	h.Write(data)                   // func (hash.Hash) Sum(b []byte) []byte
	dst := h.Sum(prior)             //Sum appends the current hash to b and returns the resulting slice. It does not change the underlying hash state.
	return dst

}
