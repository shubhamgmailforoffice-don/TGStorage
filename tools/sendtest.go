
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type TelegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func main() {
	token := os.Getenv("TGSTORE_BOT_TOKEN")
	if token == "" {
		fmt.Println("ERROR: TGSTORE_BOT_TOKEN is not set.")
		os.Exit(1)
	}

	// Replace this with the chat ID of your intended storage group.
	const chatID int64 = -1001371505134

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    "MyTGStore connection test successful.",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("ERROR: Could not prepare message.")
		os.Exit(1)
	}

	url := "https://api.telegram.org/bot" + token + "/sendMessage"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		fmt.Println("ERROR: Could not create request.")
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("ERROR: Could not connect to Telegram.")
		os.Exit(1)
	}
	defer resp.Body.Close()

	var result TelegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("ERROR: Could not read Telegram response.")
		os.Exit(1)
	}

	if !result.OK {
		fmt.Println("ERROR: Telegram rejected the message.")
		fmt.Println("Description:", result.Description)
		os.Exit(1)
	}

	fmt.Println("SUCCESS: Test message sent.")
	fmt.Println("Check the intended Telegram group.")
}
