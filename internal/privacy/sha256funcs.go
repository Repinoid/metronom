package privacy

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func EncryptB2B(bytesToEncrypt, key []byte) (encrypted []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce, _ := RandBytes(aesGCM.NonceSize())
	ciphertext := aesGCM.Seal(nonce, nonce, bytesToEncrypt, nil)
	return ciphertext, nil
}
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
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
func GenerateByteKey() (byteKey []byte, err error) {
	rb, err := RandBytes(32)
	byteKey = make([]byte, len(rb)*2)
	n := hex.Encode(byteKey, rb)
	return byteKey[:n], err
}
func RandBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
func MakeHash(prior, data, keyB []byte) []byte {
	h := hmac.New(sha256.New, keyB) // New returns a new HMAC hash using the given hash.Hash type and key.
	h.Write(data)                   // func (hash.Hash) Sum(b []byte) []byte
	dst := h.Sum(prior)             //Sum appends the current hash to b and returns the resulting slice. It does not change the underlying hash state.
	return dst

}
