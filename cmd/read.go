package cmd

import (
	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/output"
)

var readCmd = &cobra.Command{
	Use:   "read <id>",
	Short: "Mark an entry as read",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseIDArg(args[0])
		if err != nil {
			return err
		}
		if err := entryRepo().MarkRead(id); err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}
		output.PrintSuccess(map[string]any{"id": id, "read": true}, nil)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
}
