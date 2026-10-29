
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type TelegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      struct {
		ID        int64  `json:"id"`
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
	} `json:"result"`
}

func main() {
	token := os.Getenv("TGSTORE_BOT_TOKEN")
	if token == "" {
		fmt.Println("ERROR: TGSTORE_BOT_TOKEN is not set.")
		fmt.Println("Set your bot token in this CMD window first.")
		os.Exit(1)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	url := "https://api.telegram.org/bot" + token + "/getMe"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println("ERROR: Could not create request.")
		os.Exit(1)
	}

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
		fmt.Println("ERROR: Telegram rejected the token.")
		fmt.Println("Check that your current token is valid.")
		os.Exit(1)
	}

	fmt.Println("SUCCESS: Connected to Telegram!")
	fmt.Println("Bot name:", result.Result.FirstName)
	fmt.Println("Bot username:", "@"+result.Result.Username)
	fmt.Println("Bot ID:", result.Result.ID)
	fmt.Println("Token was not displayed.")
}
