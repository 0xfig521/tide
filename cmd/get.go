package cmd

import (
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/models"
	"github.com/0xfig521/tide/internal/output"
)

var (
	getFull        bool
	getContentOnly bool
	getMaxChars    int
	getTokenBudget int
	getText        bool
)

// getEntryOutput extends EntryOutput with truncation metadata
// used when --max-chars or --token-budget is active.
type getEntryOutput struct {
	models.EntryOutput
	Truncated       bool `json:"truncated"`
	CharCount       int  `json:"char_count"`
	EstimatedTokens int  `json:"estimated_tokens"`
}

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

func stripHTML(s string) string {
	result := htmlTagRe.ReplaceAllString(s, "")
	return strings.Join(strings.Fields(result), " ")
}

func estimateTokens(s string) int {
	return len(s) / 4
}

var getCmd = &cobra.Command{
	Use:   "get <entry-id>",
	Short: "Get full details of a single entry",
	Long: `Get all fields of a single entry by ID.

By default, outputs metadata and description without full content.
Use --full to include the content field, or --content-only to output
only content-related fields (title, url, description, content, author).

Content control flags:
  --text            Strip HTML tags, output plain text
  --max-chars N     Truncate text content to N characters
  --token-budget N  Truncate text to fit within estimated token budget

Examples:
  tide get 42                        # Metadata + description
  tide get 42 --full                 # Include full content
  tide get 42 --content-only         # Only content fields
  tide get 42 --text --max-chars 4000 # Plain text, 4k chars max`,
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

		content := entry.Content
		description := entry.Description

		if getText {
			content = stripHTML(content)
			description = stripHTML(description)
		}

		limit := 0
		if getMaxChars > 0 {
			limit = getMaxChars
		}
		if getTokenBudget > 0 {
			tokenChars := getTokenBudget * 4
			if limit == 0 || tokenChars < limit {
				limit = tokenChars
			}
		}

		truncEnabled := getMaxChars > 0 || getTokenBudget > 0
		truncated := false
		charCount := len(content)
		estTokens := estimateTokens(content)

		if limit > 0 && len(content) > limit {
			content = content[:limit]
			truncated = true
			charCount = len(content)
			estTokens = estimateTokens(content)
		}
		if limit > 0 && len(description) > limit {
			description = description[:limit]
			truncated = true
		}

		switch {
		case getContentOnly:
			m := map[string]any{
				"title":       entry.Title,
				"url":         entry.URL,
				"description": description,
				"content":     content,
				"author":      entry.AuthorName,
			}
			if truncEnabled {
				m["truncated"] = truncated
				m["char_count"] = charCount
				m["estimated_tokens"] = estTokens
			}
			output.PrintSuccess(m, nil)

		case getFull:
			full := entryToFullOutput(entry)
			full.Content = content
			full.Description = description
			if truncEnabled {
				output.PrintSuccess(getEntryOutput{
					EntryOutput:     full,
					Truncated:       truncated,
					CharCount:       charCount,
					EstimatedTokens: estTokens,
				}, nil)
			} else {
				output.PrintSuccess(full, nil)
			}

		default:
			summary := entryToSummaryOutput(entry)
			summary.Description = description
			summary.Content = ""
			if truncEnabled {
				output.PrintSuccess(getEntryOutput{
					EntryOutput:     summary,
					Truncated:       truncated,
					CharCount:       charCount,
					EstimatedTokens: estTokens,
				}, nil)
			} else {
				output.PrintSuccess(summary, nil)
			}
		}
		return nil
	},
}

func init() {
	getCmd.Flags().BoolVar(&getFull, "full", false, "Include full article content")
	getCmd.Flags().BoolVar(&getContentOnly, "content-only", false, "Output only content-related fields (title, url, description, content, author)")
	getCmd.Flags().IntVar(&getMaxChars, "max-chars", 0, "Truncate text content to N characters")
	getCmd.Flags().IntVar(&getTokenBudget, "token-budget", 0, "Truncate text content to fit within N estimated tokens")
	getCmd.Flags().BoolVar(&getText, "text", false, "Strip HTML tags and output plain text")
	rootCmd.AddCommand(getCmd)
}

// entryToSummaryOutput converts an entry to its JSON output form without full content.
// Includes metadata fields (id, feed_id, feed_title, published_at, guid, categories)
// plus description but excludes the content field.
func entryToSummaryOutput(e *models.Entry) models.EntryOutput {
	out := entryToFullOutput(e)
	out.Content = ""
	return out
}

// entryToContentOnlyOutput returns only content-related fields:
// title, url, description, content, author.
func entryToContentOnlyOutput(e *models.Entry) map[string]any {
	return map[string]any{
		"title":       e.Title,
		"url":         e.URL,
		"description": e.Description,
		"content":     e.Content,
		"author":      e.AuthorName,
	}
}
