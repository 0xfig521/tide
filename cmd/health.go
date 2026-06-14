package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xfig-labs/tide/internal/output"
)

var healthFormat string

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Show feed health status",
	Long:  "Show health statistics for all feeds with status classification (healthy, stale, failing, dead, unknown).\n\nOutput is JSONL by default — one JSON object per line, one per feed. Use --format json for a JSON envelope.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if healthFormat != "jsonl" && healthFormat != "json" {
			return output.PrintError(output.CodeInvalidArgs,
				fmt.Sprintf("invalid format: %q (must be jsonl or json)", healthFormat))
		}

		stats, err := feedRepo().GetHealthStats(0)
		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}

		switch healthFormat {
		case "jsonl":
			output.PrintJSONLItems(stats)
		default:
			output.PrintSuccess(stats, nil)
		}
		return nil
	},
}

func init() {
	healthCmd.Flags().StringVar(&healthFormat, "format", "jsonl", "Output format: jsonl (default), json")
	rootCmd.AddCommand(healthCmd)
}
