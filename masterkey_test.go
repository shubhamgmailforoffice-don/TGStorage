
package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestLoadOrCreateMasterKeyPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "master.key")

	first, err := LoadOrCreateMasterKey(path)
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	if len(first) != EncryptionKeySize {
		t.Fatalf("expected %d-byte key, got %d", EncryptionKeySize, len(first))
	}

	second, err := LoadOrCreateMasterKey(path)
	if err != nil {
		t.Fatalf("second load: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Fatal("master key changed between loads")
	}
}
