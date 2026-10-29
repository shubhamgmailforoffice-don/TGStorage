
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreEncryptedChunks(t *testing.T) {
	original := []byte("TGStorage restore test: original file content.")

	fileKey, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	var chunks []EncryptedChunk
	err = ProcessTestData(original, fileKey, func(chunk EncryptedChunk) error {
		chunks = append(chunks, chunk)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Reverse the chunks to verify restoration sorts them by index.
	for i, j := 0, len(chunks)-1; i < j; i, j = i+1, j-1 {
		chunks[i], chunks[j] = chunks[j], chunks[i]
	}

	outputPath := filepath.Join(t.TempDir(), "restored.bin")
	if err := RestoreEncryptedChunks(chunks, fileKey, outputPath); err != nil {
		t.Fatal(err)
	}

	restored, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(restored, original) {
		t.Fatal("restored content does not match original")
	}
}

func ProcessTestData(
	data []byte,
	fileKey []byte,
	process func(EncryptedChunk) error,
) error {
	path := filepath.Join(os.TempDir(), "tgstorage-restore-test-input.bin")
	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}
	defer os.Remove(path)

	return ProcessFileChunks(path, fileKey, process)
}
