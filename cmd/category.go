package cmd

import (
	"github.com/spf13/cobra"

	"github.com/0xfig-labs/tide/internal/models"
	"github.com/0xfig-labs/tide/internal/output"
)

var categoryDesc string

var categoryCmd = &cobra.Command{Use: "category", Short: "Manage feed categories"}

var categoryCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new category",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cat, err := categoryRepo().Create(args[0], categoryDesc)
		if err != nil {
			return output.PrintError(output.CodeAlreadyExists, err.Error())
		}
		output.PrintSuccess(cat, nil)
		return nil
	},
}

var categoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		cats, err := categoryRepo().List()
		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}
		output.PrintSuccess(cats, nil)
		return nil
	},
}

var categoryAssignCmd = &cobra.Command{
	Use:   "assign <feed-id> <category-name>",
	Short: "Assign a feed to a category",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		feedID, err := parseIDArg(args[0])
		if err != nil {
			return err
		}
		catName := args[1]

		cat, err := categoryRepo().GetByName(catName)
		if err != nil {
			cat, err = categoryRepo().Create(catName, "")
			if err != nil {
				return output.PrintError(output.CodeInternalError, err.Error())
			}
		}

		if err := feedRepo().AssignCategory(feedID, cat.ID); err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}
		output.PrintSuccess(map[string]any{"feed_id": feedID, "category": catName}, nil)
		return nil
	},
}

var categoryRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a category",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := categoryRepo().DeleteByName(args[0]); err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}
		output.PrintSuccess(map[string]any{"name": args[0]}, nil)
		return nil
	},
}

func init() {
	categoryCreateCmd.Flags().StringVar(&categoryDesc, "desc", "", "Category description")
	categoryCmd.AddCommand(categoryCreateCmd, categoryListCmd, categoryAssignCmd, categoryRemoveCmd)
	rootCmd.AddCommand(categoryCmd)
}

var _ = models.Category{}
