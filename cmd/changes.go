package cmd

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/models"
	"github.com/0xfig521/tide/internal/output"
)

var (
	changesAfter string
	changesLimit int
)

var changesCmd = &cobra.Command{
	Use:   "changes",
	Short: "Get new/changed entries since last call",
	Long: `Get entries created since a given cursor position.

Without --after, returns entries since the last time changes was called
(derived from the change_log table).

With --after <cursor>, returns entries created after that cursor.
Returns a new cursor for the next call.

Default limit is 50. Use --limit to adjust.

Examples:
  tide changes                  # entries since last call
  tide changes --limit 20       # at most 20 entries
  tide changes --after "2026-06-01 10:00:00:entry_42"  # resume from cursor`,
	RunE: runChanges,
}

func init() {
	changesCmd.Flags().StringVar(&changesAfter, "after", "", "Cursor position to fetch changes after (format: YYYY-MM-DD HH:MM:SS:entry_N)")
	changesCmd.Flags().IntVar(&changesLimit, "limit", 50, "Maximum entries to return")
	rootCmd.AddCommand(changesCmd)
}

func runChanges(cmd *cobra.Command, args []string) error {
	cursor := changesAfter
	if cursor == "" {
		// Derive default cursor from last change_log entry_created record.
		var lastCreated sql.NullString
		err := dbConn.Conn.QueryRow(
			`SELECT created_at FROM change_log
			 WHERE event_type = 'entry_created'
			 ORDER BY created_at DESC LIMIT 1`,
		).Scan(&lastCreated)
		if err != nil && err != sql.ErrNoRows {
			return output.PrintError(output.CodeInternalError,
				fmt.Sprintf("cursor lookup failed: %v", err))
		}
		if lastCreated.Valid {
			cursor = lastCreated.String
		}
	}

	// Extract the timestamp portion from the cursor string.
	// Cursor format: "YYYY-MM-DD HH:MM:SS" (from change_log) or
	//                "YYYY-MM-DD HH:MM:SS:entry_N" (from --after).
	cursorTS := cursor
	if idx := strings.Index(cursor, ":entry_"); idx > 0 {
		cursorTS = cursor[:idx]
	}

	// Query entries created since the cursor timestamp, ordered ascending.
	rows, err := dbConn.Conn.Query(
		`SELECT e.id, e.feed_id, e.title, e.url, e.guid, e.content, e.description,
		        e.author_name, e.image_url, e.categories,
		        COALESCE(e.published_at,''), e.created_at, e.updated_at,
		        COALESCE(f.title,'') AS feed_title
		 FROM entries e
		 LEFT JOIN feeds f ON e.feed_id = f.id
		 WHERE e.created_at > ?
		 ORDER BY e.created_at ASC, e.id ASC
		 LIMIT ?`,
		cursorTS, changesLimit,
	)
	if err != nil {
		return output.PrintError(output.CodeInternalError,
			fmt.Sprintf("query failed: %v", err))
	}
	defer rows.Close()

	var entries []*models.Entry
	for rows.Next() {
		e, err := scanEntryRow(rows)
		if err != nil {
			return output.PrintError(output.CodeInternalError,
				fmt.Sprintf("scan failed: %v", err))
		}
		entries = append(entries, e)

		// Record that this entry was served via changes.
		// Use the entry's own created_at as the change_log timestamp so that
		// MAX(created_at) FROM change_log stays consistent with entry.created_at.
		_, err = dbConn.Conn.Exec(
			`INSERT INTO change_log (event_type, entity_id, created_at, details)
			 VALUES ('entry_created', ?, ?, '')`,
			e.ID, e.CreatedAt.Format("2006-01-02 15:04:05"),
		)
		if err != nil {
			return output.PrintError(output.CodeInternalError,
				fmt.Sprintf("change_log insert failed: %v", err))
		}
	}
	if err := rows.Err(); err != nil {
		return output.PrintError(output.CodeInternalError,
			fmt.Sprintf("rows iteration failed: %v", err))
	}

	// Build the next cursor from the last returned entry.
	newCursor := ""
	if len(entries) > 0 {
		last := entries[len(entries)-1]
		newCursor = fmt.Sprintf("%s:entry_%d",
			last.CreatedAt.Format("2006-01-02 15:04:05"), last.ID)
	}

	outputs := make([]models.EntryOutput, 0, len(entries))
	for _, e := range entries {
		outputs = append(outputs, entryToFullOutput(e))
	}

	output.PrintSuccess(map[string]any{
		"cursor": newCursor,
		"items":  outputs,
		"count":  len(outputs),
	}, nil)
	return nil
}

// scanEntryRow scans a single entry row from the database.
// Follows the same pattern as repo.scanEntries but inlined for direct-DB access.
func scanEntryRow(rows *sql.Rows) (*models.Entry, error) {
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
	e.PublishedAt = parseTimeStr(publishedAt)
	e.CreatedAt = mustParseTimeStr(createdAt)
	e.UpdatedAt = mustParseTimeStr(updatedAt)
	return e, nil
}

// parseTimeStr parses a time string, returning nil on empty or error.
func parseTimeStr(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return nil
	}
	return &t
}

// mustParseTimeStr parses a time string, returning zero time on error.
func mustParseTimeStr(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return time.Time{}
	}
	return t
}
