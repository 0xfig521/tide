// Package opml handles OPML 2.0 import and export for RSS feed subscriptions.
package opml

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"time"
)

// OpmlFeed represents a single RSS feed discovered in an OPML file.
type OpmlFeed struct {
	Title      string   `json:"title"`
	XmlURL     string   `json:"xml_url"`
	HtmlURL    string   `json:"html_url,omitempty"`
	Categories []string `json:"categories,omitempty"`
}

// FeedGroup is used for export: a category group containing feeds.
type FeedGroup struct {
	Category string
	Feeds    []OpmlFeed
}

// ------- XML types for parsing -------

type opmlDoc struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    opmlHead `xml:"head"`
	Body    opmlBody `xml:"body"`
}

type opmlHead struct {
	Title string `xml:"title"`
}

type opmlBody struct {
	Outlines []opmlOutline `xml:"outline"`
}

type opmlOutline struct {
	Text     string        `xml:"text,attr"`
	Title    string        `xml:"title,attr"`
	Type     string        `xml:"type,attr"`
	XmlURL   string        `xml:"xmlUrl,attr"`
	HtmlURL  string        `xml:"htmlUrl,attr"`
	Outlines []opmlOutline `xml:"outline"`
}

// ------- XML types for generating -------

type opmlGenDoc struct {
	XMLName xml.Name    `xml:"opml"`
	Version string      `xml:"version,attr"`
	Head    opmlGenHead `xml:"head"`
	Body    opmlGenBody `xml:"body"`
}

type opmlGenHead struct {
	Title       string `xml:"title"`
	DateCreated string `xml:"dateCreated,omitempty"`
}

type opmlGenBody struct {
	Outlines []opmlGenOutline `xml:"outline"`
}

type opmlGenOutline struct {
	Text     string           `xml:"text,attr,omitempty"`
	Title    string           `xml:"title,attr,omitempty"`
	Type     string           `xml:"type,attr,omitempty"`
	XmlURL   string           `xml:"xmlUrl,attr,omitempty"`
	HtmlURL  string           `xml:"htmlUrl,attr,omitempty"`
	Outlines []opmlGenOutline `xml:"outline,omitempty"`
}

// Parse reads OPML XML content and extracts all feeds with their category paths.
// Non-feed outlines (groups without xmlUrl) are treated as categories.
func Parse(data []byte) ([]OpmlFeed, error) {
	var doc opmlDoc
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("opml parse: %w", err)
	}

	var feeds []OpmlFeed
	collectFeeds(doc.Body.Outlines, nil, &feeds)
	return feeds, nil
}

// collectFeeds recursively walks outlines, tracking the current category chain.
func collectFeeds(outlines []opmlOutline, catPath []string, feeds *[]OpmlFeed) {
	for _, o := range outlines {
		if o.XmlURL != "" {
			// This is a feed leaf
			title := o.Title
			if title == "" {
				title = o.Text
			}
			f := OpmlFeed{
				Title:   title,
				XmlURL:  strings.TrimSpace(o.XmlURL),
				HtmlURL: strings.TrimSpace(o.HtmlURL),
			}
			if len(catPath) > 0 {
				// Copy category path to avoid aliasing
				f.Categories = make([]string, len(catPath))
				copy(f.Categories, catPath)
			}
			*feeds = append(*feeds, f)
		} else if len(o.Outlines) > 0 {
			// This is a category group — derive name from title or text
			catName := o.Title
			if catName == "" {
				catName = o.Text
			}
			if catName == "" {
				catName = "Uncategorized"
			}
			newPath := append(catPath, catName)
			collectFeeds(o.Outlines, newPath, feeds)
		}
		// Outlines that have neither xmlUrl nor children are skipped (e.g., empty groups)
	}
}

// Generate produces OPML 2.0 XML from a set of feeds grouped by category.
// It returns indented XML bytes.
func Generate(title string, groups []FeedGroup) ([]byte, error) {
	if title == "" {
		title = "Tide Feeds"
	}

	doc := opmlGenDoc{
		Version: "2.0",
		Head: opmlGenHead{
			Title:       title,
			DateCreated: time.Now().UTC().Format(time.RFC1123Z),
		},
	}

	var bodyOutlines []opmlGenOutline
	seenCategories := map[string]bool{}

	for _, g := range groups {
		if g.Category == "" {
			// Uncategorized feeds go directly in body
			for _, f := range g.Feeds {
				bodyOutlines = append(bodyOutlines, opmlGenOutline{
					Text:    f.Title,
					Title:   f.Title,
					Type:    "rss",
					XmlURL:  f.XmlURL,
					HtmlURL: f.HtmlURL,
				})
			}
		} else {
			// Category group
			seenCategories[g.Category] = true
			var children []opmlGenOutline
			for _, f := range g.Feeds {
				children = append(children, opmlGenOutline{
					Text:    f.Title,
					Title:   f.Title,
					Type:    "rss",
					XmlURL:  f.XmlURL,
					HtmlURL: f.HtmlURL,
				})
			}
			if len(children) > 0 {
				bodyOutlines = append(bodyOutlines, opmlGenOutline{
					Text:     g.Category,
					Title:    g.Category,
					Outlines: children,
				})
			}
		}
	}

	_ = seenCategories

	// Sort outlines: category groups first, then uncategorized feeds
	sort.Slice(bodyOutlines, func(i, j int) bool {
		// Group outlines (with children) come before leaf outlines
		iIsGroup := len(bodyOutlines[i].Outlines) > 0
		jIsGroup := len(bodyOutlines[j].Outlines) > 0
		if iIsGroup != jIsGroup {
			return iIsGroup
		}
		return strings.ToLower(bodyOutlines[i].Text) < strings.ToLower(bodyOutlines[j].Text)
	})

	doc.Body.Outlines = bodyOutlines

	output, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("opml generate: %w", err)
	}

	// Prepend XML declaration
	result := []byte(xml.Header)
	result = append(result, output...)
	result = append(result, '\n')
	return result, nil
}
