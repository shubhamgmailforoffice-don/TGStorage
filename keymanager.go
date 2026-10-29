
package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

// WrapFileKey encrypts a file's data-encryption key using the master key.
// The returned bytes contain the nonce followed by authenticated ciphertext.
func WrapFileKey(masterKey, fileKey []byte) ([]byte, error) {
	if len(masterKey) != EncryptionKeySize {
		return nil, errors.New("master key must be 32 bytes")
	}
	if len(fileKey) != EncryptionKeySize {
		return nil, errors.New("file key must be 32 bytes")
	}

	return Encrypt(masterKey, fileKey)
}

// UnwrapFileKey recovers a file's data-encryption key.
func UnwrapFileKey(masterKey, wrappedKey []byte) ([]byte, error) {
	if len(masterKey) != EncryptionKeySize {
		return nil, errors.New("master key must be 32 bytes")
	}

	fileKey, err := Decrypt(masterKey, wrappedKey)
	if err != nil {
		return nil, fmt.Errorf("unwrap file key: %w", err)
	}
	if len(fileKey) != EncryptionKeySize {
		return nil, errors.New("unwrapped file key has invalid length")
	}

	return fileKey, nil
}

// GenerateMasterKey creates a random 32-byte master key.
// Store it securely outside the repository; do not print it.
func GenerateMasterKey() ([]byte, error) {
	key := make([]byte, EncryptionKeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("generate master key: %w", err)
	}
	return key, nil
}
