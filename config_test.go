
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestGetMasterKeyPersists(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tempDir := t.TempDir()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	first, err := GetMasterKey()
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tempDir, "master.key")); err != nil {
		t.Fatalf("master key file was not created: %v", err)
	}

	second, err := GetMasterKey()
	if err != nil {
		t.Fatalf("second call: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Fatal("master key changed between calls")
	}
}
