package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// Encrypt encrypts plainText using the 32-byte key via AES-GCM.
// It prepends the generated nonce to the ciphertext.
func Encrypt(plainText []byte, key []byte) ([]byte, error) {
	// AES-256 requires a 32-byte key.
	if len(key) != 32 {
		return nil, errors.New("key must be exactly 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// GCM (Galois/Counter Mode) is an authenticated encryption mode.
	// It provides both confidentiality and integrity.
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Never reuse a nonce with the same key.
	// 12 bytes is the standard nonce size for GCM.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Seal appends the encrypted data to the first argument (nonce),
	// so the final result is [Nonce | CipherText | Tag].
	return gcm.Seal(nonce, nonce, plainText, nil), nil
}

// Decrypt decrypts data encrypted by Encrypt.
// It expects the logic: [Nonce | CipherText | Tag].
func Decrypt(data []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be exactly 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("malformed ciphertext")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Open checks the authenticity (Tag) and decrypts the ciphertext.
	return gcm.Open(nil, nonce, ciphertext, nil)
}
