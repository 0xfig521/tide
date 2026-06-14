package cmd

import (
	"github.com/spf13/cobra"

	"github.com/0xfig-labs/tide/internal/output"
)

var (
	pruneDays int
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Delete entries older than a retention period",
	Long: `Delete entries that are older than the specified number of days.

By default, entries older than 7 days are deleted. Use --days to customize
the retention period. Entry states are cleaned up automatically via cascade.

This command only deletes entries — it preserves feeds, categories, and rules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if pruneDays < 1 {
			return output.PrintError(output.CodeInvalidArgs, "--days must be at least 1")
		}

		er := entryRepo()
		deleted, err := er.DeleteOlderThan(pruneDays)
		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}

		output.PrintSuccess(map[string]any{
			"deleted":        deleted,
			"retention_days": pruneDays,
		}, nil)
		return nil
	},
}

func init() {
	pruneCmd.Flags().IntVarP(&pruneDays, "days", "d", 7, "Retention period in days (entries older than this are deleted)")
	rootCmd.AddCommand(pruneCmd)
}
