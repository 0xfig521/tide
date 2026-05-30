package cmd

import (
	"github.com/spf13/cobra"
)

var starCmd = &cobra.Command{
	Use:   "star <id>",
	Short: "Toggle star/bookmark on an entry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := parseIDArg(args[0])
		starred, err := entryRepo().ToggleStar(id)
		if err != nil {
			printJSON(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		printJSON(map[string]any{"ok": true, "id": id, "starred": starred})
	},
}

func init() {
	rootCmd.AddCommand(starCmd)
}
