package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	s, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("NewSQLite: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestSaveAndGetLatest(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	snap := &Snapshot{
		Owner:      "testowner",
		Repo:       "testrepo",
		CommitSHA:  "abc123",
		TotalLOC:   1500,
		TotalFiles: 10,
		CreatedAt:  time.Now().UTC(),
		Languages: []LanguageRecord{
			{Language: "Go", Lines: 1200, Code: 1000, Comments: 100, Blanks: 100, Files: 8},
			{Language: "Markdown", Lines: 300, Code: 300, Comments: 0, Blanks: 0, Files: 2},
		},
	}

	if err := s.SaveSnapshot(ctx, snap); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}

	if snap.ID == 0 {
		t.Error("expected snapshot ID to be set after save")
	}

	got, err := s.GetLatest(ctx, "testowner", "testrepo")
	if err != nil {
		t.Fatalf("GetLatest: %v", err)
	}
	if got == nil {
		t.Fatal("GetLatest returned nil")
	}

	if got.TotalLOC != 1500 {
		t.Errorf("TotalLOC: got %d, want 1500", got.TotalLOC)
	}
	if got.TotalFiles != 10 {
		t.Errorf("TotalFiles: got %d, want 10", got.TotalFiles)
	}
	if got.CommitSHA != "abc123" {
		t.Errorf("CommitSHA: got %q, want %q", got.CommitSHA, "abc123")
	}
	if len(got.Languages) != 2 {
		t.Fatalf("Languages: got %d, want 2", len(got.Languages))
	}
	// Languages are ordered by code DESC
	if got.Languages[0].Language != "Go" {
		t.Errorf("first language: got %q, want Go", got.Languages[0].Language)
	}
	if got.Languages[0].Code != 1000 {
		t.Errorf("Go code: got %d, want 1000", got.Languages[0].Code)
	}
}

func TestGetLatest_ReturnsNewest(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Save two snapshots
	for i, sha := range []string{"first", "second"} {
		snap := &Snapshot{
			Owner:      "owner",
			Repo:       "repo",
			CommitSHA:  sha,
			TotalLOC:   int64((i + 1) * 1000),
			TotalFiles: int64((i + 1) * 5),
			CreatedAt:  time.Now().UTC().Add(time.Duration(i) * time.Hour),
		}
		if err := s.SaveSnapshot(ctx, snap); err != nil {
			t.Fatalf("SaveSnapshot %s: %v", sha, err)
		}
	}

	got, err := s.GetLatest(ctx, "owner", "repo")
	if err != nil {
		t.Fatalf("GetLatest: %v", err)
	}

	if got.CommitSHA != "second" {
		t.Errorf("expected latest to be 'second', got %q", got.CommitSHA)
	}
	if got.TotalLOC != 2000 {
		t.Errorf("expected TotalLOC 2000, got %d", got.TotalLOC)
	}
}

func TestGetLatest_NoData(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	got, err := s.GetLatest(ctx, "nonexistent", "repo")
	if err != nil {
		t.Fatalf("GetLatest: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for nonexistent repo, got %+v", got)
	}
}

func TestGetHistory(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		snap := &Snapshot{
			Owner:      "owner",
			Repo:       "repo",
			CommitSHA:  "sha" + string(rune('a'+i)),
			TotalLOC:   int64((i + 1) * 100),
			TotalFiles: int64(i + 1),
			CreatedAt:  base.Add(time.Duration(i) * 24 * time.Hour),
		}
		if err := s.SaveSnapshot(ctx, snap); err != nil {
			t.Fatalf("SaveSnapshot %d: %v", i, err)
		}
	}

	history, err := s.GetHistory(ctx, "owner", "repo")
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}

	if len(history) != 5 {
		t.Fatalf("expected 5 snapshots, got %d", len(history))
	}

	// Should be ordered by time ascending
	for i := 1; i < len(history); i++ {
		if !history[i].CreatedAt.After(history[i-1].CreatedAt) {
			t.Errorf("snapshot %d not after snapshot %d", i, i-1)
		}
	}

	if history[0].TotalLOC != 100 {
		t.Errorf("first snapshot TotalLOC: got %d, want 100", history[0].TotalLOC)
	}
	if history[4].TotalLOC != 500 {
		t.Errorf("last snapshot TotalLOC: got %d, want 500", history[4].TotalLOC)
	}
}

func TestGetHistory_Empty(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	history, err := s.GetHistory(ctx, "nonexistent", "repo")
	if err != nil {
		t.Fatalf("GetHistory: %v", err)
	}
	if len(history) != 0 {
		t.Errorf("expected empty history, got %d", len(history))
	}
}

func TestGetHistory_IsolatesRepos(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Save snapshots for two different repos
	for _, repo := range []string{"repo-a", "repo-b"} {
		snap := &Snapshot{
			Owner:      "owner",
			Repo:       repo,
			CommitSHA:  "sha1",
			TotalLOC:   100,
			TotalFiles: 1,
			CreatedAt:  time.Now().UTC(),
		}
		if err := s.SaveSnapshot(ctx, snap); err != nil {
			t.Fatal(err)
		}
	}

	history, err := s.GetHistory(ctx, "owner", "repo-a")
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 {
		t.Errorf("expected 1 snapshot for repo-a, got %d", len(history))
	}
}
