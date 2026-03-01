// Package store provides SQLite persistence for hexboard messages.
// It wraps database/sql + mattn/go-sqlite3 with the narrow interface
// required by the web handler: open, save, and retrieve recent messages.
package store

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbPath = "/var/lib/hexboard/hexboard.db"
	schema = `CREATE TABLE IF NOT EXISTS messages (
		id      INTEGER  PRIMARY KEY AUTOINCREMENT,
		type    TEXT     NOT NULL DEFAULT 'text',
		content TEXT     NOT NULL,
		sent_at DATETIME NOT NULL
	)`
)

// OpenDB opens (or creates) the SQLite database at /var/lib/hexboard/hexboard.db.
// WAL mode is enabled via DSN. SetMaxOpenConns(1) serialises writes to avoid
// SQLITE_BUSY. The schema is created if it does not exist.
// Returns an error if the data directory does not exist or the DB cannot be opened.
func OpenDB() (*sql.DB, error) {
	dsn := dbPath + "?_journal_mode=WAL"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}

// Save inserts a text message into the messages table with the current UTC time.
// Returns an error on failure; the caller is responsible for logging and
// continuing (fail-soft per project decision — the message still displays).
func Save(db *sql.DB, content string) error {
	_, err := db.Exec(
		`INSERT INTO messages (type, content, sent_at) VALUES (?, ?, ?)`,
		"text", content, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// Recent returns the n most-recent message contents, newest first.
// Returns (nil, error) on query failure; the caller falls back to an empty list.
func Recent(db *sql.DB, n int) ([]string, error) {
	rows, err := db.Query(
		`SELECT content FROM messages ORDER BY id DESC LIMIT ?`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
