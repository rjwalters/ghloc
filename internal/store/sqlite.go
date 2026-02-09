package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLite opens (or creates) a SQLite database and runs migrations.
func NewSQLite(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS snapshots (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		owner      TEXT NOT NULL,
		repo       TEXT NOT NULL,
		commit_sha TEXT NOT NULL,
		total_loc  INTEGER NOT NULL,
		total_files INTEGER NOT NULL,
		created_at DATETIME NOT NULL DEFAULT (datetime('now'))
	);

	CREATE INDEX IF NOT EXISTS idx_snapshots_repo ON snapshots(owner, repo, created_at);

	CREATE TABLE IF NOT EXISTS languages (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		snapshot_id INTEGER NOT NULL REFERENCES snapshots(id) ON DELETE CASCADE,
		language    TEXT NOT NULL,
		lines       INTEGER NOT NULL,
		code        INTEGER NOT NULL,
		comments    INTEGER NOT NULL,
		blanks      INTEGER NOT NULL,
		files       INTEGER NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_languages_snapshot ON languages(snapshot_id);
	`
	_, err := db.Exec(schema)
	return err
}

func (s *SQLiteStore) SaveSnapshot(ctx context.Context, snap *Snapshot) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO snapshots (owner, repo, commit_sha, total_loc, total_files, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		snap.Owner, snap.Repo, snap.CommitSHA, snap.TotalLOC, snap.TotalFiles, snap.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert snapshot: %w", err)
	}

	snapID, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	snap.ID = snapID

	for _, lang := range snap.Languages {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO languages (snapshot_id, language, lines, code, comments, blanks, files)
			 VALUES (?, ?, ?, ?, ?, ?, ?)`,
			snapID, lang.Language, lang.Lines, lang.Code, lang.Comments, lang.Blanks, lang.Files,
		)
		if err != nil {
			return fmt.Errorf("insert language %s: %w", lang.Language, err)
		}
	}

	return tx.Commit()
}

func (s *SQLiteStore) GetLatest(ctx context.Context, owner, repo string) (*Snapshot, error) {
	snap := &Snapshot{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, owner, repo, commit_sha, total_loc, total_files, created_at
		 FROM snapshots WHERE owner = ? AND repo = ?
		 ORDER BY created_at DESC LIMIT 1`,
		owner, repo,
	).Scan(&snap.ID, &snap.Owner, &snap.Repo, &snap.CommitSHA, &snap.TotalLOC, &snap.TotalFiles, &snap.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query latest: %w", err)
	}

	langs, err := s.getLanguages(ctx, snap.ID)
	if err != nil {
		return nil, err
	}
	snap.Languages = langs

	return snap, nil
}

func (s *SQLiteStore) GetHistory(ctx context.Context, owner, repo string) ([]Snapshot, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, owner, repo, commit_sha, total_loc, total_files, created_at
		 FROM snapshots WHERE owner = ? AND repo = ?
		 ORDER BY created_at ASC`,
		owner, repo,
	)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	var snaps []Snapshot
	for rows.Next() {
		var snap Snapshot
		if err := rows.Scan(&snap.ID, &snap.Owner, &snap.Repo, &snap.CommitSHA, &snap.TotalLOC, &snap.TotalFiles, &snap.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan snapshot: %w", err)
		}
		snaps = append(snaps, snap)
	}

	return snaps, rows.Err()
}

func (s *SQLiteStore) getLanguages(ctx context.Context, snapshotID int64) ([]LanguageRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT language, lines, code, comments, blanks, files
		 FROM languages WHERE snapshot_id = ? ORDER BY code DESC`,
		snapshotID,
	)
	if err != nil {
		return nil, fmt.Errorf("query languages: %w", err)
	}
	defer rows.Close()

	var langs []LanguageRecord
	for rows.Next() {
		var l LanguageRecord
		if err := rows.Scan(&l.Language, &l.Lines, &l.Code, &l.Comments, &l.Blanks, &l.Files); err != nil {
			return nil, fmt.Errorf("scan language: %w", err)
		}
		langs = append(langs, l)
	}

	return langs, rows.Err()
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// ensure SQLiteStore implements Store
var _ Store = (*SQLiteStore)(nil)
