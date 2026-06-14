package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/0xfig-labs/tide/internal/output"
)

// batchFeedInput represents one feed in the batch-add JSON input.
// URL is required; Category is optional.
type batchFeedInput struct {
	URL      string `json:"url"`
	Category string `json:"category"`
}

// batchAddResult represents the outcome of adding a single feed.
type batchAddResult struct {
	URL      string `json:"url"`
	Status   string `json:"status"`             // "imported", "skipped", "error"
	Reason   string `json:"reason,omitempty"`   // e.g. "already_exists", "invalid url"
	ID       int64  `json:"id,omitempty"`       // feed ID when imported
	Title    string `json:"title,omitempty"`    // feed title when imported
	Category string `json:"category,omitempty"` // assigned category
}

var (
	batchAddStrict bool
)

var batchAddCmd = &cobra.Command{
	Use:   "batch-add [file]",
	Short: "Subscribe to multiple RSS feeds from a JSON array",
	Long: `Add multiple RSS feeds from a JSON array via file or stdin.

Each element can be a plain URL string or an object with "url" and optional "category".
Categories are auto-created if they don't exist.
Duplicate feeds (already subscribed) are skipped.

Examples:
  tide batch-add feeds.json
  cat feeds.json | tide batch-add
  echo '["https://example.com/feed.xml", {"url":"https://other.com/rss","category":"tech"}]' | tide batch-add`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBatchAdd,
}

func init() {
	batchAddCmd.Flags().BoolVar(&batchAddStrict, "strict", false, "Fail entirely if any feed fails to add")
	rootCmd.AddCommand(batchAddCmd)
}

func runBatchAdd(cmd *cobra.Command, args []string) error {
	var rawJSON []byte
	var err error

	if len(args) == 1 {
		rawJSON, err = os.ReadFile(args[0])
		if err != nil {
			return output.PrintError(output.CodeInvalidArgs, "cannot read file: "+err.Error())
		}
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return output.PrintError(output.CodeInvalidArgs, "no file provided and stdin is a terminal; pipe JSON or provide a file path")
		}
		rawJSON, err = io.ReadAll(os.Stdin)
		if err != nil {
			return output.PrintError(output.CodeInternalError, "cannot read stdin: "+err.Error())
		}
	}

	if len(rawJSON) == 0 {
		return output.PrintError(output.CodeInvalidArgs, "empty input")
	}

	// Parse into a generic JSON array first, then decode each element.
	// This allows mixed types: strings and objects in the same array.
	var rawItems []json.RawMessage
	if err := json.Unmarshal(rawJSON, &rawItems); err != nil {
		return output.PrintError(output.CodeInvalidArgs, "invalid JSON: "+err.Error())
	}

	if len(rawItems) == 0 {
		return output.PrintError(output.CodeInvalidArgs, "empty feed list")
	}

	var (
		results  []batchAddResult
		imported int
		skipped  int
		errored  int
	)

	for _, raw := range rawItems {
		feed, isURLStr := parseBatchItem(raw)
		if feed == nil && !isURLStr {
			skipped++
			errored++
			results = append(results, batchAddResult{
				URL:    string(raw),
				Status: "error",
				Reason: "invalid input: must be a string URL or object with 'url' field",
			})
			continue
		}

		if feed.URL == "" {
			skipped++
			errored++
			results = append(results, batchAddResult{
				URL:    string(raw),
				Status: "error",
				Reason: "missing url",
			})
			continue
		}

		// Check for duplicates
		existing, _ := feedRepo().GetByURL(feed.URL)
		if existing != nil {
			skipped++
			results = append(results, batchAddResult{
				URL:    feed.URL,
				Status: "skipped",
				Reason: "already_exists",
				ID:     existing.ID,
				Title:  existing.Title,
			})
			continue
		}

		// Create the feed
		f, err := feedRepo().Create(feed.URL)
		if err != nil {
			skipped++
			errored++
			results = append(results, batchAddResult{
				URL:    feed.URL,
				Status: "error",
				Reason: err.Error(),
			})
			fmt.Fprintln(cmd.ErrOrStderr(), output.Warn(fmt.Sprintf("Failed to add %s: %v", feed.URL, err)))
			continue
		}

		// Assign category if specified
		if feed.Category != "" {
			cat, err := categoryRepo().GetByName(feed.Category)
			if err != nil {
				cat, err = categoryRepo().Create(feed.Category, "")
				if err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), output.Warn(fmt.Sprintf("Created feed %d but failed to assign category '%s': %v", f.ID, feed.Category, err)))
				}
			}
			if cat != nil {
				_ = feedRepo().AssignCategory(f.ID, cat.ID)
			}
		}

		imported++
		results = append(results, batchAddResult{
			URL:      feed.URL,
			Status:   "imported",
			ID:       f.ID,
			Title:    f.Title,
			Category: feed.Category,
		})
	}

	summary := map[string]any{
		"total":    len(rawItems),
		"imported": imported,
		"skipped":  skipped,
		"errored":  errored,
		"results":  results,
	}

	if batchAddStrict && errored > 0 {
		return output.PrintError(output.CodeInvalidArgs, fmt.Sprintf("%d feed(s) failed to add", errored))
	}

	output.PrintSuccess(summary, nil)
	return nil
}

// parseBatchItem parses a single raw JSON element from the input array.
// It handles both plain strings (URL only) and objects with "url" + optional "category".
// Returns (nil, false) if the item was an object but didn't have a recognizable "url" field.
func parseBatchItem(raw json.RawMessage) (feed *batchFeedInput, isPlainString bool) {
	// Try as a plain string first
	var urlStr string
	if err := json.Unmarshal(raw, &urlStr); err == nil {
		return &batchFeedInput{URL: urlStr}, true
	}

	// Try as an object
	var obj batchFeedInput
	if err := json.Unmarshal(raw, &obj); err == nil && obj.URL != "" {
		return &obj, false
	}

	// Try with lowercase field names (for flexibility)
	var altObj struct {
		URL      string `json:"Url"`
		Category string `json:"Category"`
	}
	if err := json.Unmarshal(raw, &altObj); err == nil && altObj.URL != "" {
		return &batchFeedInput{URL: altObj.URL, Category: altObj.Category}, false
	}

	return nil, false
}
