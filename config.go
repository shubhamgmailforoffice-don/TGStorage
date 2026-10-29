
package main

import "fmt"

// GetMasterKey loads the persistent master key from the local file.
func GetMasterKey() ([]byte, error) {
	key, err := LoadOrCreateMasterKey("master.key")
	if err != nil {
		return nil, fmt.Errorf("load master key: %w", err)
	}

	return key, nil
}
