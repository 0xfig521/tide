package repo

import (
	"database/sql"
	"time"

	"github.com/0xfig521/tide/internal/db"
	"github.com/0xfig521/tide/internal/models"
)

const entryCols = `e.id, e.feed_id, e.title, e.url, e.guid, e.content, e.description,
	e.author_name, e.image_url, e.categories,
	COALESCE(e.published_at,''), e.created_at, e.updated_at,
	f.title as feed_title`

const entryFrom = `FROM entries e INNER JOIN feeds f ON f.id = e.feed_id`

type EntryRepo struct {
	db *db.DB
}

func NewEntryRepo(db *db.DB) *EntryRepo {
	return &EntryRepo{db: db}
}

func (r *EntryRepo) InsertOrIgnore(e *models.Entry) error {
	_, err := r.db.Conn.Exec(`
		INSERT OR IGNORE INTO entries
			(feed_id, title, url, guid, content, description, author_name,
			 image_url, categories, published_at, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, e.FeedID, e.Title, e.URL, e.GUID, e.Content, e.Description,
		e.AuthorName, e.ImageURL, e.Categories, formatTime(e.PublishedAt), e.Hash)
	return err
}

func (r *EntryRepo) ListByFeed(feedID int64, limit, offset int) ([]*models.Entry, error) {
	rows, err := r.db.Conn.Query(`
		SELECT `+entryCols+`
		`+entryFrom+`
		WHERE e.feed_id = ?
		ORDER BY e.published_at DESC LIMIT ? OFFSET ?
	`, feedID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEntries(rows)
}

// GetByID returns a single entry by its ID.
func (r *EntryRepo) GetByID(id int64) (*models.Entry, error) {
	row := r.db.Conn.QueryRow(`
		SELECT `+entryCols+`
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
	Page         int
	PageSize     int
}

// ListEntries is the unified entry listing with all filters + pagination + time range.
func (r *EntryRepo) ListEntries(q EntryQuery) ([]*models.Entry, error) {
	query := `SELECT ` + entryCols + ` ` + entryFrom
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

	for i, cond := range conditions {
		if i == 0 {
			query += " WHERE " + cond
		} else {
			query += " AND " + cond
		}
	}

	query += " ORDER BY e.published_at DESC"

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
	return scanEntries(rows)
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

func scanEntries(rows *sql.Rows) ([]*models.Entry, error) {
	var entries []*models.Entry
	for rows.Next() {
		e := &models.Entry{}
		var publishedAt, createdAt, updatedAt string
		if err := rows.Scan(
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
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func formatTime(t *time.Time) interface{} {
	if t == nil || t.IsZero() {
		return nil
	}
	return t.Format("2006-01-02 15:04:05")
}
