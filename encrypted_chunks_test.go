
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestProcessFileChunks(t *testing.T) {
	original := bytes.Repeat([]byte("TGStorage"), int(ChunkSize/9)+1)

	path := filepath.Join(t.TempDir(), "input.bin")
	if err := os.WriteFile(path, original, 0600); err != nil {
		t.Fatal(err)
	}

	fileKey, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	var restored []byte
	count := 0

	err = ProcessFileChunks(path, fileKey, func(chunk EncryptedChunk) error {
		if chunk.Index != count {
			t.Fatalf("expected index %d, got %d", count, chunk.Index)
		}

		plaintext, err := Decrypt(fileKey, chunk.Data)
		if err != nil {
			t.Fatalf("decrypt chunk %d: %v", chunk.Index, err)
		}

		restored = append(restored, plaintext...)
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("ProcessFileChunks failed: %v", err)
	}

	if count != 2 {
		t.Fatalf("expected 2 chunks, got %d", count)
	}

	if !bytes.Equal(restored, original) {
		t.Fatal("restored data does not match original")
	}
}
