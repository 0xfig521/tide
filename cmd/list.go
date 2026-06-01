package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/models"
	"github.com/0xfig521/tide/internal/output"
	"github.com/0xfig521/tide/internal/repo"
)

var (
	listKeyword  string
	listCategory string
	listFeedID   int64
	listSince    string
	listState    string
	listFormat   string
	listPage     int
	listPageSize int
	listJSON     bool
)

var csvHeaders = []string{"id", "title", "url", "author", "published_at", "feed_id", "feed_title", "description", "categories", "guid"}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List articles with filtering and pagination",
	Long: `List articles. Default output is JSONL for AI agent consumption.

Use --format json for full JSON envelope or --format csv for CSV output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if listFormat != "jsonl" && listFormat != "json" && listFormat != "csv" {
			return output.PrintError(output.CodeInvalidArgs,
				fmt.Sprintf("invalid format: %q (must be jsonl, json, or csv)", listFormat))
		}

		q := repo.EntryQuery{
			Keyword:      listKeyword,
			CategoryName: listCategory,
			FeedID:       listFeedID,
			Since:        sinceExpr(listSince),
			State:        listState,
			Page:         listPage,
			PageSize:     listPageSize,
		}

		entries, err := entryRepo().ListEntries(q)
		if err != nil {
			return output.PrintError(output.CodeInternalError, fmt.Sprintf("List failed: %v", err))
		}

		switch listFormat {
		case "jsonl":
			outputs := make([]models.EntryOutput, 0, len(entries))
			for _, e := range entries {
				outputs = append(outputs, entryToFullOutput(e))
			}
			output.PrintJSONLItems(outputs)
			return nil

		case "json":
			total, _ := entryRepo().CountEntries(q)
			outputs := make([]models.EntryOutput, 0, len(entries))
			for _, e := range entries {
				outputs = append(outputs, entryToFullOutput(e))
			}
			output.PrintSuccess(map[string]any{
				"items":     outputs,
				"total":     total,
				"page":      q.Page,
				"page_size": q.PageSize,
			}, nil)
			return nil

		case "csv":
			rows := make([][]string, 0, len(entries))
			for _, e := range entries {
				rows = append(rows, entryToCSV(e))
			}
			output.PrintCSV(csvHeaders, rows)
			return nil
		}

		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listKeyword, "search", "", "Search keyword")
	listCmd.Flags().StringVarP(&listCategory, "category", "c", "", "Filter by category")
	listCmd.Flags().Int64Var(&listFeedID, "feed", 0, "Filter by feed ID")
	listCmd.Flags().StringVar(&listSince, "since", "", "Time range (1h, 6h, 12h, 24h, 3d, 7d, 14d, 30d)")
	listCmd.Flags().StringVar(&listState, "state", "", "Filter by processing state (new, seen, processed, ignored, failed)")
	listCmd.Flags().IntVarP(&listPage, "page", "p", 1, "Page number")
	listCmd.Flags().IntVar(&listPageSize, "page-size", 20, "Articles per page")
	listCmd.Flags().StringVar(&listFormat, "format", "jsonl", "Output format: jsonl (default), json, csv")
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

func entryToCSV(e *models.Entry) []string {
	pubDate := ""
	if e.PublishedAt != nil {
		pubDate = e.PublishedAt.Format("2006-01-02 15:04:05")
	}
	return []string{
		strconv.FormatInt(e.ID, 10),
		e.Title,
		e.URL,
		e.AuthorName,
		pubDate,
		strconv.FormatInt(e.FeedID, 10),
		e.FeedTitle,
		e.Description,
		e.Categories,
		e.GUID,
	}
}

func entryToFullOutput(e *models.Entry) models.EntryOutput {
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
	}
}
