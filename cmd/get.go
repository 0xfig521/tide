package cmd

import (
	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/output"
)

var getCmd = &cobra.Command{
	Use:   "get <entry-id>",
	Short: "Get full details of a single entry",
	Long: `Get all fields of a single entry by ID, including description and content.

Examples:
  tide get 42                    # Basic entry info
  tide get 42 --include content  # Full content included`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseIDArg(args[0])
		if err != nil {
			return err
		}

		entry, err := entryRepo().GetByID(id)
		if err != nil {
			return output.PrintError(output.CodeEntryNotFound, "entry not found")
		}

		output.PrintSuccess(entryToFullOutput(entry), nil)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
