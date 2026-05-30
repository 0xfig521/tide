package cmd

import (
	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/models"
	"github.com/0xfig521/tide/internal/output"
)

var infoCmd = &cobra.Command{
	Use:   "info <id>",
	Short: "Show feed details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseIDArg(args[0])
		if err != nil {
			return err
		}

		f, err := feedRepo().GetByID(id)
		if err != nil {
			return output.PrintError(output.CodeFeedNotFound, "feed not found")
		}

		cats, _ := feedRepo().GetCategories(f.ID)
		total, unread, _ := feedRepo().GetEntryCount(f.ID)

		lastFetched := ""
		if f.LastFetchedAt != nil {
			lastFetched = f.LastFetchedAt.Format("2006-01-02 15:04:05")
		}

		output.PrintSuccess(models.FeedOutput{
			ID: f.ID, Title: f.Title, FeedURL: f.FeedURL,
			SiteURL: f.SiteURL, Description: f.Description,
			ImageURL: f.ImageURL, Categories: cats,
			EntryCount: total, UnreadCount: unread,
			LastFetched: lastFetched, IsActive: f.IsActive,
		}, nil)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
