
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// LoadOrCreateMasterKey loads an existing master key or creates one.
func LoadOrCreateMasterKey(path string) ([]byte, error) {
	if path == "" {
		return nil, errors.New("master key path is required")
	}

	key, err := os.ReadFile(path)
	if err == nil {
		if len(key) != EncryptionKeySize {
			return nil, errors.New("stored master key has invalid length")
		}
		return key, nil
	}

	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read master key: %w", err)
	}

	key, err = GenerateMasterKey()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("create master key directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, fmt.Errorf("create master key file: %w", err)
	}

	_, writeErr := file.Write(key)
	closeErr := file.Close()

	if writeErr != nil {
		_ = os.Remove(path)
		return nil, fmt.Errorf("write master key: %w", writeErr)
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return nil, fmt.Errorf("close master key file: %w", closeErr)
	}

	return key, nil
}
