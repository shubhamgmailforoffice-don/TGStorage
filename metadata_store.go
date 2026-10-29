package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type MetadataStore struct {
	mu    sync.RWMutex
	path  string
	files map[string]StoredFile
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

// Save stores metadata without exposing the plaintext file key.
func (s *MetadataStore) Save(id string, file StoredFile) error {
	if id == "" {
		return errors.New("file ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Build a separate map so existing in-memory data remains unchanged
	// if encoding or writing fails.
	next := make(map[string]StoredFile, len(s.files)+1)
	for key, value := range s.files {
		next[key] = value
	}
	next[id] = file

	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create metadata directory: %w", err)
	}

	temp, err := os.CreateTemp(dir, ".metadata-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary metadata file: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)

	if err := temp.Chmod(0600); err != nil {
		_ = temp.Close()
		return fmt.Errorf("set metadata permissions: %w", err)
	}

	if _, err := temp.Write(data); err != nil {
		_ = temp.Close()
		return fmt.Errorf("write temporary metadata: %w", err)
	}

	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("sync temporary metadata: %w", err)
	}

	if err := temp.Close(); err != nil {
		return fmt.Errorf("close temporary metadata: %w", err)
	}

	if err := os.Rename(tempPath, s.path); err != nil {
		return fmt.Errorf("replace metadata file: %w", err)
	}

	s.files = next
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

// ListIDs returns saved file IDs in a stable order.
func (s *MetadataStore) ListIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.files))
	for id := range s.files {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	return ids
}
