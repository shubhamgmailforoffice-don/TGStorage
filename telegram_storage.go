
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

type TelegramStorage struct {
	BotToken   string
	ChatID     string
	HTTPClient *http.Client
}

type TelegramChunkRef struct {
	Index     int
	ChatID    int64
	MessageID int64
	FileID    string
}

func NewTelegramStorageFromEnv() (*TelegramStorage, error) {
	token := strings.TrimSpace(os.Getenv("TGSTORE_BOT_TOKEN"))
	chatID := strings.TrimSpace(os.Getenv("TGSTORE_CHAT_ID"))

	if token == "" {
		return nil, fmt.Errorf("TGSTORE_BOT_TOKEN is not set")
	}
	if chatID == "" {
		return nil, fmt.Errorf("TGSTORE_CHAT_ID is not set")
	}

	return &TelegramStorage{
		BotToken: token,
		ChatID:   chatID,
		HTTPClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}, nil
}

// UploadEncryptedChunk uploads encrypted bytes as a Telegram document.
// Save the returned reference in the metadata database later.
func (s *TelegramStorage) UploadEncryptedChunk(
	ctx context.Context,
	chunk EncryptedChunk,
) (TelegramChunkRef, error) {
	var ref TelegramChunkRef

	if s == nil || s.BotToken == "" || s.ChatID == "" {
		return ref, fmt.Errorf("Telegram storage is not configured")
	}
	if chunk.Index < 0 {
		return ref, fmt.Errorf("chunk index cannot be negative")
	}
	if len(chunk.Data) == 0 {
		return ref, fmt.Errorf("encrypted chunk is empty")
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("chat_id", s.ChatID); err != nil {
		return ref, fmt.Errorf("add chat ID: %w", err)
	}

	filename := fmt.Sprintf("chunk-%06d.enc", chunk.Index)
	part, err := writer.CreateFormFile("document", filename)
	if err != nil {
		return ref, fmt.Errorf("create upload form: %w", err)
	}
	if _, err := part.Write(chunk.Data); err != nil {
		return ref, fmt.Errorf("write encrypted chunk: %w", err)
	}
	if err := writer.Close(); err != nil {
		return ref, fmt.Errorf("finish upload form: %w", err)
	}

	endpoint := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendDocument",
		s.BotToken,
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		&body,
	)
	if err != nil {
		return ref, fmt.Errorf("create Telegram request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}

	resp, err := client.Do(req)
	if err != nil {
		return ref, fmt.Errorf("upload encrypted chunk: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
		Result      struct {
			MessageID int64 `json:"message_id"`
			Chat      struct {
				ID int64 `json:"id"`
			} `json:"chat"`
			Document struct {
				FileID string `json:"file_id"`
			} `json:"document"`
		} `json:"result"`
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return ref, fmt.Errorf("read Telegram response (HTTP %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 || !result.OK {
		return ref, fmt.Errorf("Telegram upload failed (HTTP %d): %s", resp.StatusCode, result.Description)
	}
	if result.Result.Document.FileID == "" || result.Result.MessageID == 0 {
		return ref, fmt.Errorf("Telegram response did not include a document reference")
	}

	return TelegramChunkRef{
		Index:     chunk.Index,
		ChatID:    result.Result.Chat.ID,
		MessageID: result.Result.MessageID,
		FileID:    result.Result.Document.FileID,
	}, nil
}



// DownloadEncryptedChunk downloads a previously uploaded Telegram document.
func (s *TelegramStorage) DownloadEncryptedChunk(
	ctx context.Context,
	ref TelegramChunkRef,
) ([]byte, error) {
	if s == nil || s.BotToken == "" {
		return nil, fmt.Errorf("Telegram storage is not configured")
	}
	if ref.FileID == "" {
		return nil, fmt.Errorf("Telegram file ID is required")
	}

	endpoint := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getFile?file_id=%s",
		s.BotToken,
		ref.FileID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create getFile request: %w", err)
	}

	client := s.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Minute}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request Telegram file info: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
		Result      struct {
			FilePath string `json:"file_path"`
		} `json:"result"`
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return nil, fmt.Errorf("read getFile response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || !result.OK {
		return nil, fmt.Errorf("Telegram getFile failed: %s", result.Description)
	}
	if result.Result.FilePath == "" {
		return nil, fmt.Errorf("Telegram did not return a file path")
	}

	downloadURL := fmt.Sprintf(
		"https://api.telegram.org/file/bot%s/%s",
		s.BotToken,
		result.Result.FilePath,
	)

	downloadReq, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create download request: %w", err)
	}

	downloadResp, err := client.Do(downloadReq)
	if err != nil {
		return nil, fmt.Errorf("download Telegram file: %w", err)
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode < 200 || downloadResp.StatusCode >= 300 {
		return nil, fmt.Errorf("Telegram download returned HTTP %d", downloadResp.StatusCode)
	}

	const maxEncryptedChunkSize = int64(16 * 1024 * 1024)
	data, err := io.ReadAll(io.LimitReader(downloadResp.Body, maxEncryptedChunkSize+1))
	if err != nil {
		return nil, fmt.Errorf("read downloaded chunk: %w", err)
	}
	if int64(len(data)) > maxEncryptedChunkSize {
		return nil, fmt.Errorf("downloaded chunk exceeds maximum allowed size")
	}

	return data, nil
}
