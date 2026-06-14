package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xfig-labs/tide/internal/output"
)

var addCategory string

var addCmd = &cobra.Command{
	Use:   "add <url>",
	Short: "Add an RSS feed URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		feedURL := args[0]

		existing, _ := feedRepo().GetByURL(feedURL)
		if existing != nil {
			return output.PrintError(output.CodeFeedAlreadyExists, "already exists")
		}

		f, err := feedRepo().Create(feedURL)
		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
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

		output.PrintSuccess(map[string]any{
			"id": f.ID, "feed_url": f.FeedURL, "title": f.Title,
		}, nil)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addCategory, "category", "c", "", "Assign feed to a category")
	rootCmd.AddCommand(addCmd)
}
