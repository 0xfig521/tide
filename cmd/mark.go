package cmd

import (
	"github.com/spf13/cobra"

	"github.com/0xfig-labs/tide/internal/output"
)

var (
	markState string
	markTags  string
	markNote  string
)

var validStates = map[string]bool{
	"new":       true,
	"seen":      true,
	"processed": true,
	"ignored":   true,
	"failed":    true,
}

var markCmd = &cobra.Command{
	Use:   "mark <entry-id>",
	Short: "Set processing state on an entry",
	Long: `Set processing state on an entry for agent workflow tracking.

Valid states: new, seen, processed, ignored, failed.
Use --tag and --note for additional metadata.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseIDArg(args[0])
		if err != nil {
			return err
		}

		if !validStates[markState] {
			return output.PrintError(output.CodeInvalidArgs,
				"invalid state: "+markState+". Must be one of: new, seen, processed, ignored, failed")
		}

		entry, err := entryRepo().GetByID(id)
		if err != nil || entry == nil {
			return output.PrintError(output.CodeEntryNotFound, "entry not found")
		}

		switch {
		case markTags != "" && markNote != "":
			err = stateRepo().SetStateFull(id, markState, markTags, markNote)
		case markTags != "":
			err = stateRepo().SetStateWithTags(id, markState, markTags)
		case markNote != "":
			err = stateRepo().SetStateFull(id, markState, "", markNote)
		default:
			err = stateRepo().SetState(id, markState)
		}

		if err != nil {
			return output.PrintError(output.CodeInternalError, err.Error())
		}

		result := map[string]any{
			"entry_id": id,
			"state":    markState,
		}
		if markTags != "" {
			result["tags"] = markTags
		}
		if markNote != "" {
			result["note"] = markNote
		}

		output.PrintSuccess(result, nil)
		return nil
	},
}

func init() {
	markCmd.Flags().StringVar(&markState, "state", "", "Processing state (new, seen, processed, ignored, failed)")
	markCmd.Flags().StringVar(&markTags, "tag", "", "Comma-separated tags")
	markCmd.Flags().StringVar(&markNote, "note", "", "Optional note")
	markCmd.MarkFlagRequired("state")
	rootCmd.AddCommand(markCmd)
}
