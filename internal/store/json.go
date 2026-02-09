package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadHistory reads a slice of Snapshots from a JSON file.
// Returns an empty slice if the file does not exist.
func LoadHistory(path string) ([]Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read history: %w", err)
	}

	var snapshots []Snapshot
	if err := json.Unmarshal(data, &snapshots); err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}
	return snapshots, nil
}

// SaveHistory writes a slice of Snapshots to a JSON file, creating parent directories as needed.
func SaveHistory(path string, snapshots []Snapshot) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	data, err := json.MarshalIndent(snapshots, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write history: %w", err)
	}
	return nil
}
