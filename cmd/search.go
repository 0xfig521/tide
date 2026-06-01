package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/models"
	"github.com/0xfig521/tide/internal/output"
	"github.com/0xfig521/tide/internal/repo"
)

var (
	searchCategory string
	searchFeedID   int64
	searchLimit    int
	searchSince    string
	searchSort     string
	searchState    string
	searchFormat   string
)

var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search articles (FTS5 full-text search)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if searchFormat != "jsonl" && searchFormat != "json" && searchFormat != "csv" {
			return output.PrintError(output.CodeInvalidArgs,
				fmt.Sprintf("invalid format: %q (must be jsonl, json, or csv)", searchFormat))
		}
		if searchSort != "relevance" && searchSort != "published" {
			return output.PrintError(output.CodeInvalidArgs,
				fmt.Sprintf("invalid sort: %q (must be relevance or published)", searchSort))
		}

		q := repo.EntryQuery{
			Keyword:      args[0],
			CategoryName: searchCategory,
			FeedID:       searchFeedID,
			Since:        sinceExpr(searchSince),
			SortBy:       searchSort,
			State:        searchState,
			Page:         1,
			PageSize:     searchLimit,
		}

		entries, err := entryRepo().ListEntries(q)
		if err != nil {
			return output.PrintError(output.CodeInternalError, fmt.Sprintf("Search failed: %v", err))
		}

		switch searchFormat {
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
				"page":      1,
				"page_size": searchLimit,
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
	searchCmd.Flags().StringVarP(&searchCategory, "category", "c", "", "Filter by category")
	searchCmd.Flags().Int64Var(&searchFeedID, "feed", 0, "Filter by feed ID")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "n", 50, "Maximum results")
	searchCmd.Flags().StringVar(&searchSince, "since", "", "Time range (1h, 6h, 12h, 24h, 3d, 7d, 14d, 30d)")
	searchCmd.Flags().StringVar(&searchSort, "sort", "relevance", "Sort order: relevance or published")
	searchCmd.Flags().StringVar(&searchState, "state", "", "Filter by entry state: new, seen, processed, ignored, failed")
	searchCmd.Flags().StringVar(&searchFormat, "format", "jsonl", "Output format: jsonl (default), json, csv")
	rootCmd.AddCommand(searchCmd)
}

var _ = models.Entry{}
