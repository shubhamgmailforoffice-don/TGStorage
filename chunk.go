
package main

import (
	"fmt"
	"io"
	"os"
)

const ChunkSize int64 = 15 * 1024 * 1024 // 15 MiB

// SplitFile reads a file in fixed-size chunks and passes each chunk
// to the supplied callback. It keeps memory use bounded.
func SplitFile(path string, process func(index int, data []byte) error) error {
	if process == nil {
		return fmt.Errorf("chunk processor is required")
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open input file: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, ChunkSize)
	index := 0

	for {
		n, readErr := io.ReadFull(file, buffer)

		if n > 0 {
			chunk := make([]byte, n)
			copy(chunk, buffer[:n])

			if err := process(index, chunk); err != nil {
				return fmt.Errorf("process chunk %d: %w", index, err)
			}
			index++
		}

		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("read input file: %w", readErr)
		}
	}
}
