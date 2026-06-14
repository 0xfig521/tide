package repo

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/0xfig-labs/tide/internal/db"
	"github.com/0xfig-labs/tide/internal/models"
)

// entryListCols is the column list for list/search operations — no content field.
// This avoids reading the large content blob when it's not needed, reducing I/O,
// memory, and output token costs for AI agent workflows.
const entryListCols = `e.id, e.feed_id, e.title, e.url, e.guid, e.description,
	e.author_name, e.image_url, e.categories,
	COALESCE(e.published_at,''), e.created_at, e.updated_at,
	f.title as feed_title`

// entryFullCols is the column list for get operations — includes content.
const entryFullCols = `e.id, e.feed_id, e.title, e.url, e.guid, e.content, e.description,
	e.author_name, e.image_url, e.categories,
	COALESCE(e.published_at,''), e.created_at, e.updated_at,
	f.title as feed_title`

const entryFrom = `FROM entries e INNER JOIN feeds f ON f.id = e.feed_id`

type EntryRepo struct {
	db            *db.DB
	insertOnce    sync.Once
	insertStmt    *sql.Stmt
	insertStmtErr error
}

func NewEntryRepo(db *db.DB) *EntryRepo {
	return &EntryRepo{db: db}
}

func (r *EntryRepo) prepareInsertStmt() error {
	r.insertOnce.Do(func() {
		r.insertStmt, r.insertStmtErr = r.db.Conn.Prepare(`
			INSERT OR IGNORE INTO entries
				(feed_id, title, url, guid, content, description, author_name,
				 image_url, categories, published_at, hash)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)
	})
	return r.insertStmtErr
}

func (r *EntryRepo) InsertOrIgnore(e *models.Entry) error {
	if err := r.prepareInsertStmt(); err != nil {
		return err
	}
	_, err := r.insertStmt.Exec(e.FeedID, e.Title, e.URL, e.GUID, e.Content,
		e.Description, e.AuthorName, e.ImageURL, e.Categories,
		formatTime(e.PublishedAt), e.Hash)
	return err
}

// BatchInsertEntries inserts multiple entries within a single transaction.
// Returns the count of new entries inserted. On error the transaction is rolled back.
// This is significantly faster than calling InsertOrIgnore in a loop as it
// avoids per-row transaction overhead and uses a single prepared statement.
func (r *EntryRepo) BatchInsertEntries(entries []*models.Entry) (int, error) {
	if len(entries) == 0 {
		return 0, nil
	}

	tx, err := r.db.Conn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO entries
			(feed_id, title, url, guid, content, description, author_name,
			 image_url, categories, published_at, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	count := 0
	for _, e := range entries {
		result, err := stmt.Exec(e.FeedID, e.Title, e.URL, e.GUID, e.Content,
			e.Description, e.AuthorName, e.ImageURL, e.Categories,
			formatTime(e.PublishedAt), e.Hash)
		if err != nil {
			return 0, err
		}
		if n, _ := result.RowsAffected(); n > 0 {
			count++
		}
	}

	return count, tx.Commit()
}

func (r *EntryRepo) ListByFeed(feedID int64, limit, offset int) ([]*models.Entry, error) {
	rows, err := r.db.Conn.Query(`
		SELECT `+entryListCols+`
		`+entryFrom+`
		WHERE e.feed_id = ?
		ORDER BY e.published_at DESC LIMIT ? OFFSET ?
	`, feedID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEntriesLight(rows)
}

// GetByHash returns a single entry by its dedup hash.
func (r *EntryRepo) GetByHash(hash string) (*models.Entry, error) {
	row := r.db.Conn.QueryRow(`
		SELECT `+entryFullCols+`
		`+entryFrom+`
		WHERE e.hash = ?
	`, hash)
	e := &models.Entry{}
	var publishedAt, createdAt, updatedAt string
	if err := row.Scan(
		&e.ID, &e.FeedID, &e.Title, &e.URL, &e.GUID, &e.Content, &e.Description,
		&e.AuthorName, &e.ImageURL, &e.Categories,
		&publishedAt, &createdAt, &updatedAt,
		&e.FeedTitle,
	); err != nil {
		return nil, err
	}
	e.PublishedAt = parseTime(publishedAt)
	e.CreatedAt = mustParseTime(createdAt)
	e.UpdatedAt = mustParseTime(updatedAt)
	return e, nil
}

// GetByID returns a single entry by its ID.
func (r *EntryRepo) GetByID(id int64) (*models.Entry, error) {
	row := r.db.Conn.QueryRow(`
		SELECT `+entryFullCols+`
		`+entryFrom+`
		WHERE e.id = ?
	`, id)
	e := &models.Entry{}
	var publishedAt, createdAt, updatedAt string
	if err := row.Scan(
		&e.ID, &e.FeedID, &e.Title, &e.URL, &e.GUID, &e.Content, &e.Description,
		&e.AuthorName, &e.ImageURL, &e.Categories,
		&publishedAt, &createdAt, &updatedAt,
		&e.FeedTitle,
	); err != nil {
		return nil, err
	}
	e.PublishedAt = parseTime(publishedAt)
	e.CreatedAt = mustParseTime(createdAt)
	e.UpdatedAt = mustParseTime(updatedAt)
	return e, nil
}

// EntryQuery holds all filter/pagination params for listing entries.
type EntryQuery struct {
	Keyword      string
	CategoryName string
	FeedID       int64
	Since        string // SQL time expression, e.g. "-24 hours", "-7 days"
	SortBy       string // "relevance" or "published" (default). relevance requires keyword.
	State        string // entry state filter: new, seen, processed, ignored, failed
	Page         int
	PageSize     int
}

// ListEntries is the unified entry listing with all filters + pagination + time range.
// Uses entryListCols (without content) for performance — full content requires GetByID.
func (r *EntryRepo) ListEntries(q EntryQuery) ([]*models.Entry, error) {
	query := `SELECT ` + entryListCols + ` ` + entryFrom
	var conditions []string
	var args []any

	if q.Keyword != "" {
		query += " INNER JOIN entries_fts ON entries_fts.rowid = e.id"
		if q.SortBy == "relevance" {
			query += ", bm25(entries_fts) as rank"
		}
		conditions = append(conditions, "entries_fts MATCH ?")
		args = append(args, q.Keyword)
	}

	if q.CategoryName != "" {
		query += " INNER JOIN feed_categories fc ON fc.feed_id = f.id INNER JOIN categories c ON c.id = fc.category_id"
		conditions = append(conditions, "c.name = ?")
		args = append(args, q.CategoryName)
	}

	if q.FeedID > 0 {
		conditions = append(conditions, "e.feed_id = ?")
		args = append(args, q.FeedID)
	}

	if q.Since != "" {
		conditions = append(conditions, "e.published_at >= datetime('now', ?)")
		args = append(args, q.Since)
	}

	if q.State != "" {
		query += " LEFT JOIN entry_states ON entry_states.entry_id = e.id"
		conditions = append(conditions, "entry_states.state = ?")
		args = append(args, q.State)
	}

	for i, cond := range conditions {
		if i == 0 {
			query += " WHERE " + cond
		} else {
			query += " AND " + cond
		}
	}

	if q.SortBy == "relevance" && q.Keyword != "" {
		query += " ORDER BY rank"
	} else {
		query += " ORDER BY e.published_at DESC"
	}

	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.Page <= 0 {
		q.Page = 1
	}
	offset := (q.Page - 1) * q.PageSize
	query += " LIMIT ? OFFSET ?"
	args = append(args, q.PageSize, offset)

	rows, err := r.db.Conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEntriesLight(rows)
}

// CountEntries returns the total count matching the query (for pagination info).
func (r *EntryRepo) CountEntries(q EntryQuery) (int, error) {
	query := `SELECT COUNT(*) ` + entryFrom
	var conditions []string
	var args []any

	if q.Keyword != "" {
		query += " INNER JOIN entries_fts ON entries_fts.rowid = e.id"
		conditions = append(conditions, "entries_fts MATCH ?")
		args = append(args, q.Keyword)
	}

	if q.CategoryName != "" {
		query += " INNER JOIN feed_categories fc ON fc.feed_id = f.id INNER JOIN categories c ON c.id = fc.category_id"
		conditions = append(conditions, "c.name = ?")
		args = append(args, q.CategoryName)
	}

	if q.FeedID > 0 {
		conditions = append(conditions, "e.feed_id = ?")
		args = append(args, q.FeedID)
	}

	if q.Since != "" {
		conditions = append(conditions, "e.published_at >= datetime('now', ?)")
		args = append(args, q.Since)
	}

	if q.State != "" {
		query += " LEFT JOIN entry_states ON entry_states.entry_id = e.id"
		conditions = append(conditions, "entry_states.state = ?")
		args = append(args, q.State)
	}

	for i, cond := range conditions {
		if i == 0 {
			query += " WHERE " + cond
		} else {
			query += " AND " + cond
		}
	}

	var count int
	err := r.db.Conn.QueryRow(query, args...).Scan(&count)
	return count, err
}

// DeleteOlderThan deletes entries older than the specified number of days (by created_at).
// Cascade rules automatically clean up entry_states and FTS index.
// Returns the number of deleted entries.
func (r *EntryRepo) DeleteOlderThan(days int) (int64, error) {
	since := fmt.Sprintf("-%d days", days)
	result, err := r.db.Conn.Exec(`
		DELETE FROM entries
		WHERE created_at < datetime('now', ?)
	`, since)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// scanEntriesLight scans 13 fields — no content (matches entryListCols).
func scanEntriesLight(rows *sql.Rows) ([]*models.Entry, error) {
	var entries []*models.Entry
	for rows.Next() {
		e := &models.Entry{}
		var publishedAt, createdAt, updatedAt string
		if err := rows.Scan(
			&e.ID, &e.FeedID, &e.Title, &e.URL, &e.GUID,
			&e.Description, &e.AuthorName, &e.ImageURL, &e.Categories,
			&publishedAt, &createdAt, &updatedAt,
			&e.FeedTitle,
		); err != nil {
			return nil, err
		}
		e.PublishedAt = parseTime(publishedAt)
		e.CreatedAt = mustParseTime(createdAt)
		e.UpdatedAt = mustParseTime(updatedAt)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func formatTime(t *time.Time) any {
	if t == nil || t.IsZero() {
		return nil
	}
	return t.Format("2006-01-02 15:04:05")
}
