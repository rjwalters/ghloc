package store

import (
	"context"
	"time"
)

// Snapshot represents a single LOC measurement for a repository at a point in time.
type Snapshot struct {
	ID        int64
	Owner     string
	Repo      string
	CommitSHA string
	TotalLOC  int64
	TotalFiles int64
	Languages []LanguageRecord
	CreatedAt time.Time
}

// LanguageRecord stores LOC for a single language within a snapshot.
type LanguageRecord struct {
	Language string
	Lines    int64
	Code     int64
	Comments int64
	Blanks   int64
	Files    int64
}

// Store defines the persistence interface for LOC data.
type Store interface {
	// SaveSnapshot persists a LOC snapshot for a repo.
	SaveSnapshot(ctx context.Context, snap *Snapshot) error

	// GetLatest returns the most recent snapshot for a repo.
	GetLatest(ctx context.Context, owner, repo string) (*Snapshot, error)

	// GetHistory returns all snapshots for a repo, ordered by time ascending.
	GetHistory(ctx context.Context, owner, repo string) ([]Snapshot, error)

	// Close releases resources.
	Close() error
}
