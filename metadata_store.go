
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type MetadataStore struct {
	mu       sync.RWMutex
	path     string
	files    map[string]StoredFile
}

func NewMetadataStore(path string) (*MetadataStore, error) {
	if path == "" {
		return nil, errors.New("metadata path is required")
	}

	store := &MetadataStore{
		path:  path,
		files: make(map[string]StoredFile),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil
		}
		return nil, fmt.Errorf("read metadata: %w", err)
	}

	if err := json.Unmarshal(data, &store.files); err != nil {
		return nil, fmt.Errorf("parse metadata file: %w", err)
	}
	if store.files == nil {
		store.files = make(map[string]StoredFile)
	}

	return store, nil
}

// Save stores metadata under a caller-provided file ID.
// Store only wrapped file keys, never plaintext file keys.
func (s *MetadataStore) Save(id string, file StoredFile) error {
	if id == "" {
		return errors.New("file ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.files[id] = file

	data, err := json.MarshalIndent(s.files, "", "  ")
	if err != nil {
		delete(s.files, id)
		return fmt.Errorf("encode metadata: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		delete(s.files, id)
		return fmt.Errorf("create metadata directory: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0600); err != nil {
		delete(s.files, id)
		return fmt.Errorf("write metadata: %w", err)
	}

	return nil
}

// Get retrieves saved metadata for a file ID.
func (s *MetadataStore) Get(id string) (StoredFile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	file, ok := s.files[id]
	if !ok {
		return StoredFile{}, fmt.Errorf("file ID %q not found", id)
	}

	return file, nil
}

// ListIDs returns all saved file IDs.
func (s *MetadataStore) ListIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.files))
	for id := range s.files {
		ids = append(ids, id)
	}

	return ids
}
