
package main

import (
	"fmt"
)

// EncryptedChunk represents one encrypted piece of a file.
// Data contains the nonce and authenticated ciphertext.
type EncryptedChunk struct {
	Index int
	Data  []byte
}

// ProcessFileChunks reads a file in chunks, encrypts each chunk using
// the supplied per-file key, and sends it to the provided callback.
func ProcessFileChunks(
	path string,
	fileKey []byte,
	process func(EncryptedChunk) error,
) error {
	if len(fileKey) != EncryptionKeySize {
		return fmt.Errorf("file key must be %d bytes", EncryptionKeySize)
	}
	if process == nil {
		return fmt.Errorf("encrypted chunk processor is required")
	}

	return SplitFile(path, func(index int, plaintext []byte) error {
		encrypted, err := Encrypt(fileKey, plaintext)
		if err != nil {
			return fmt.Errorf("encrypt chunk %d: %w", index, err)
		}

		chunk := EncryptedChunk{
			Index: index,
			Data:  encrypted,
		}

		if err := process(chunk); err != nil {
			return fmt.Errorf("process encrypted chunk %d: %w", index, err)
		}
		return nil
	})
}
