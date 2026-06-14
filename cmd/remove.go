package cmd

import (
	"github.com/spf13/cobra"

	"github.com/0xfig-labs/tide/internal/output"
)

var removeCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove an RSS feed",
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

		if err := feedRepo().Delete(id); err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}

		output.PrintSuccess(map[string]any{"id": id, "title": f.Title, "feed_url": f.FeedURL}, nil)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
