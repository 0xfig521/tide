package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xfig-labs/tide/internal/models"
	"github.com/0xfig-labs/tide/internal/output"
	"github.com/0xfig-labs/tide/internal/repo"
)

var (
	failuresThreshold int
	failuresType      string
	failuresFormat    string
	failuresLimit     int
	failuresYes       bool
)

var failuresCmd = &cobra.Command{
	Use:   "failures",
	Short: "Manage feed sources that are persistently failing",
	Long: `Inspect, triage, and clear feed sources that have crossed the failure threshold.

A feed is classified as "failing" when parsing_error_count meets or exceeds the
threshold (default: 3 consecutive failures). Use this command to:
  - list    — show currently failing feeds with their last error
  - inspect — show the full failure history for a specific feed
  - clear   — hard-delete failing feeds (removes feed + entries + failure history)
  - retry   — reset error count and failure history so a feed retries immediately`,
}

var failuresListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show feeds currently in a failing state",
	Long: `List feeds whose parsing_error_count meets the failure threshold
(default: 3). Shows feed metadata plus the most recent failure reason.

Examples:
  tide failures list
  tide failures list --threshold 5
  tide failures list --type http_5xx
  tide failures list --format json
  tide failures list --threshold 10 --type timeout`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if failuresFormat != "jsonl" && failuresFormat != "json" {
			return output.PrintError(output.CodeInvalidArgs,
				fmt.Sprintf("invalid format: %q (must be jsonl or json)", failuresFormat))
		}

		if failuresThreshold < 1 {
			return output.PrintError(output.CodeInvalidArgs, "--threshold must be >= 1")
		}

		var typeFilter models.FailureType
		if failuresType != "" {
			typeFilter = models.FailureType(failuresType)
			if !models.ValidFailureTypes[typeFilter] {
				return output.PrintError(output.CodeInvalidArgs,
					fmt.Sprintf("invalid --type: %q. Valid types: %s", failuresType, repo.FailureTypeList()))
			}
		}

		feeds, err := failureRepo().ListFailingFeeds(failuresThreshold, typeFilter)
		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}

		if len(feeds) == 0 {
			switch failuresFormat {
			case "jsonl":
				return nil
			default:
				output.PrintSuccess([]repo.FailingFeed{}, nil)
				return nil
			}
		}

		switch failuresFormat {
		case "jsonl":
			output.PrintJSONLItems(feeds)
		default:
			output.PrintSuccess(feeds, nil)
		}
		return nil
	},
}

var failuresInspectCmd = &cobra.Command{
	Use:   "inspect <feed-id>",
	Short: "Show failure history for a single feed",
	Long: `Show the most recent failure records for a specific feed, newest first.
Includes each failure's type, HTTP status, error message, and timestamp.

Examples:
  tide failures inspect 5
  tide failures inspect 5 --limit 10`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseIDArg(args[0])
		if err != nil {
			return err
		}

		feed, err := feedRepo().GetByID(id)
		if err != nil {
			return output.PrintError(output.CodeFeedNotFound, "feed not found")
		}

		history, err := failureRepo().GetHistory(id, failuresLimit)
		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}

		result := map[string]any{
			"feed_id":              feed.ID,
			"title":                feed.Title,
			"feed_url":             feed.FeedURL,
			"consecutive_failures": feed.ParsingErrorCount,
			"last_error":           feed.ParsingErrorMsg,
			"failure_history":      history,
		}

		if failuresFormat == "jsonl" && len(history) > 0 {
			output.PrintJSONLItems(history)
			return nil
		}

		output.PrintSuccess(result, nil)
		return nil
	},
}

var failuresClearCmd = &cobra.Command{
	Use:   "clear [feed-id]",
	Short: "Hard-delete feeds that are persistently failing",
	Long: `Remove failing feeds from the database entirely (hard delete — cascade
removes all entries and failure history for each feed).

Without a feed ID, this removes ALL feeds whose parsing_error_count meets
the failure threshold (default: 3). Use --threshold to change the cutoff.

With a feed ID, that single feed is removed regardless of its error count.

This is a destructive, irreversible operation. The bulk form (no feed ID)
requires --yes to proceed.

Examples:
  tide failures clear --yes
  tide failures clear --threshold 5 --yes
  tide failures clear 42`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			// Single feed: no --yes required (same as tide remove).
			id, err := parseIDArg(args[0])
			if err != nil {
				return err
			}
			f, err := feedRepo().GetByID(id)
			if err != nil {
				return output.PrintError(output.CodeFeedNotFound, "feed not found")
			}
			if err := feedRepo().Delete(id); err != nil {
				return output.PrintError(output.CodeInternalError, err.Error())
			}
			output.PrintSuccess(map[string]any{
				"action":   "cleared",
				"id":       f.ID,
				"title":    f.Title,
				"feed_url": f.FeedURL,
			}, nil)
			return nil
		}

		if failuresThreshold < 1 {
			return output.PrintError(output.CodeInvalidArgs, "--threshold must be >= 1")
		}
		if !failuresYes {
			return output.PrintError(output.CodeInvalidArgs,
				"bulk clear is destructive and permanent. Pass --yes to confirm.")
		}

		// Bulk clear: list all failing feeds and delete each.
		feeds, err := failureRepo().ListFailingFeeds(failuresThreshold, "")
		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}

		if len(feeds) == 0 {
			output.PrintSuccess(map[string]any{
				"action":    "cleared",
				"feeds":     0,
				"message":   "no failing feeds found at or above threshold",
				"threshold": failuresThreshold,
			}, nil)
			return nil
		}

		type clearedFeed struct {
			ID      int64  `json:"id"`
			Title   string `json:"title"`
			FeedURL string `json:"feed_url"`
		}
		cleared := make([]clearedFeed, 0, len(feeds))
		for _, f := range feeds {
			if err := feedRepo().Delete(f.FeedID); err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), output.Warn(
					fmt.Sprintf("failed to delete feed %d (%s): %v", f.FeedID, f.Title, err)))
				continue
			}
			cleared = append(cleared, clearedFeed{ID: f.FeedID, Title: f.Title, FeedURL: f.FeedURL})
		}

		if failuresFormat == "jsonl" {
			output.PrintJSONLItems(cleared)
		} else {
			output.PrintSuccess(map[string]any{
				"action":    "cleared",
				"feeds":     len(cleared),
				"threshold": failuresThreshold,
				"cleared":   cleared,
			}, nil)
		}
		return nil
	},
}

var failuresRetryCmd = &cobra.Command{
	Use:   "retry <feed-id>",
	Short: "Reset error count and retry immediately",
	Long: `Reset a feed's parsing_error_count back to 0 and set its
next_check_at to now so the next tide fetch picks it up.

Also removes all stored failure rows for the feed so its
history starts fresh.

This is useful for transient failures you want to retry immediately
without waiting for the exponential backoff to expire.

Examples:
  tide failures retry 5`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseIDArg(args[0])
		if err != nil {
			return err
		}

		feed, err := feedRepo().GetByID(id)
		if err != nil {
			return output.PrintError(output.CodeFeedNotFound, "feed not found")
		}

		n, delErr := failureRepo().DeleteForFeed(id)
		if delErr != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), output.Warn(
				fmt.Sprintf("failed to clear failure history for feed %d: %v", id, delErr)))
		}

		// Reset counter + next_check_at so the next fetch picks it up.
		if _, err := dbConn.Conn.Exec(`
			UPDATE feeds SET
				parsing_error_count = 0,
				parsing_error_msg = '',
				next_check_at = datetime('now'),
				updated_at = datetime('now')
			WHERE id = ?
		`, id); err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}

		output.PrintSuccess(map[string]any{
			"action":                  "retry",
			"id":                      id,
			"title":                   feed.Title,
			"feed_url":                feed.FeedURL,
			"cleared_failure_records": n,
		}, nil)
		return nil
	},
}

func init() {
	failuresListCmd.Flags().IntVar(&failuresThreshold, "threshold", 3, "Failure threshold (minimum parsing_error_count)")
	failuresListCmd.Flags().StringVar(&failuresType, "type", "", fmt.Sprintf("Filter by failure type: %s", repo.FailureTypeList()))
	failuresListCmd.Flags().StringVar(&failuresFormat, "format", "jsonl", "Output format: jsonl (default), json")

	failuresInspectCmd.Flags().IntVar(&failuresLimit, "limit", 20, "Max failure history rows")
	failuresInspectCmd.Flags().StringVar(&failuresFormat, "format", "jsonl", "Output format: jsonl (default), json")

	failuresClearCmd.Flags().IntVar(&failuresThreshold, "threshold", 3, "Failure threshold for bulk clear")
	failuresClearCmd.Flags().BoolVarP(&failuresYes, "yes", "y", false, "Confirm bulk clear (required)")
	failuresClearCmd.Flags().StringVar(&failuresFormat, "format", "jsonl", "Output format: jsonl (default), json")

	failuresCmd.AddCommand(failuresListCmd)
	failuresCmd.AddCommand(failuresInspectCmd)
	failuresCmd.AddCommand(failuresClearCmd)
	failuresCmd.AddCommand(failuresRetryCmd)
	rootCmd.AddCommand(failuresCmd)
}
