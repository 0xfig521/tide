package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps the SQLite connection with application-specific helpers.
type DB struct {
	Conn *sql.DB
	Path string
}

// Open initializes the SQLite database at the given path with WAL mode.
func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	// modernc.org/sqlite DSN: pragmas via _pragma prefix
	dsn := fmt.Sprintf("file:%s?_pragma.journal_mode=WAL&_pragma.busy_timeout=5000&_pragma.foreign_keys=ON", dbPath)

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// SQLite only supports single writer; set pool size to 1
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	db := &DB{Conn: conn, Path: dbPath}

	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.Conn.Close()
}

// DefaultDBPath returns the default database path following XDG conventions.
func DefaultDBPath() string {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(".", "tide.db")
		}
		dataDir = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataDir, "tide", "tide.db")
}
