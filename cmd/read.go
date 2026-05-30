package cmd

import (
	"github.com/spf13/cobra"
)

var readCmd = &cobra.Command{
	Use:   "read <id>",
	Short: "Mark an entry as read",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		id := parseIDArg(args[0])
		if err := entryRepo().MarkRead(id); err != nil {
			printJSON(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		printJSON(map[string]any{"ok": true, "id": id, "read": true})
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
}
