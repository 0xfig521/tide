package repo

import (
	"database/sql"
	"sync"
	"time"

	"github.com/0xfig521/tide/internal/db"
	"github.com/0xfig521/tide/internal/models"
)

// FeedRepo handles feed data access.
type FeedRepo struct {
	db               *db.DB
	updateMetaOnce   sync.Once
	updateResultOnce sync.Once
	updateErrorOnce  sync.Once
	updateMetaStmt   *sql.Stmt
	updateResultStmt *sql.Stmt
	updateErrorStmt  *sql.Stmt
}

func NewFeedRepo(db *db.DB) *FeedRepo {
	return &FeedRepo{db: db}
}

func (r *FeedRepo) prepareUpdateMeta() error {
	var err error
	r.updateMetaOnce.Do(func() {
		r.updateMetaStmt, err = r.db.Conn.Prepare(`
			UPDATE feeds SET
				title = ?, description = ?, site_url = ?, image_url = ?,
				language = ?, feed_type = ?,
				updated_at = datetime('now')
			WHERE id = ?
		`)
	})
	return err
}

func (r *FeedRepo) prepareUpdateResult() error {
	var err error
	r.updateResultOnce.Do(func() {
		r.updateResultStmt, err = r.db.Conn.Prepare(`
			UPDATE feeds SET
				etag_header = ?, last_modified_header = ?,
				http_status_code = ?, last_fetched_at = ?, next_check_at = ?,
				checked_at = datetime('now'),
				parsing_error_count = 0, parsing_error_msg = '',
				updated_at = datetime('now')
			WHERE id = ?
		`)
	})
	return err
}

func (r *FeedRepo) prepareUpdateError() error {
	var err error
	r.updateErrorOnce.Do(func() {
		r.updateErrorStmt, err = r.db.Conn.Prepare(`
			UPDATE feeds SET
				parsing_error_count = parsing_error_count + 1,
				parsing_error_msg = ?,
				checked_at = datetime('now'),
				next_check_at = datetime('now', '+' || (1 << MIN(parsing_error_count, 6)) || ' hours'),
				updated_at = datetime('now')
			WHERE id = ?
		`)
	})
	return err
}

// Create inserts a new feed subscription.
func (r *FeedRepo) Create(feedURL string) (*models.Feed, error) {
	result, err := r.db.Conn.Exec(`
		INSERT INTO feeds (feed_url) VALUES (?)
	`, feedURL)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return r.GetByID(id)
}

// GetByID retrieves a feed by ID.
func (r *FeedRepo) GetByID(id int64) (*models.Feed, error) {
	f := &models.Feed{}
	var checkedAt, lastFetchedAt, nextCheckAt, createdAt, updatedAt string
	err := r.db.Conn.QueryRow(`
		SELECT id, title, description, site_url, feed_url, image_url,
			language, feed_type, etag_header, last_modified_header,
			COALESCE(checked_at,''), COALESCE(last_fetched_at,''),
			http_status_code, next_check_at, parsing_error_count,
			COALESCE(parsing_error_msg,''), is_active, created_at, updated_at
		FROM feeds WHERE id = ?
	`, id).Scan(
		&f.ID, &f.Title, &f.Description, &f.SiteURL, &f.FeedURL, &f.ImageURL,
		&f.Language, &f.FeedType, &f.ETagHeader, &f.LastModifiedHeader,
		&checkedAt, &lastFetchedAt,
		&f.HTTPStatusCode, &nextCheckAt, &f.ParsingErrorCount,
		&f.ParsingErrorMsg, &f.IsActive, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	f.CheckedAt = parseTime(checkedAt)
	f.LastFetchedAt = parseTime(lastFetchedAt)
	f.NextCheckAt = mustParseTime(nextCheckAt)
	f.CreatedAt = mustParseTime(createdAt)
	f.UpdatedAt = mustParseTime(updatedAt)

	return f, nil
}

// GetByURL retrieves a feed by its feed URL.
func (r *FeedRepo) GetByURL(feedURL string) (*models.Feed, error) {
	f := &models.Feed{}
	var checkedAt, lastFetchedAt, nextCheckAt, createdAt, updatedAt string
	err := r.db.Conn.QueryRow(`
		SELECT id, title, description, site_url, feed_url, image_url,
			language, feed_type, etag_header, last_modified_header,
			COALESCE(checked_at,''), COALESCE(last_fetched_at,''),
			http_status_code, next_check_at, parsing_error_count,
			COALESCE(parsing_error_msg,''), is_active, created_at, updated_at
		FROM feeds WHERE feed_url = ?
	`, feedURL).Scan(
		&f.ID, &f.Title, &f.Description, &f.SiteURL, &f.FeedURL, &f.ImageURL,
		&f.Language, &f.FeedType, &f.ETagHeader, &f.LastModifiedHeader,
		&checkedAt, &lastFetchedAt,
		&f.HTTPStatusCode, &nextCheckAt, &f.ParsingErrorCount,
		&f.ParsingErrorMsg, &f.IsActive, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	f.CheckedAt = parseTime(checkedAt)
	f.LastFetchedAt = parseTime(lastFetchedAt)
	f.NextCheckAt = mustParseTime(nextCheckAt)
	f.CreatedAt = mustParseTime(createdAt)
	f.UpdatedAt = mustParseTime(updatedAt)

	return f, nil
}

// List retrieves all feeds, optionally filtered by category.
func (r *FeedRepo) List(categoryName string) ([]*models.Feed, error) {
	var rows *sql.Rows
	var err error
	if categoryName != "" {
		rows, err = r.db.Conn.Query(`
			SELECT f.id, f.title, f.description, f.site_url, f.feed_url, f.image_url,
				f.language, f.feed_type,
				COALESCE(f.checked_at,''), COALESCE(f.last_fetched_at,''),
				f.http_status_code, f.next_check_at, f.parsing_error_count,
				COALESCE(f.parsing_error_msg,''), f.is_active, f.created_at, f.updated_at
			FROM feeds f
			INNER JOIN feed_categories fc ON fc.feed_id = f.id
			INNER JOIN categories c ON c.id = fc.category_id
			WHERE c.name = ?
			ORDER BY f.title
		`, categoryName)
	} else {
		rows, err = r.db.Conn.Query(`
			SELECT id, title, description, site_url, feed_url, image_url,
				language, feed_type,
				COALESCE(checked_at,''), COALESCE(last_fetched_at,''),
				http_status_code, next_check_at, parsing_error_count,
				COALESCE(parsing_error_msg,''), is_active, created_at, updated_at
			FROM feeds
			ORDER BY title
		`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []*models.Feed
	for rows.Next() {
		f := &models.Feed{}
		var checkedAt, lastFetchedAt, nextCheckAt, createdAt, updatedAt string
		if err := rows.Scan(
			&f.ID, &f.Title, &f.Description, &f.SiteURL, &f.FeedURL, &f.ImageURL,
			&f.Language, &f.FeedType,
			&checkedAt, &lastFetchedAt,
			&f.HTTPStatusCode, &nextCheckAt, &f.ParsingErrorCount,
			&f.ParsingErrorMsg, &f.IsActive, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		f.CheckedAt = parseTime(checkedAt)
		f.LastFetchedAt = parseTime(lastFetchedAt)
		f.NextCheckAt = mustParseTime(nextCheckAt)
		f.CreatedAt = mustParseTime(createdAt)
		f.UpdatedAt = mustParseTime(updatedAt)
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}

// UpdateMeta updates feed metadata after a successful fetch.
func (r *FeedRepo) UpdateMeta(id int64, title, description, siteURL, imageURL, language, feedType string) error {
	if err := r.prepareUpdateMeta(); err != nil {
		return err
	}
	_, err := r.updateMetaStmt.Exec(title, description, siteURL, imageURL, language, feedType, id)
	return err
}

// UpdateFetchResult updates fetch-related fields after a fetch attempt.
func (r *FeedRepo) UpdateFetchResult(id int64, etag, lastModified string, statusCode int, fetchedAt time.Time, nextCheckAt time.Time) error {
	if err := r.prepareUpdateResult(); err != nil {
		return err
	}
	_, err := r.updateResultStmt.Exec(etag, lastModified, statusCode,
		fetchedAt.Format("2006-01-02 15:04:05"),
		nextCheckAt.Format("2006-01-02 15:04:05"),
		id)
	return err
}

// UpdateFetchError records a fetch error and applies backoff.
func (r *FeedRepo) UpdateFetchError(id int64, errMsg string) error {
	if err := r.prepareUpdateError(); err != nil {
		return err
	}
	_, err := r.updateErrorStmt.Exec(errMsg, id)
	return err
}

// Delete removes a feed and all associated data (cascades to entries and feed_categories).
func (r *FeedRepo) Delete(id int64) error {
	_, err := r.db.Conn.Exec(`DELETE FROM feeds WHERE id = ?`, id)
	return err
}

// SetActive enables or disables a feed.
func (r *FeedRepo) SetActive(id int64, active bool) error {
	_, err := r.db.Conn.Exec(`UPDATE feeds SET is_active = ?, updated_at = datetime('now') WHERE id = ?`, active, id)
	return err
}

// GetDueFeeds returns feeds that are due for a check (next_check_at < now).
func (r *FeedRepo) GetDueFeeds(limit int) ([]*models.Feed, error) {
	rows, err := r.db.Conn.Query(`
		SELECT id, feed_url, etag_header, last_modified_header,
			COALESCE(last_fetched_at,''), parsing_error_count, is_active
		FROM feeds
		WHERE next_check_at < datetime('now') AND is_active = 1
		ORDER BY next_check_at ASC LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []*models.Feed
	for rows.Next() {
		f := &models.Feed{}
		var lastFetchedAt string
		if err := rows.Scan(
			&f.ID, &f.FeedURL, &f.ETagHeader, &f.LastModifiedHeader,
			&lastFetchedAt, &f.ParsingErrorCount, &f.IsActive,
		); err != nil {
			return nil, err
		}
		f.LastFetchedAt = parseTime(lastFetchedAt)
		feeds = append(feeds, f)
	}
	return feeds, rows.Err()
}

// AssignCategory links a feed to a category.
func (r *FeedRepo) AssignCategory(feedID int64, categoryID int64) error {
	_, err := r.db.Conn.Exec(`
		INSERT OR IGNORE INTO feed_categories (feed_id, category_id) VALUES (?, ?)
	`, feedID, categoryID)
	return err
}

// UnassignCategory removes a feed from a category.
func (r *FeedRepo) UnassignCategory(feedID int64, categoryID int64) error {
	_, err := r.db.Conn.Exec(`
		DELETE FROM feed_categories WHERE feed_id = ? AND category_id = ?
	`, feedID, categoryID)
	return err
}

// GetCategories returns category names for a feed.
func (r *FeedRepo) GetCategories(feedID int64) ([]string, error) {
	rows, err := r.db.Conn.Query(`
		SELECT c.name FROM categories c
		INNER JOIN feed_categories fc ON fc.category_id = c.id
		WHERE fc.feed_id = ?
		ORDER BY c.name
	`, feedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

// GetEntryCount returns the total entry count for a feed.
func (r *FeedRepo) GetEntryCount(feedID int64) (int, error) {
	var total int
	err := r.db.Conn.QueryRow(`
		SELECT COUNT(*) FROM entries WHERE feed_id = ?
	`, feedID).Scan(&total)
	return total, err
}

// FeedHealth represents feed health statistics for machine-readable status output.
type FeedHealth struct {
	FeedID              int64   `json:"feed_id"`
	Title               string  `json:"title"`
	FeedURL             string  `json:"feed_url"`
	Status              string  `json:"status"`
	LastFetchedAt       string  `json:"last_fetched_at"`
	LastSuccessAt       string  `json:"last_success_at"`
	ConsecutiveFailures int     `json:"consecutive_failures"`
	SuccessRate7d       float64 `json:"success_rate_7d"`
	Entries7d           int     `json:"entries_7d"`
	Entries30d          int     `json:"entries_30d"`
	StaleDays           int     `json:"stale_days"`
}

// GetHealthStats returns health statistics for all feeds, or a single feed if feedID > 0.
func (r *FeedRepo) GetHealthStats(feedID int64) ([]FeedHealth, error) {
	query := `
		SELECT
			f.id,
			COALESCE(f.title, ''),
			f.feed_url,
			COALESCE(f.last_fetched_at, ''),
			f.parsing_error_count,
			COALESCE(e.entries_7d, 0),
			COALESCE(e.entries_30d, 0)
		FROM feeds f
		LEFT JOIN (
			SELECT feed_id,
				COUNT(CASE WHEN published_at >= datetime('now', '-7 days') THEN 1 END) AS entries_7d,
				COUNT(CASE WHEN published_at >= datetime('now', '-30 days') THEN 1 END) AS entries_30d
			FROM entries
			GROUP BY feed_id
		) e ON f.id = e.feed_id`

	var rows *sql.Rows
	var err error
	if feedID > 0 {
		query += " WHERE f.id = ?"
		rows, err = r.db.Conn.Query(query, feedID)
	} else {
		query += " ORDER BY f.title"
		rows, err = r.db.Conn.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []FeedHealth
	for rows.Next() {
		var (
			id                int64
			title, feedURL    string
			lastFetchedStr    string
			parsingErrorCount int
			entries7d         int
			entries30d        int
		)
		if err := rows.Scan(&id, &title, &feedURL, &lastFetchedStr, &parsingErrorCount, &entries7d, &entries30d); err != nil {
			return nil, err
		}

		var staleDays int
		if lastFetchedStr != "" {
			t := parseTime(lastFetchedStr)
			if t != nil {
				staleDays = int(time.Since(*t).Hours() / 24)
			}
		}

		status := classifyHealth(staleDays, parsingErrorCount, lastFetchedStr)
		successRate := 0.0
		if lastFetchedStr != "" && parsingErrorCount == 0 {
			successRate = 1.0
		}

		results = append(results, FeedHealth{
			FeedID:              id,
			Title:               title,
			FeedURL:             feedURL,
			Status:              status,
			LastFetchedAt:       lastFetchedStr,
			LastSuccessAt:       lastFetchedStr, // last_fetched_at is only set on success
			ConsecutiveFailures: parsingErrorCount,
			SuccessRate7d:       successRate,
			Entries7d:           entries7d,
			Entries30d:          entries30d,
			StaleDays:           staleDays,
		})
	}
	return results, rows.Err()
}

// classifyHealth returns the health status string based on metrics.
func classifyHealth(staleDays, consecutiveFailures int, lastFetchedStr string) string {
	if lastFetchedStr == "" {
		return "unknown"
	}
	if consecutiveFailures >= 10 && staleDays > 30 {
		return "dead"
	}
	if consecutiveFailures >= 3 {
		return "failing"
	}
	if staleDays >= 7 {
		return "stale"
	}
	return "healthy"
}

// parseTime parses a time string, returning nil on empty/error.
func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return nil
	}
	return &t
}

// mustParseTime parses a time string, returning zero time on error.
func mustParseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return time.Time{}
	}
	return t
}
