package cmd

import (
	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/models"
	"github.com/0xfig521/tide/internal/output"
)

var sourcesCategory string

var sourcesCmd = &cobra.Command{
	Use:     "sources",
	Short:   "List all RSS feed subscriptions",
	Aliases: []string{"feeds"},
	RunE: func(cmd *cobra.Command, args []string) error {
		feeds, err := feedRepo().List(sourcesCategory)
		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}

		if len(feeds) == 0 {
			output.PrintSuccess([]models.FeedOutput{}, nil)
			return nil
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
		output.PrintSuccess(outputs, nil)
		return nil
	},
}

func init() {
	sourcesCmd.Flags().StringVarP(&sourcesCategory, "category", "c", "", "Filter by category name")
	rootCmd.AddCommand(sourcesCmd)
}
