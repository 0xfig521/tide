package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/models"
	"github.com/0xfig521/tide/internal/output"
	"github.com/0xfig521/tide/internal/repo"
)

var (
	listKeyword  string
	listCategory string
	listFeedID   int64
	listUnread   bool
	listStarred  bool
	listSince    string
	listPage     int
	listPageSize int
	listFormat   string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List articles with filtering and pagination",
	Long: `List articles. Default output is JSON.

Examples:
  tide list                          # All articles, page 1, 20 per page
  tide list --unread                 # Unread articles
  tide list --unread --since 24h     # Unread from last 24 hours
  tide list --search golang          # Search by keyword
  tide list --category tech --page 2 # Category filtered, page 2
  tide list --page-size 50           # 50 per page
  tide list --format table           # Terminal table view`,
	RunE: func(cmd *cobra.Command, args []string) error {
		q := repo.EntryQuery{
			Keyword:      listKeyword,
			CategoryName: listCategory,
			FeedID:       listFeedID,
			UnreadOnly:   listUnread,
			StarredOnly:  listStarred,
			Since:        sinceExpr(listSince),
			Page:         listPage,
			PageSize:     listPageSize,
		}

		entries, err := entryRepo().ListEntries(q)
		if err != nil {
			return output.PrintError(output.CodeInternalError, fmt.Sprintf("List failed: %v", err))
		}

		if listFormat == "table" {
			if len(entries) == 0 {
				output.PrintTable(output.Warn("No articles found."))
				return nil
			}
			headers := []string{"ID", "Title", "Feed", "Date", "★"}
			var rows [][]string
			for _, e := range entries {
				star := ""
				if e.IsStarred {
					star = "★"
				}
				pubDate := ""
				if e.PublishedAt != nil {
					pubDate = e.PublishedAt.Format("01-02 15:04")
				}
				rows = append(rows, []string{
					fmt.Sprintf("%d", e.ID),
					truncate(e.Title, 50),
					truncate(e.FeedTitle, 20),
					pubDate,
					star,
				})
			}
			output.PrintTable(output.EntryTable(headers, rows))
			return nil
		}

		total, _ := entryRepo().CountEntries(q)
		outputs := make([]models.EntryOutput, 0, len(entries))
		for _, e := range entries {
			outputs = append(outputs, entryToOutput(e))
		}
		output.PrintSuccess(map[string]any{
			"items":     outputs,
			"total":     total,
			"page":      q.Page,
			"page_size": q.PageSize,
		}, nil)
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listKeyword, "search", "", "Search keyword")
	listCmd.Flags().StringVarP(&listCategory, "category", "c", "", "Filter by category")
	listCmd.Flags().Int64Var(&listFeedID, "feed", 0, "Filter by feed ID")
	listCmd.Flags().BoolVarP(&listUnread, "unread", "u", false, "Only unread articles")
	listCmd.Flags().BoolVar(&listStarred, "starred", false, "Only starred articles")
	listCmd.Flags().StringVar(&listSince, "since", "", "Time range (1h, 6h, 12h, 24h, 3d, 7d, 14d, 30d)")
	listCmd.Flags().IntVarP(&listPage, "page", "p", 1, "Page number")
	listCmd.Flags().IntVar(&listPageSize, "page-size", 20, "Articles per page")
	listCmd.Flags().StringVar(&listFormat, "format", "", "Output format: 'table' (default: JSON)")
	rootCmd.AddCommand(listCmd)
}

func sinceExpr(s string) string {
	switch s {
	case "1h":
		return "-1 hours"
	case "6h":
		return "-6 hours"
	case "12h":
		return "-12 hours"
	case "24h":
		return "-24 hours"
	case "3d":
		return "-3 days"
	case "7d":
		return "-7 days"
	case "14d":
		return "-14 days"
	case "30d":
		return "-30 days"
	default:
		return ""
	}
}

func entryToOutput(e *models.Entry) models.EntryOutput {
	pubDate := ""
	if e.PublishedAt != nil {
		pubDate = e.PublishedAt.Format("2006-01-02 15:04:05")
	}
	return models.EntryOutput{
		ID: e.ID, Title: e.Title, URL: e.URL,
		Author: e.AuthorName, PublishedAt: pubDate,
		FeedTitle: e.FeedTitle, FeedID: e.FeedID,
		Description: e.Description, Content: e.Content,
		Categories: e.Categories, GUID: e.GUID,
		IsRead: e.IsRead, IsStarred: e.IsStarred,
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
