
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSplitFile(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{name: "empty file", data: nil, want: 0},
		{name: "small file", data: []byte("TGStorage test"), want: 1},
		{
			name: "larger than one chunk",
			data: bytes.Repeat([]byte("A"), int(ChunkSize)+123),
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "input.bin")
			if err := os.WriteFile(path, tt.data, 0600); err != nil {
				t.Fatal(err)
			}

			var combined []byte
			count := 0

			err := SplitFile(path, func(index int, chunk []byte) error {
				if index != count {
					t.Fatalf("expected chunk index %d, got %d", count, index)
				}
				count++
				combined = append(combined, chunk...)
				return nil
			})
			if err != nil {
				t.Fatalf("SplitFile failed: %v", err)
			}

			if count != tt.want {
				t.Fatalf("expected %d chunks, got %d", tt.want, count)
			}
			if !bytes.Equal(combined, tt.data) {
				t.Fatal("combined chunks do not match original data")
			}
		})
	}
}
