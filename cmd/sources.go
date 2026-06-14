package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/0xfig-labs/tide/internal/models"
	"github.com/0xfig-labs/tide/internal/output"
)

var (
	sourcesCategory string
	sourcesFormat   string
)

var sourcesCmd = &cobra.Command{
	Use:     "sources",
	Short:   "List all RSS feed subscriptions",
	Aliases: []string{"feeds"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if sourcesFormat != "jsonl" && sourcesFormat != "json" && sourcesFormat != "csv" {
			return output.PrintError(output.CodeInvalidArgs,
				fmt.Sprintf("invalid format: %q (must be jsonl, json, or csv)", sourcesFormat))
		}

		feeds, err := feedRepo().List(sourcesCategory)
		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}

		if len(feeds) == 0 {
			switch sourcesFormat {
			case "jsonl":
				return nil
			case "csv":
				output.PrintCSV(sourceCSVHeaders, nil)
				return nil
			default:
				output.PrintSuccess([]models.FeedOutput{}, nil)
				return nil
			}
		}

		outputs := make([]models.FeedOutput, 0, len(feeds))
		for _, f := range feeds {
			cats, _ := feedRepo().GetCategories(f.ID)
			total, _ := feedRepo().GetEntryCount(f.ID)
			lastFetched := ""
			if f.LastFetchedAt != nil {
				lastFetched = f.LastFetchedAt.Format("2006-01-02 15:04:05")
			}
			outputs = append(outputs, models.FeedOutput{
				ID: f.ID, Title: f.Title, FeedURL: f.FeedURL,
				SiteURL: f.SiteURL, Description: f.Description,
				ImageURL: f.ImageURL, Categories: cats,
				EntryCount:  total,
				LastFetched: lastFetched, IsActive: f.IsActive,
			})
		}

		switch sourcesFormat {
		case "jsonl":
			output.PrintJSONLItems(outputs)
		case "csv":
			rows := make([][]string, 0, len(outputs))
			for _, f := range outputs {
				rows = append(rows, feedToCSV(f))
			}
			output.PrintCSV(sourceCSVHeaders, rows)
		default:
			output.PrintSuccess(outputs, nil)
		}
		return nil
	},
}

var sourceCSVHeaders = []string{"id", "title", "feed_url", "site_url", "description", "categories", "entry_count", "last_fetched", "is_active"}

func feedToCSV(f models.FeedOutput) []string {
	cats := ""
	for i, c := range f.Categories {
		if i > 0 {
			cats += "; "
		}
		cats += c
	}
	return []string{
		strconv.FormatInt(f.ID, 10),
		f.Title,
		f.FeedURL,
		f.SiteURL,
		f.Description,
		cats,
		strconv.Itoa(f.EntryCount),
		f.LastFetched,
		strconv.FormatBool(f.IsActive),
	}
}

func init() {
	sourcesCmd.Flags().StringVarP(&sourcesCategory, "category", "c", "", "Filter by category name")
	sourcesCmd.Flags().StringVar(&sourcesFormat, "format", "jsonl", "Output format: jsonl (default), json, csv")
	rootCmd.AddCommand(sourcesCmd)
}
