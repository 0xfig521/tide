package db

import (
	"fmt"
)

// migrate runs all pending migrations in order.
func (db *DB) migrate() error {
	migrations := []struct {
		version int
		sql     string
	}{
		{1, schemaV1},
		{2, schemaV2},
		{3, schemaV3},
	}

	// Create schema version table if not exists
	if _, err := db.Conn.Exec(`CREATE TABLE IF NOT EXISTS schema_version (version INTEGER PRIMARY KEY, applied_at TEXT DEFAULT (datetime('now')))`); err != nil {
		return fmt.Errorf("create schema_version: %w", err)
	}

	var currentVersion int
	err := db.Conn.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}

	for _, m := range migrations {
		if m.version <= currentVersion {
			continue
		}
		if _, err := db.Conn.Exec(m.sql); err != nil {
			return fmt.Errorf("migration v%d: %w", m.version, err)
		}
		if _, err := db.Conn.Exec(`INSERT INTO schema_version (version) VALUES (?)`, m.version); err != nil {
			return fmt.Errorf("record migration v%d: %w", m.version, err)
		}
	}

	return nil
}

const schemaV1 = `
CREATE TABLE IF NOT EXISTS categories (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    created_at  TEXT DEFAULT (datetime('now')),
    updated_at  TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS feeds (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    title                 TEXT NOT NULL DEFAULT '',
    description           TEXT DEFAULT '',
    site_url              TEXT DEFAULT '',
    feed_url              TEXT NOT NULL UNIQUE,
    image_url             TEXT DEFAULT '',
    language              TEXT DEFAULT '',
    feed_type             TEXT DEFAULT '',
    etag_header           TEXT DEFAULT '',
    last_modified_header  TEXT DEFAULT '',
    checked_at            TEXT,
    last_fetched_at       TEXT,
    http_status_code      INTEGER NOT NULL DEFAULT 0,
    next_check_at         TEXT NOT NULL DEFAULT (datetime('now')),
    parsing_error_count   INTEGER NOT NULL DEFAULT 0,
    parsing_error_msg     TEXT DEFAULT '',
    is_active             INTEGER NOT NULL DEFAULT 1,
    created_at            TEXT DEFAULT (datetime('now')),
    updated_at            TEXT DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS feed_categories (
    feed_id     INTEGER NOT NULL,
    category_id INTEGER NOT NULL,
    PRIMARY KEY (feed_id, category_id),
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS entries (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id         INTEGER NOT NULL,
    title           TEXT NOT NULL DEFAULT '',
    url             TEXT NOT NULL,
    guid            TEXT NOT NULL,
    content         TEXT DEFAULT '',
    description     TEXT DEFAULT '',
    author_name     TEXT DEFAULT '',
    image_url       TEXT DEFAULT '',
    categories      TEXT DEFAULT '',
    published_at    TEXT,
    updated_at      TEXT DEFAULT (datetime('now')),
    is_read         INTEGER NOT NULL DEFAULT 0,
    is_starred      INTEGER NOT NULL DEFAULT 0,
    hash            TEXT NOT NULL,
    created_at      TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_entries_hash ON entries(feed_id, hash);
CREATE INDEX IF NOT EXISTS idx_entries_feed_id ON entries(feed_id);
CREATE INDEX IF NOT EXISTS idx_entries_published_at ON entries(published_at DESC);
CREATE INDEX IF NOT EXISTS idx_entries_is_read ON entries(feed_id, is_read);
CREATE INDEX IF NOT EXISTS idx_entries_is_starred ON entries(is_starred);
CREATE INDEX IF NOT EXISTS idx_feeds_next_check ON feeds(next_check_at, is_active);
CREATE INDEX IF NOT EXISTS idx_feeds_url ON feeds(feed_url);
`

const schemaV2 = `
CREATE VIRTUAL TABLE IF NOT EXISTS entries_fts USING fts5(
    title,
    description,
    content,
    content='entries',
    content_rowid='id'
);

CREATE TRIGGER IF NOT EXISTS entries_ai AFTER INSERT ON entries BEGIN
    INSERT INTO entries_fts(rowid, title, description, content)
    VALUES (new.id, new.title, new.description, new.content);
END;

CREATE TRIGGER IF NOT EXISTS entries_ad AFTER DELETE ON entries BEGIN
    INSERT INTO entries_fts(entries_fts, rowid, title, description, content)
    VALUES ('delete', old.id, old.title, old.description, old.content);
END;

CREATE TRIGGER IF NOT EXISTS entries_au AFTER UPDATE ON entries BEGIN
    INSERT INTO entries_fts(entries_fts, rowid, title, description, content)
    VALUES ('delete', old.id, old.title, old.description, old.content);
    INSERT INTO entries_fts(rowid, title, description, content)
    VALUES (new.id, new.title, new.description, new.content);
END;
`

const schemaV3 = `
DROP INDEX IF EXISTS idx_entries_is_read;
DROP INDEX IF EXISTS idx_entries_is_starred;

ALTER TABLE entries RENAME TO entries_old;

CREATE TABLE entries (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id         INTEGER NOT NULL,
    title           TEXT NOT NULL DEFAULT '',
    url             TEXT NOT NULL,
    guid            TEXT NOT NULL,
    content         TEXT DEFAULT '',
    description     TEXT DEFAULT '',
    author_name     TEXT DEFAULT '',
    image_url       TEXT DEFAULT '',
    categories      TEXT DEFAULT '',
    published_at    TEXT,
    updated_at      TEXT DEFAULT (datetime('now')),
    hash            TEXT NOT NULL,
    created_at      TEXT DEFAULT (datetime('now')),
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

INSERT INTO entries
  (id, feed_id, title, url, guid, content, description, author_name,
   image_url, categories, published_at, updated_at, hash, created_at)
SELECT
  id, feed_id, title, url, guid, content, description, author_name,
  image_url, categories, published_at, updated_at, hash, created_at
FROM entries_old;

DROP TABLE entries_old;

CREATE UNIQUE INDEX IF NOT EXISTS idx_entries_hash ON entries(feed_id, hash);
CREATE INDEX IF NOT EXISTS idx_entries_feed_id ON entries(feed_id);
CREATE INDEX IF NOT EXISTS idx_entries_published_at ON entries(published_at DESC);
`
