package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadHistory_NoFile(t *testing.T) {
	snapshots, err := LoadHistory(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err != nil {
		t.Fatalf("LoadHistory() error: %v", err)
	}
	if snapshots != nil {
		t.Errorf("expected nil, got %v", snapshots)
	}
}

func TestSaveAndLoadHistory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "history.json")

	now := time.Now().UTC().Truncate(time.Second)
	snapshots := []Snapshot{
		{
			TotalLOC:   1000,
			TotalFiles: 10,
			Languages: []LanguageRecord{
				{Language: "Go", Lines: 800, Code: 600, Comments: 100, Blanks: 100, Files: 8},
				{Language: "Markdown", Lines: 200, Code: 200, Comments: 0, Blanks: 0, Files: 2},
			},
			CreatedAt: now,
		},
	}

	if err := SaveHistory(path, snapshots); err != nil {
		t.Fatalf("SaveHistory() error: %v", err)
	}

	loaded, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("LoadHistory() error: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("expected 1 snapshot, got %d", len(loaded))
	}
	if loaded[0].TotalLOC != 1000 {
		t.Errorf("TotalLOC: got %d, want 1000", loaded[0].TotalLOC)
	}
	if loaded[0].TotalFiles != 10 {
		t.Errorf("TotalFiles: got %d, want 10", loaded[0].TotalFiles)
	}
	if len(loaded[0].Languages) != 2 {
		t.Fatalf("expected 2 languages, got %d", len(loaded[0].Languages))
	}
	if loaded[0].Languages[0].Language != "Go" {
		t.Errorf("first language: got %q, want Go", loaded[0].Languages[0].Language)
	}
	if !loaded[0].CreatedAt.Equal(now) {
		t.Errorf("CreatedAt: got %v, want %v", loaded[0].CreatedAt, now)
	}
}

func TestSaveHistory_Append(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	// Save first snapshot
	snap1 := []Snapshot{{TotalLOC: 100, TotalFiles: 5, CreatedAt: base}}
	if err := SaveHistory(path, snap1); err != nil {
		t.Fatalf("SaveHistory() error: %v", err)
	}

	// Load, append, save
	loaded, _ := LoadHistory(path)
	loaded = append(loaded, Snapshot{TotalLOC: 200, TotalFiles: 8, CreatedAt: base.Add(24 * time.Hour)})
	if err := SaveHistory(path, loaded); err != nil {
		t.Fatalf("SaveHistory() error: %v", err)
	}

	// Verify
	final, _ := LoadHistory(path)
	if len(final) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(final))
	}
	if final[0].TotalLOC != 100 {
		t.Errorf("first snapshot TotalLOC: got %d, want 100", final[0].TotalLOC)
	}
	if final[1].TotalLOC != 200 {
		t.Errorf("second snapshot TotalLOC: got %d, want 200", final[1].TotalLOC)
	}
}

func TestLoadHistory_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	os.WriteFile(path, []byte("not json"), 0644)

	_, err := LoadHistory(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
