package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type ConnectionRequest struct {
	ID         int64
	ProfileURL string
	Name       string
	SentAt     time.Time
	Accepted   bool
	Note       string
}

type Message struct {
	ID         int64
	ProfileURL string
	Content    string
	SentAt     time.Time
}

func New(dbPath string) (*Store, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &Store{db: db}
	if err := store.createTables(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS connection_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_url TEXT UNIQUE NOT NULL,
			name TEXT,
			sent_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			accepted BOOLEAN DEFAULT 0,
			note TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			profile_url TEXT NOT NULL,
			content TEXT NOT NULL,
			sent_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_profile_url ON connection_requests(profile_url)`,
		`CREATE INDEX IF NOT EXISTS idx_sent_at ON connection_requests(sent_at)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("failed to create tables: %w", err)
		}
	}

	return nil
}

func (s *Store) SaveConnectionRequest(profileURL, name, note string) error {
	query := `INSERT INTO connection_requests (profile_url, name, note) VALUES (?, ?, ?)`
	_, err := s.db.Exec(query, profileURL, name, note)
	if err != nil {
		return fmt.Errorf("failed to save connection request: %w", err)
	}
	return nil
}

func (s *Store) IsConnectionSent(profileURL string) (bool, error) {
	query := `SELECT COUNT(*) FROM connection_requests WHERE profile_url = ?`
	var count int
	err := s.db.QueryRow(query, profileURL).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check connection: %w", err)
	}
	return count > 0, nil
}

func (s *Store) GetConnectionsCountToday() (int, error) {
	query := `SELECT COUNT(*) FROM connection_requests WHERE DATE(sent_at) = DATE('now')`
	var count int
	err := s.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get connection count: %w", err)
	}
	return count, nil
}

func (s *Store) SaveMessage(profileURL, content string) error {
	query := `INSERT INTO messages (profile_url, content) VALUES (?, ?)`
	_, err := s.db.Exec(query, profileURL, content)
	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}

func (s *Store) GetMessagesCountToday() (int, error) {
	query := `SELECT COUNT(*) FROM messages WHERE DATE(sent_at) = DATE('now')`
	var count int
	err := s.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get message count: %w", err)
	}
	return count, nil
}

func (s *Store) MarkConnectionAccepted(profileURL string) error {
	query := `UPDATE connection_requests SET accepted = 1 WHERE profile_url = ?`
	_, err := s.db.Exec(query, profileURL)
	return err
}

func (s *Store) GetPendingConnections() ([]ConnectionRequest, error) {
	query := `SELECT id, profile_url, name, sent_at, accepted, note 
	          FROM connection_requests WHERE accepted = 0 ORDER BY sent_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []ConnectionRequest
	for rows.Next() {
		var r ConnectionRequest
		if err := rows.Scan(&r.ID, &r.ProfileURL, &r.Name, &r.SentAt, &r.Accepted, &r.Note); err != nil {
			return nil, err
		}
		requests = append(requests, r)
	}

	return requests, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
