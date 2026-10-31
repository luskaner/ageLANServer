// Package web serves the LAN dashboard: a read-only public site and API with
// everything the server knows (lobbies, queue, matches, ladder, players, map
// pool) plus an admin panel that edits the 1v1 map pool and the settings of
// config.toml that affect the game.
//
// The dashboard runs on its own listener, independent of the game's TLS
// listener on :443, so enabling it never changes how the client talks to the
// server.
package web

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

// AdminStore keeps the small amount of state the dashboard must remember
// across restarts: the admin password hash and a few settings. It lives in its
// own SQLite file next to the per-game user data.
type AdminStore struct {
	mu   sync.Mutex
	db   *sql.DB
	path string
}

// NewAdminStore opens (and creates when missing) the dashboard database in
// dataDir.
func NewAdminStore(dataDir string) (*AdminStore, error) {
	store := &AdminStore{path: filepath.Join(dataDir, "web.db")}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create %s: %w", dataDir, err)
	}
	db, err := sql.Open("sqlite", store.path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", store.path, err)
	}
	db.SetMaxOpenConns(1)
	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA busy_timeout=5000",
	} {
		if _, err = db.Exec(pragma); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("%s: %w", pragma, err)
		}
	}
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS settings (
		key   TEXT PRIMARY KEY,
		value TEXT NOT NULL DEFAULT ''
	)`); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create settings table: %w", err)
	}
	store.db = db
	return store, nil
}

// Path returns the database file.
func (s *AdminStore) Path() string { return s.path }

// Close closes the database.
func (s *AdminStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

// Get returns a stored setting and distinguishes a missing value from a
// database failure.
func (s *AdminStore) Get(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return "", false, fmt.Errorf("admin database is not open")
	}
	var value string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

// Set stores a setting.
func (s *AdminStore) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return fmt.Errorf("admin database is not open")
	}
	_, err := s.db.Exec(
		`INSERT INTO settings (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}

// SetIfAbsent atomically stores a setting only when no value exists yet.
func (s *AdminStore) SetIfAbsent(key, value string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return false, fmt.Errorf("admin database is not open")
	}
	result, err := s.db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES (?, ?)`, key, value)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows == 1, nil
}

func (s *AdminStore) CompleteSetup(passwordHash, publicTitle string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return false, fmt.Errorf("admin database is not open")
	}
	tx, err := s.db.Begin()
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.Exec(
		`INSERT OR IGNORE INTO settings (key, value) VALUES ('admin_password_hash', ?)`,
		passwordHash,
	)
	if err != nil {
		return false, err
	}
	created, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if created == 0 {
		return false, nil
	}
	if _, err = tx.Exec(
		`INSERT INTO settings (key, value) VALUES ('public_title', ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		publicTitle,
	); err != nil {
		return false, err
	}
	if err = tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}
