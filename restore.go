
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// RestoreEncryptedChunks decrypts encrypted chunks in index order
// and writes the original file to outputPath.
func RestoreEncryptedChunks(
	chunks []EncryptedChunk,
	fileKey []byte,
	outputPath string,
) error {
	if len(fileKey) != EncryptionKeySize {
		return fmt.Errorf("file key must be %d bytes", EncryptionKeySize)
	}

	ordered := append([]EncryptedChunk(nil), chunks...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Index < ordered[j].Index
	})

	for i, chunk := range ordered {
		if chunk.Index != i {
			return fmt.Errorf("missing or duplicate chunk index at %d", i)
		}
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0700); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	output, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}

	for _, chunk := range ordered {
		plaintext, err := Decrypt(fileKey, chunk.Data)
		if err != nil {
			output.Close()
			os.Remove(outputPath)
			return fmt.Errorf("decrypt chunk %d: %w", chunk.Index, err)
		}

		if _, err := output.Write(plaintext); err != nil {
			output.Close()
			os.Remove(outputPath)
			return fmt.Errorf("write restored chunk %d: %w", chunk.Index, err)
		}
	}

	if err := output.Close(); err != nil {
		os.Remove(outputPath)
		return fmt.Errorf("close restored file: %w", err)
	}

	return nil
}
