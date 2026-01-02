package crypto

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	// Generate a random 32-byte key
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	plainText := []byte("This is a secret message 123!")

	// Encrypt
	encrypted, err := Encrypt(plainText, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Decrypt
	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	// Verify
	if !bytes.Equal(plainText, decrypted) {
		t.Errorf("Decrypted message does not match original. Got %q, want %q", decrypted, plainText)
	}
}

func TestInvalidKeyLength(t *testing.T) {
	key := make([]byte, 16) // Only 16 bytes, need 32 for AES-256
	plainText := []byte("fail")

	_, err := Encrypt(plainText, key)
	if err == nil {
		t.Error("Encrypt should fail with invalid key length")
	}

	_, err = Decrypt([]byte("stuff"), key)
	if err == nil {
		t.Error("Decrypt should fail with invalid key length")
	}
}

func TestTamperedCiphertext(t *testing.T) {
	key := make([]byte, 32)
	io.ReadFull(rand.Reader, key)

	plainText := []byte("secret data")
	encrypted, _ := Encrypt(plainText, key)

	// Tamper with the last byte of the ciphertext
	encrypted[len(encrypted)-1] ^= 0xFF

	_, err := Decrypt(encrypted, key)
	if err == nil {
		t.Error("Decrypt should ensure integrity and fail on tampered data")
	}
}
