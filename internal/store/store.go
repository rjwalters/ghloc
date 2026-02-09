package store

import "time"

// Snapshot represents a single LOC measurement at a point in time.
type Snapshot struct {
	TotalLOC   int64            `json:"total_loc"`
	TotalFiles int64            `json:"total_files"`
	Languages  []LanguageRecord `json:"languages"`
	CreatedAt  time.Time        `json:"created_at"`
}

// LanguageRecord stores LOC for a single language within a snapshot.
type LanguageRecord struct {
	Language string `json:"language"`
	Lines    int64  `json:"lines"`
	Code     int64  `json:"code"`
	Comments int64  `json:"comments"`
	Blanks   int64  `json:"blanks"`
	Files    int64  `json:"files"`
}
