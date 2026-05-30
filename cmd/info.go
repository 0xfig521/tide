package cmd

import (
	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/models"
)

var infoCmd = &cobra.Command{
	Use:   "info <id>",
	Short: "Show feed details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := parseIDArg(args[0])

		f, err := feedRepo().GetByID(id)
		if err != nil {
			printJSON(map[string]any{"ok": false, "error": "feed not found"})
			return
		}

		cats, _ := feedRepo().GetCategories(f.ID)
		total, unread, _ := feedRepo().GetEntryCount(f.ID)

		lastFetched := ""
		if f.LastFetchedAt != nil {
			lastFetched = f.LastFetchedAt.Format("2006-01-02 15:04:05")
		}

		printJSON(models.FeedOutput{
			ID: f.ID, Title: f.Title, FeedURL: f.FeedURL,
			SiteURL: f.SiteURL, Description: f.Description,
			ImageURL: f.ImageURL, Categories: cats,
			EntryCount: total, UnreadCount: unread,
			LastFetched: lastFetched, IsActive: f.IsActive,
		})
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
