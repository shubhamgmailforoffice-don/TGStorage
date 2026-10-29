
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

const (
	// AES-256 requires a 32-byte key.
	EncryptionKeySize = 32

	// GCM uses a 12-byte nonce by default.
	NonceSize = 12
)

// GenerateKey creates a cryptographically random AES-256 key.
func GenerateKey() ([]byte, error) {
	key := make([]byte, EncryptionKeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate encryption key: %w", err)
	}
	return key, nil
}

// Encrypt encrypts plaintext and returns nonce + authenticated ciphertext.
// Store the returned bytes, not the plaintext.
func Encrypt(key, plaintext []byte) ([]byte, error) {
	if len(key) != EncryptionKeySize {
		return nil, errors.New("encryption key must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	// Seal appends the ciphertext and authentication tag to nonce.
	sealed := gcm.Seal(nil, nonce, plaintext, nil)
	return append(nonce, sealed...), nil
}

// Decrypt authenticates and decrypts data produced by Encrypt.
func Decrypt(key, encrypted []byte) ([]byte, error) {
	if len(key) != EncryptionKeySize {
		return nil, errors.New("encryption key must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	if len(encrypted) < gcm.NonceSize()+gcm.Overhead() {
		return nil, errors.New("encrypted data is too short or corrupted")
	}

	nonce := encrypted[:gcm.NonceSize()]
	ciphertext := encrypted[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed: invalid key or corrupted data")
	}

	return plaintext, nil
}
