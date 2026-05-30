package cmd

import (
	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/output"
)

var starCmd = &cobra.Command{
	Use:   "star <id>",
	Short: "Toggle star/bookmark on an entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseIDArg(args[0])
		if err != nil {
			return err
		}
		starred, err := entryRepo().ToggleStar(id)
		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}
		output.PrintSuccess(map[string]any{"id": id, "starred": starred}, nil)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(starCmd)
}
