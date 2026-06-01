package cmd

import (
	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/output"
	"github.com/0xfig521/tide/internal/repo"
)

var (
	ruleMatch    string
	ruleField    string
	ruleAction   string
	ruleValue    string
	rulePriority int
)

var ruleCmd = &cobra.Command{Use: "rule", Short: "Manage entry classification rules"}

var ruleAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new matching rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		if ruleMatch == "" {
			return output.PrintError(output.CodeInvalidArgs, "--match flag is required")
		}
		if ruleAction != "ignore" && ruleValue == "" {
			return output.PrintError(output.CodeInvalidArgs, "--value is required unless --action=ignore")
		}

		rule := repo.Rule{
			Priority:    rulePriority,
			IsActive:    true,
			MatchField:  ruleField,
			MatchRegex:  ruleMatch,
			Action:      ruleAction,
			ActionValue: ruleValue,
		}

		id, err := ruleRepo().Create(&rule)
		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}
		rule.ID = id
		output.PrintSuccess(rule, nil)
		return nil
	},
}

var ruleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all rules",
	RunE: func(cmd *cobra.Command, args []string) error {
		rules, err := ruleRepo().List()
		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}
		output.PrintSuccess(rules, nil)
		return nil
	},
}

var ruleRemoveCmd = &cobra.Command{
	Use:   "remove <id>",
	Short: "Remove a rule by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseIDArg(args[0])
		if err != nil {
			return err
		}
		if err := ruleRepo().Delete(id); err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}
		output.PrintSuccess(map[string]any{"id": id}, nil)
		return nil
	},
}

var ruleApplyCmd = &cobra.Command{
	Use:   "apply <entry-id>",
	Short: "Apply rules to a specific entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseIDArg(args[0])
		if err != nil {
			return err
		}

		entry, err := entryRepo().GetByID(id)
		if err != nil {
			return output.PrintError(output.CodeEntryNotFound, err.Error())
		}

		actions, err := ruleRepo().Apply(entry)
		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}

		for _, action := range []string{"state", "tag", "ignore", "priority", "category"} {
			if val, ok := actions[action]; ok {
				switch action {
				case "state", "ignore":
					stateRepo().SetState(entry.ID, val)
				case "tag":
					stateRepo().SetStateWithTags(entry.ID, "processed", val)
				}
			}
		}

		output.PrintSuccess(map[string]any{
			"entry_id": id,
			"actions":  actions,
		}, nil)
		return nil
	},
}

func init() {
	ruleAddCmd.Flags().StringVar(&ruleMatch, "match", "", "Regex pattern to match (required)")
	ruleAddCmd.Flags().StringVar(&ruleField, "field", "title", "Field to match: title, description, content, author, category")
	ruleAddCmd.Flags().StringVar(&ruleAction, "action", "tag", "Action: tag, state, priority, category, ignore")
	ruleAddCmd.Flags().StringVar(&ruleValue, "value", "", "Action value (required unless action=ignore)")
	ruleAddCmd.Flags().IntVar(&rulePriority, "priority", 0, "Rule priority (higher runs first)")
	ruleCmd.AddCommand(ruleAddCmd, ruleListCmd, ruleRemoveCmd, ruleApplyCmd)
	rootCmd.AddCommand(ruleCmd)
}
