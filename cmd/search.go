package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/models"
	"github.com/0xfig521/tide/internal/output"
	"github.com/0xfig521/tide/internal/repo"
)

var (
	searchCategory string
	searchFeedID   int64
	searchUnread   bool
	searchStarred  bool
	searchLimit    int
)

var searchCmd = &cobra.Command{
	Use:   "search <keyword>",
	Short: "Search articles (alias for list --search)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		q := repo.EntryQuery{
			Keyword:      args[0],
			CategoryName: searchCategory,
			FeedID:       searchFeedID,
			UnreadOnly:   searchUnread,
			StarredOnly:  searchStarred,
			Page:         1,
			PageSize:     searchLimit,
		}

		entries, err := entryRepo().ListEntries(q)
		if err != nil {
			return output.PrintError(output.CodeInternalError, fmt.Sprintf("Search failed: %v", err))
		}

		total, _ := entryRepo().CountEntries(q)
		outputs := make([]models.EntryOutput, 0, len(entries))
		for _, e := range entries {
			outputs = append(outputs, entryToOutput(e))
		}
		output.PrintSuccess(map[string]any{
			"items": outputs, "total": total, "page": 1, "page_size": searchLimit,
		}, nil)
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVarP(&searchCategory, "category", "c", "", "Filter by category")
	searchCmd.Flags().Int64Var(&searchFeedID, "feed", 0, "Filter by feed ID")
	searchCmd.Flags().BoolVar(&searchUnread, "unread", false, "Only unread entries")
	searchCmd.Flags().BoolVar(&searchStarred, "starred", false, "Only starred entries")
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "n", 50, "Maximum results")
	rootCmd.AddCommand(searchCmd)
}

var _ = models.Entry{}
