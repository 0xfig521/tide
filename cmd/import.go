package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/opml"
	"github.com/0xfig521/tide/internal/output"
)

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import feeds from an OPML file",
	Long: `Import RSS feed subscriptions from an OPML 2.0 file.
	
Feeds are added to the database along with their category structure.
Duplicates (feeds already subscribed) are skipped.
Results are returned as a JSON summary.`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	data, err := os.ReadFile(filePath)
	if err != nil {
		return output.PrintError(output.CodeInvalidArgs, "cannot read file: "+err.Error())
	}

	feeds, err := opml.Parse(data)
	if err != nil {
		return output.PrintError(output.CodeInvalidArgs, "invalid OPML: "+err.Error())
	}

	type importError struct {
		XMLURL  string `json:"xml_url"`
		Message string `json:"message"`
	}

	var (
		imported int
		skipped  int
		errors   []importError
	)

	for _, f := range feeds {
		if f.XmlURL == "" {
			skipped++
			errors = append(errors, importError{XMLURL: f.Title, Message: "missing feed URL"})
			continue
		}

		existing, _ := feedRepo().GetByURL(f.XmlURL)
		if existing != nil {
			skipped++
			continue
		}

		created, err := feedRepo().Create(f.XmlURL)
		if err != nil {
			skipped++
			errors = append(errors, importError{XMLURL: f.XmlURL, Message: err.Error()})
			continue
		}

		// Best-effort: store OPML metadata (title, site URL) so export preserves them
		if f.Title != "" || f.HtmlURL != "" {
			_ = feedRepo().UpdateMeta(created.ID, f.Title, "", f.HtmlURL, "", "", "")
		}

		// Assign categories (best-effort)
		for _, catName := range f.Categories {
			cat, err := categoryRepo().GetByName(catName)
			if err != nil {
				cat, err = categoryRepo().Create(catName, "")
				if err != nil {
					continue
				}
			}
			if cat != nil {
				_ = feedRepo().AssignCategory(created.ID, cat.ID)
			}
		}

		imported++
	}

	summary := map[string]any{
		"imported": imported,
		"skipped":  skipped,
	}
	if len(errors) > 0 {
		summary["errors"] = errors
	}

	output.PrintSuccess(summary, nil)
	return nil
}
