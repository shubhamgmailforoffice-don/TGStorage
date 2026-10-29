package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("================================")
	fmt.Println("       TGStorage")
	fmt.Println("  Encrypted Telegram Storage")
	fmt.Println("================================")
	fmt.Println()

	masterKey, err := GetMasterKey()
	if err != nil {
		fmt.Println("ERROR: Could not load master key:", err)
		pause(reader)
		return
	}

	storage, err := NewTelegramStorageFromEnv()
	if err != nil {
		fmt.Println("ERROR: Telegram configuration:", err)
		fmt.Println("Check TGSTORE_BOT_TOKEN and TGSTORE_CHAT_ID.")
		pause(reader)
		return
	}

	ctx := context.Background()

	for {
		fmt.Println()
		fmt.Println("========== MAIN MENU ==========")
		fmt.Println("1. Upload a file")
		fmt.Println("2. List stored file IDs")
		fmt.Println("3. Download a file")
		fmt.Println("4. Exit")
		fmt.Println("===============================")

		choice := readInput(reader, "Choose an option: ")

		switch choice {
		case "1":
			fmt.Println()
			inputPath := readInput(reader, "Enter the full path of the file to upload: ")

			if inputPath == "" {
				fmt.Println("ERROR: File path cannot be empty.")
				continue
			}

			info, err := os.Stat(inputPath)
			if err != nil {
				fmt.Println("ERROR: Cannot access file:", err)
				continue
			}
			if !info.Mode().IsRegular() {
				fmt.Println("ERROR: Please select a regular file.")
				continue
			}

			fmt.Println("Encrypting and uploading file chunks...")

			stored, err := UploadFile(ctx, storage, masterKey, inputPath)
			if err != nil {
				fmt.Println("ERROR: Upload failed:", err)
				continue
			}

			fmt.Println()
			fmt.Println("SUCCESS: File uploaded and encrypted.")
			fmt.Println("File ID:", stored.ID)
			fmt.Println("Chunks uploaded:", len(stored.ChunkRefs))
			fmt.Println("Save the file ID to download this file later.")

		case "2":
			fmt.Println()
			ids, err := ListStoredFileIDs()
			if err != nil {
				fmt.Println("ERROR: Could not list files:", err)
				continue
			}

			if len(ids) == 0 {
				fmt.Println("No stored files found.")
				continue
			}

			fmt.Println("Stored file IDs:")
			for i, id := range ids {
				fmt.Printf("%d. %s\n", i+1, id)
			}

		case "3":
			fmt.Println()
			id := readInput(reader, "Enter the file ID: ")
			if id == "" {
				fmt.Println("ERROR: File ID cannot be empty.")
				continue
			}

			outputPath := readInput(reader, "Enter the output file path: ")
			if outputPath == "" {
				fmt.Println("ERROR: Output path cannot be empty.")
				continue
			}

			if _, err := os.Stat(outputPath); err == nil {
				overwrite := readInput(reader, "File exists. Overwrite it? (yes/no): ")
				if !strings.EqualFold(overwrite, "yes") {
					fmt.Println("Download cancelled.")
					continue
				}
			}

			fmt.Println("Downloading and decrypting file...")

			err := DownloadStoredFile(
				ctx,
				storage,
				masterKey,
				id,
				outputPath,
			)
			if err != nil {
				fmt.Println("ERROR: Download failed:", err)
				continue
			}

			fmt.Println("SUCCESS: File restored to:", outputPath)

		case "4":
			fmt.Println("TGStorage closed.")
			return

		default:
			fmt.Println("Invalid option. Enter 1, 2, 3, or 4.")
		}
	}
}

func readInput(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil && len(input) == 0 {
		return ""
	}
	return strings.TrimSpace(input)
}

func pause(reader *bufio.Reader) {
	fmt.Print("Press Enter to close...")
	_, _ = reader.ReadString('\n')
}
