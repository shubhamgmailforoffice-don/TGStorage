
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type UpdateResponse struct {
	OK     bool `json:"ok"`
	Result []struct {
		Message struct {
			Text string `json:"text"`
			Chat struct {
				ID   int64  `json:"id"`
				Type string `json:"type"`
				Title string `json:"title"`
			} `json:"chat"`
		} `json:"message"`
	} `json:"result"`
}

func main() {
	token := os.Getenv("TGSTORE_BOT_TOKEN")
	if token == "" {
		fmt.Println("ERROR: TGSTORE_BOT_TOKEN is not set.")
		os.Exit(1)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	url := "https://api.telegram.org/bot" + token + "/getUpdates"

	resp, err := client.Get(url)
	if err != nil {
		fmt.Println("ERROR: Could not connect to Telegram.")
		os.Exit(1)
	}
	defer resp.Body.Close()

	var data UpdateResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("ERROR: Could not read Telegram response.")
		os.Exit(1)
	}

	if !data.OK {
		fmt.Println("ERROR: Telegram API request failed.")
		fmt.Println("Check the bot token and whether another process is polling updates.")
		os.Exit(1)
	}

	found := false
	for _, update := range data.Result {
		if update.Message.Text == "TGSTORE_CONNECT" {
			fmt.Println("Storage group found!")
			fmt.Println("Group title:", update.Message.Chat.Title)
			fmt.Println("Chat type:", update.Message.Chat.Type)
			fmt.Println("Chat ID:", update.Message.Chat.ID)
			found = true
		}
	}

	if !found {
		fmt.Println("No matching message found.")
		fmt.Println("Send TGSTORE_CONNECT in your group, then run this program again.")
	}
}
