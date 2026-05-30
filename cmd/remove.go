package cmd

import (
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove an RSS feed",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := parseIDArg(args[0])

		f, err := feedRepo().GetByID(id)
		if err != nil {
			printJSON(map[string]any{"ok": false, "error": "feed not found"})
			return
		}

		if err := feedRepo().Delete(id); err != nil {
			printJSON(map[string]any{"ok": false, "error": err.Error()})
			return
		}

		printJSON(map[string]any{"ok": true, "id": id, "title": f.Title, "feed_url": f.FeedURL})
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
