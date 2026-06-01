package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/opml"
	"github.com/0xfig521/tide/internal/output"
	"github.com/0xfig521/tide/internal/repo"
)

// exportEntry is the JSON-serializable view for entry export (includes hash and feed_url).
type exportEntry struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	FeedID      int64  `json:"feed_id"`
	FeedTitle   string `json:"feed_title"`
	FeedURL     string `json:"feed_url"`
	PublishedAt string `json:"published_at"`
	Hash        string `json:"hash"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

var (
	exportOutput string

	entriesFormat   string
	entriesSince    string
	entriesState    string
	entriesCategory string
	entriesLimit    int
	entriesOutput   string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export feeds to an OPML file",
	Long: `Export all RSS feed subscriptions as an OPML 2.0 file.
	
By default, the OPML XML is written to stdout.
Use --output to write to a file instead.`,
	RunE: runExport,
}

var entriesCmd = &cobra.Command{
	Use:   "entries",
	Short: "Export entries in JSONL or Markdown format",
	RunE:  runExportEntries,
}

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file path (default: stdout)")

	entriesCmd.Flags().StringVar(&entriesFormat, "format", "jsonl", "Output format: jsonl (default) or markdown")
	entriesCmd.Flags().StringVar(&entriesSince, "since", "", "Time range (1h, 6h, 12h, 24h, 3d, 7d, 14d, 30d)")
	entriesCmd.Flags().StringVar(&entriesState, "state", "", "Filter by processing state (new, seen, processed, ignored, failed)")
	entriesCmd.Flags().StringVar(&entriesCategory, "category", "", "Filter by category")
	entriesCmd.Flags().IntVar(&entriesLimit, "limit", 100, "Maximum entries to export")
	entriesCmd.Flags().StringVar(&entriesOutput, "output", "", "Output file path (default: stdout)")

	exportCmd.AddCommand(entriesCmd)
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	feeds, err := feedRepo().List("")
	if err != nil {
		return output.PrintError(output.CodeInternalError, err.Error())
	}

	// Group feeds by category
	type catKey struct {
		name string
	}
	catGroups := map[string][]opml.OpmlFeed{}
	var uncategorized []opml.OpmlFeed

	for _, f := range feeds {
		cats, _ := feedRepo().GetCategories(f.ID)

		of := opml.OpmlFeed{
			Title:   f.Title,
			XmlURL:  f.FeedURL,
			HtmlURL: f.SiteURL,
		}

		if len(cats) == 0 {
			uncategorized = append(uncategorized, of)
		} else {
			for _, catName := range cats {
				catGroups[catName] = append(catGroups[catName], of)
			}
		}
	}

	// Build ordered group list
	var groups []opml.FeedGroup
	for catName, catFeeds := range catGroups {
		groups = append(groups, opml.FeedGroup{Category: catName, Feeds: catFeeds})
	}
	if len(uncategorized) > 0 {
		groups = append(groups, opml.FeedGroup{Category: "", Feeds: uncategorized})
	}

	xmlData, err := opml.Generate("Tide Feeds", groups)
	if err != nil {
		return output.PrintError(output.CodeInternalError, err.Error())
	}

	if exportOutput != "" {
		if err := os.WriteFile(exportOutput, xmlData, 0644); err != nil {
			return output.PrintError(output.CodeInternalError, "cannot write file: "+err.Error())
		}
		output.PrintSuccess(map[string]string{"file": exportOutput, "feeds": fmt.Sprintf("%d", len(feeds))}, nil)
	} else {
		// Write OPML XML to stdout directly (not JSON-wrapped, by design)
		fmt.Print(string(xmlData))
	}

	return nil
}

func runExportEntries(cmd *cobra.Command, args []string) error {
	if entriesFormat != "jsonl" && entriesFormat != "markdown" {
		return output.PrintError(output.CodeInvalidArgs,
			fmt.Sprintf("invalid format: %q (must be jsonl or markdown)", entriesFormat))
	}

	// Build feed URL map for provenance fields
	feeds, err := feedRepo().List("")
	if err != nil {
		return output.PrintError(output.CodeInternalError, "failed to list feeds: "+err.Error())
	}
	feedURLs := make(map[int64]string, len(feeds))
	for _, f := range feeds {
		feedURLs[f.ID] = f.FeedURL
	}

	q := repo.EntryQuery{
		CategoryName: entriesCategory,
		Since:        sinceExpr(entriesSince),
		State:        entriesState,
		Page:         1,
		PageSize:     entriesLimit,
	}

	entries, err := entryRepo().ListEntries(q)
	if err != nil {
		return output.PrintError(output.CodeInternalError, "failed to list entries: "+err.Error())
	}

	// Convert to export struct with provenance
	exports := make([]exportEntry, 0, len(entries))
	for _, e := range entries {
		pubDate := ""
		if e.PublishedAt != nil {
			pubDate = e.PublishedAt.Format(time.RFC3339)
		}
		exports = append(exports, exportEntry{
			ID:          e.ID,
			Title:       e.Title,
			URL:         e.URL,
			FeedID:      e.FeedID,
			FeedTitle:   e.FeedTitle,
			FeedURL:     feedURLs[e.FeedID],
			PublishedAt: pubDate,
			Hash:        e.Hash,
			Description: e.Description,
			Content:     e.Content,
		})
	}

	if entriesFormat == "jsonl" {
		if entriesOutput != "" {
			var buf bytes.Buffer
			for _, e := range exports {
				b, err := json.Marshal(e)
				if err != nil {
					return output.PrintError(output.CodeInternalError, "failed to marshal entry: "+err.Error())
				}
				buf.Write(b)
				buf.WriteByte('\n')
			}
			if err := os.WriteFile(entriesOutput, buf.Bytes(), 0644); err != nil {
				return output.PrintError(output.CodeInternalError, "cannot write file: "+err.Error())
			}
			output.PrintSuccess(map[string]string{
				"file":    entriesOutput,
				"entries": fmt.Sprintf("%d", len(exports)),
			}, nil)
		} else {
			output.PrintJSONLItems(exports)
		}
		return nil
	}

	// Markdown format
	var buf bytes.Buffer
	for _, e := range exports {
		writeExportMarkdown(&buf, e)
	}

	if entriesOutput != "" {
		if err := os.WriteFile(entriesOutput, buf.Bytes(), 0644); err != nil {
			return output.PrintError(output.CodeInternalError, "cannot write file: "+err.Error())
		}
		output.PrintSuccess(map[string]string{
			"file":    entriesOutput,
			"entries": fmt.Sprintf("%d", len(exports)),
		}, nil)
	} else {
		fmt.Print(buf.String())
	}
	return nil
}

// writeExportMarkdown writes a single entry as a markdown block with YAML frontmatter.
func writeExportMarkdown(w io.Writer, e exportEntry) {
	fmt.Fprintf(w, "---\n")
	fmt.Fprintf(w, "id: %d\n", e.ID)
	fmt.Fprintf(w, "url: %s\n", e.URL)
	fmt.Fprintf(w, "feed: %s\n", e.FeedTitle)
	fmt.Fprintf(w, "published_at: %s\n", e.PublishedAt)
	fmt.Fprintf(w, "---\n\n")
	fmt.Fprintf(w, "# %s\n\n", e.Title)
	if e.Content != "" {
		fmt.Fprintf(w, "%s\n\n", e.Content)
	} else if e.Description != "" {
		fmt.Fprintf(w, "%s\n\n", e.Description)
	}
	fmt.Fprintf(w, "---\n")
}
