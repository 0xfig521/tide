package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/output"
)

var addCategory string

var addCmd = &cobra.Command{
	Use:   "add <url>",
	Short: "Add an RSS feed URL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		feedURL := args[0]

		existing, _ := feedRepo().GetByURL(feedURL)
		if existing != nil {
			printJSON(map[string]any{
				"ok": false, "error": "already exists",
				"id": existing.ID, "title": existing.Title, "feed_url": existing.FeedURL,
			})
			return
		}

		f, err := feedRepo().Create(feedURL)
		if err != nil {
			printJSON(map[string]any{"ok": false, "error": err.Error()})
			return
		}

		if addCategory != "" {
			cat, err := categoryRepo().GetByName(addCategory)
			if err != nil {
				cat, err = categoryRepo().Create(addCategory, "")
				if err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), output.Warn(fmt.Sprintf("Created feed but failed to assign category: %v", err)))
				}
			}
			if cat != nil {
				feedRepo().AssignCategory(f.ID, cat.ID)
			}
		}

		printJSON(map[string]any{
			"ok": true, "id": f.ID, "feed_url": f.FeedURL, "title": f.Title,
		})
	},
}

func init() {
	addCmd.Flags().StringVarP(&addCategory, "category", "c", "", "Assign feed to a category")
	rootCmd.AddCommand(addCmd)
}
