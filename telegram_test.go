package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestTelegramStorageConfigured(t *testing.T) {
	storage, err := NewTelegramStorageFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	url := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getMe",
		storage.BotToken,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := storage.HTTPClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Telegram getMe returned HTTP %d", resp.StatusCode)
	}
}