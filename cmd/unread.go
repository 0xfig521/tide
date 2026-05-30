package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/models"
	"github.com/0xfig521/tide/internal/repo"
)

var (
	unreadCategory string
	unreadLimit    int
)

var unreadCmd = &cobra.Command{
	Use:   "unread",
	Short: "List unread articles (alias for list --unread)",
	Run: func(cmd *cobra.Command, args []string) {
		q := repo.EntryQuery{
			CategoryName: unreadCategory,
			UnreadOnly:   true,
			Page:         1,
			PageSize:     unreadLimit,
		}

		entries, err := entryRepo().ListEntries(q)
		if err != nil {
			fatal(fmt.Sprintf("Unread list failed: %v", err))
			return
		}

		total, _ := entryRepo().CountEntries(q)
		outputs := make([]models.EntryOutput, 0, len(entries))
		for _, e := range entries {
			outputs = append(outputs, entryToOutput(e))
		}
		printJSON(map[string]any{
			"items": outputs, "total": total, "page": 1, "page_size": unreadLimit,
		})
	},
}

func init() {
	unreadCmd.Flags().StringVarP(&unreadCategory, "category", "c", "", "Filter by category")
	unreadCmd.Flags().IntVarP(&unreadLimit, "limit", "n", 50, "Maximum entries")
	rootCmd.AddCommand(unreadCmd)
}

var _ = models.Entry{}
