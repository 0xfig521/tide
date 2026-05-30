package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/opml"
	"github.com/0xfig521/tide/internal/output"
)

var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export feeds to an OPML file",
	Long: `Export all RSS feed subscriptions as an OPML 2.0 file.
	
By default, the OPML XML is written to stdout.
Use --output to write to a file instead.`,
	RunE: runExport,
}

func init() {
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file path (default: stdout)")
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	feeds, err := feedRepo().List("")
	if err != nil {
		return output.PrintError(output.CodeInternalError, err.Error())
	}

	// Group feeds by category
	type catKey struct {
		name string
	}
	catGroups := map[string][]opml.OpmlFeed{}
	var uncategorized []opml.OpmlFeed

	for _, f := range feeds {
		cats, _ := feedRepo().GetCategories(f.ID)

		of := opml.OpmlFeed{
			Title:   f.Title,
			XmlURL:  f.FeedURL,
			HtmlURL: f.SiteURL,
		}

		if len(cats) == 0 {
			uncategorized = append(uncategorized, of)
		} else {
			for _, catName := range cats {
				catGroups[catName] = append(catGroups[catName], of)
			}
		}
	}

	// Build ordered group list
	var groups []opml.FeedGroup
	for catName, catFeeds := range catGroups {
		groups = append(groups, opml.FeedGroup{Category: catName, Feeds: catFeeds})
	}
	if len(uncategorized) > 0 {
		groups = append(groups, opml.FeedGroup{Category: "", Feeds: uncategorized})
	}

	xmlData, err := opml.Generate("Tide Feeds", groups)
	if err != nil {
		return output.PrintError(output.CodeInternalError, err.Error())
	}

	if exportOutput != "" {
		if err := os.WriteFile(exportOutput, xmlData, 0644); err != nil {
			return output.PrintError(output.CodeInternalError, "cannot write file: "+err.Error())
		}
		output.PrintSuccess(map[string]string{"file": exportOutput, "feeds": fmt.Sprintf("%d", len(feeds))}, nil)
	} else {
		// Write OPML XML to stdout directly (not JSON-wrapped, by design)
		fmt.Print(string(xmlData))
	}

	return nil
}
