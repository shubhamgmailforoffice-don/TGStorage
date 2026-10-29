
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
)

type StoredFile struct {
	ID         string
	ChunkRefs  []TelegramChunkRef
	WrappedKey []byte
}

// UploadFile encrypts a file in chunks and uploads each encrypted chunk.
// Keep the master key safe: it is needed to recover the file key.
func UploadFile(
	ctx context.Context,
	storage *TelegramStorage,
	masterKey []byte,
	inputPath string,
) (*StoredFile, error) {
	if storage == nil {
		return nil, fmt.Errorf("Telegram storage is required")
	}

	fileKey, err := GenerateKey()
	if err != nil {
		return nil, err
	}

	wrappedKey, err := WrapFileKey(masterKey, fileKey)
	if err != nil {
		return nil, err
	}

	stored := &StoredFile{
    ID:        uuid.NewString(),
    WrappedKey: wrappedKey,
    ChunkRefs:  []TelegramChunkRef{},
}

	err = ProcessFileChunks(inputPath, fileKey, func(chunk EncryptedChunk) error {
		ref, err := storage.UploadEncryptedChunk(ctx, chunk)
		if err != nil {
			return err
		}
		stored.ChunkRefs = append(stored.ChunkRefs, ref)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("upload file chunks: %w", err)
	}

	metadataStore, err := NewMetadataStore("metadata.json")
if err != nil {
	return nil, fmt.Errorf("initialize metadata store: %w", err)
}

if err := metadataStore.Save(stored.ID, *stored); err != nil {
	return nil, fmt.Errorf("save file metadata: %w", err)
}

return stored, nil
}

// DownloadFile downloads, decrypts, and reconstructs a previously stored file.
func DownloadFile(
	ctx context.Context,
	storage *TelegramStorage,
	masterKey []byte,
	stored *StoredFile,
	outputPath string,
) error {
	if storage == nil || stored == nil {
		return fmt.Errorf("Telegram storage and file metadata are required")
	}

	fileKey, err := UnwrapFileKey(masterKey, stored.WrappedKey)
	if err != nil {
		return err
	}

	chunks := make([]EncryptedChunk, 0, len(stored.ChunkRefs))
	for _, ref := range stored.ChunkRefs {
		data, err := storage.DownloadEncryptedChunk(ctx, ref)
		if err != nil {
			return fmt.Errorf("download chunk %d: %w", ref.Index, err)
		}
		chunks = append(chunks, EncryptedChunk{
			Index: ref.Index,
			Data:  data,
		})
	}

	if len(chunks) == 0 {
		// Empty files contain no chunks.
		return os.WriteFile(outputPath, []byte{}, 0600)
	}

	return RestoreEncryptedChunks(chunks, fileKey, outputPath)
}

// GetStoredFile retrieves saved file metadata by its ID.
func GetStoredFile(id string) (*StoredFile, error) {
	metadataStore, err := NewMetadataStore("metadata.json")
	if err != nil {
		return nil, fmt.Errorf("initialize metadata store: %w", err)
	}

	stored, err := metadataStore.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get stored file: %w", err)
	}

	return &stored, nil
}
