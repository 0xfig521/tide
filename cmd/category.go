package cmd

import (
	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/models"
)

var categoryDesc string

var categoryCmd = &cobra.Command{Use: "category", Short: "Manage feed categories"}

var categoryCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new category",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cat, err := categoryRepo().Create(args[0], categoryDesc)
		if err != nil {
			printJSON(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		printJSON(cat)
	},
}

var categoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all categories",
	Run: func(cmd *cobra.Command, args []string) {
		cats, err := categoryRepo().List()
		if err != nil {
			printJSON(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		printJSON(cats)
	},
}

var categoryAssignCmd = &cobra.Command{
	Use:   "assign <feed-id> <category-name>",
	Short: "Assign a feed to a category",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		feedID := parseIDArg(args[0])
		catName := args[1]

		cat, err := categoryRepo().GetByName(catName)
		if err != nil {
			cat, err = categoryRepo().Create(catName, "")
			if err != nil {
				printJSON(map[string]any{"ok": false, "error": err.Error()})
				return
			}
		}

		if err := feedRepo().AssignCategory(feedID, cat.ID); err != nil {
			printJSON(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		printJSON(map[string]any{"ok": true, "feed_id": feedID, "category": catName})
	},
}

var categoryRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a category",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := categoryRepo().DeleteByName(args[0]); err != nil {
			printJSON(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		printJSON(map[string]any{"ok": true, "name": args[0]})
	},
}

func init() {
	categoryCreateCmd.Flags().StringVar(&categoryDesc, "desc", "", "Category description")
	categoryCmd.AddCommand(categoryCreateCmd, categoryListCmd, categoryAssignCmd, categoryRemoveCmd)
	rootCmd.AddCommand(categoryCmd)
}

var _ = models.Category{}
