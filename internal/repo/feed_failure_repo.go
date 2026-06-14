package repo

import (
	"database/sql"
	"strings"

	"github.com/0xfig-labs/tide/internal/db"
	"github.com/0xfig-labs/tide/internal/models"
)

// FeedFailureRepo handles persistence of feed fetch failures.
type FeedFailureRepo struct {
	db *db.DB
}

func NewFeedFailureRepo(database *db.DB) *FeedFailureRepo {
	return &FeedFailureRepo{db: database}
}

// FailingFeed is the joined view of a feed that has crossed the failure
// threshold, plus the metadata of its most recent failure.
type FailingFeed struct {
	FeedID              int64               `json:"feed_id"`
	Title               string              `json:"title"`
	FeedURL             string              `json:"feed_url"`
	IsActive            bool                `json:"is_active"`
	ConsecutiveFailures int                 `json:"consecutive_failures"`
	LastFetchedAt       string              `json:"last_fetched_at"`
	LastFailure         *models.FeedFailure `json:"last_failure"`
}

// Record appends a single failure row for a feed. The caller is expected
// to classify the error via fetcher.ClassifyError first.
func (r *FeedFailureRepo) Record(feedID int64, errorType models.FailureType, errorMessage string, httpStatus int) error {
	if feedID <= 0 {
		return nil
	}
	_, err := r.db.Conn.Exec(`
		INSERT INTO feed_failures (feed_id, error_type, error_message, http_status)
		VALUES (?, ?, ?, ?)
	`, feedID, string(errorType), errorMessage, httpStatus)
	return err
}

// ListFailingFeeds returns active feeds whose parsing_error_count has met
// or exceeded threshold, paired with their most recent failure row. If
// errorTypeFilter is non-empty, only feeds whose last failure matches
// the filter are returned.
func (r *FeedFailureRepo) ListFailingFeeds(threshold int, errorTypeFilter models.FailureType) ([]FailingFeed, error) {
	query := `
		SELECT
			f.id, f.title, f.feed_url, f.is_active,
			f.parsing_error_count, COALESCE(f.last_fetched_at, ''),
			ff.id, ff.error_type, ff.error_message, ff.http_status, ff.occurred_at
		FROM feeds f
		INNER JOIN feed_failures ff ON ff.id = (
			SELECT id FROM feed_failures
			WHERE feed_id = f.id
			ORDER BY occurred_at DESC, id DESC
			LIMIT 1
		)
		WHERE f.parsing_error_count >= ? AND f.is_active = 1`

	args := []any{threshold}
	if errorTypeFilter != "" {
		query += " AND ff.error_type = ?"
		args = append(args, string(errorTypeFilter))
	}
	query += " ORDER BY f.parsing_error_count DESC, ff.occurred_at DESC"

	rows, err := r.db.Conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []FailingFeed
	for rows.Next() {
		var (
			ff         FailingFeed
			failureID  int64
			errType    string
			errMsg     string
			httpStatus int
			occurredAt string
			isActive   int
		)
		if err := rows.Scan(
			&ff.FeedID, &ff.Title, &ff.FeedURL, &isActive,
			&ff.ConsecutiveFailures, &ff.LastFetchedAt,
			&failureID, &errType, &errMsg, &httpStatus, &occurredAt,
		); err != nil {
			return nil, err
		}
		ff.IsActive = isActive != 0
		ff.LastFailure = &models.FeedFailure{
			ID:           failureID,
			FeedID:       ff.FeedID,
			ErrorType:    models.FailureType(errType),
			ErrorMessage: errMsg,
			HTTPStatus:   httpStatus,
			OccurredAt:   occurredAt,
		}
		results = append(results, ff)
	}
	return results, rows.Err()
}

// GetHistory returns the most recent failure rows for a feed, newest first.
func (r *FeedFailureRepo) GetHistory(feedID int64, limit int) ([]models.FeedFailure, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.Conn.Query(`
		SELECT id, feed_id, error_type, error_message, http_status, occurred_at
		FROM feed_failures
		WHERE feed_id = ?
		ORDER BY occurred_at DESC, id DESC
		LIMIT ?
	`, feedID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.FeedFailure
	for rows.Next() {
		var f models.FeedFailure
		var errType string
		if err := rows.Scan(
			&f.ID, &f.FeedID, &errType, &f.ErrorMessage, &f.HTTPStatus, &f.OccurredAt,
		); err != nil {
			return nil, err
		}
		f.ErrorType = models.FailureType(errType)
		results = append(results, f)
	}
	return results, rows.Err()
}

// LastFailureForFeed returns the most recent failure row, or nil if none.
func (r *FeedFailureRepo) LastFailureForFeed(feedID int64) (*models.FeedFailure, error) {
	var f models.FeedFailure
	var errType string
	err := r.db.Conn.QueryRow(`
		SELECT id, feed_id, error_type, error_message, http_status, occurred_at
		FROM feed_failures
		WHERE feed_id = ?
		ORDER BY occurred_at DESC, id DESC
		LIMIT 1
	`, feedID).Scan(
		&f.ID, &f.FeedID, &errType, &f.ErrorMessage, &f.HTTPStatus, &f.OccurredAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	f.ErrorType = models.FailureType(errType)
	return &f, nil
}

// CountFailingFeeds returns the number of active feeds currently meeting
// the failure threshold. Used by tide failures list --summary.
func (r *FeedFailureRepo) CountFailingFeeds(threshold int) (int, error) {
	var n int
	err := r.db.Conn.QueryRow(`
		SELECT COUNT(*) FROM feeds
		WHERE parsing_error_count >= ? AND is_active = 1
	`, threshold).Scan(&n)
	return n, err
}

// DeleteForFeed removes all failure rows for a feed. CASCADE handles this
// automatically when the feed row itself is deleted; this method exists
// for the rare case of a partial cleanup (e.g. tide failures retry).
func (r *FeedFailureRepo) DeleteForFeed(feedID int64) (int64, error) {
	res, err := r.db.Conn.Exec(`DELETE FROM feed_failures WHERE feed_id = ?`, feedID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// FailureTypeList returns the canonical, comma-separated list of valid
// FailureType values, suitable for help text.
func FailureTypeList() string {
	values := make([]string, 0, len(models.ValidFailureTypes))
	for t := range models.ValidFailureTypes {
		values = append(values, string(t))
	}
	return strings.Join(values, ", ")
}
