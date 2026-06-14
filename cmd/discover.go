package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/0xfig-labs/tide/internal/output"
)

// DiscoverFeed represents a discovered RSS/Atom/JSON Feed.
type DiscoverFeed struct {
	URL   string `json:"url"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

var discoverCmd = &cobra.Command{
	Use:   "discover <url>",
	Short: "Discover RSS/Atom/JSON Feeds from a website",
	Long: `Discover RSS/Atom/JSON Feed URLs from a website's HTML.

Parses <link> tags for feed types (rss+xml, atom+xml, feed+json) and
also checks common feed URL paths as a fallback.

Example:
  tide discover https://example.com`,
	Args: cobra.ExactArgs(1),
	RunE: runDiscover,
}

func init() {
	rootCmd.AddCommand(discoverCmd)
}

// Common feed URL paths checked as a fallback when no <link> tags are found.
var commonFeedPaths = []string{
	"/feed.xml", "/rss.xml", "/feed", "/rss",
	"/atom.xml", "/feed.atom", "/index.xml", "/feed.json",
}

// Regex patterns for parsing HTML <link> tags.
var (
	linkTagRE   = regexp.MustCompile(`<link[^>]*?>`)
	hrefAttrRE  = regexp.MustCompile(`href\s*=\s*"([^"]*)"`)
	typeAttrRE  = regexp.MustCompile(`type\s*=\s*"([^"]*)"`)
	titleAttrRE = regexp.MustCompile(`title\s*=\s*"([^"]*)"`)
)

// feedMIMETypes maps MIME types to short feed type identifiers.
var feedMIMETypes = map[string]string{
	"application/rss+xml":   "rss",
	"application/atom+xml":  "atom",
	"application/feed+json": "json",
	"application/json":      "json",
}

func runDiscover(cmd *cobra.Command, args []string) error {
	siteURL := args[0]

	parsedURL, err := url.Parse(siteURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return output.PrintError(output.CodeInvalidArgs, "invalid URL: "+siteURL)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			if req.URL.Host != parsedURL.Host {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	resp, err := client.Get(siteURL)
	if err != nil {
		return output.PrintError(output.CodeFetchFailed, "cannot fetch page: "+err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return output.PrintError(output.CodeFetchFailed,
			fmt.Sprintf("page returned status %d", resp.StatusCode))
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return output.PrintError(output.CodeFetchFailed, "cannot read page body: "+err.Error())
	}

	htmlStr := string(bodyBytes)

	// Step 1: Discover feeds from <link> tags.
	seen := make(map[string]bool)
	var feeds []DiscoverFeed

	for _, tag := range linkTagRE.FindAllString(htmlStr, -1) {
		typeMatch := typeAttrRE.FindStringSubmatch(tag)
		if typeMatch == nil {
			continue
		}
		feedType, ok := feedMIMETypes[typeMatch[1]]
		if !ok {
			continue
		}

		hrefMatch := hrefAttrRE.FindStringSubmatch(tag)
		if hrefMatch == nil {
			continue
		}

		feedURL, err := resolveURL(parsedURL, hrefMatch[1])
		if err != nil || feedURL == "" {
			continue
		}

		// Ensure feed is on the same domain.
		feedParsed, err := url.Parse(feedURL)
		if err != nil || feedParsed.Host != parsedURL.Host {
			continue
		}

		if seen[feedURL] {
			continue
		}
		seen[feedURL] = true

		title := ""
		if titleMatch := titleAttrRE.FindStringSubmatch(tag); titleMatch != nil {
			title = titleMatch[1]
		}

		feeds = append(feeds, DiscoverFeed{
			URL:   feedURL,
			Type:  feedType,
			Title: title,
		})
	}

	// Step 2: Fallback — check common feed paths when no <link> feeds found.
	if len(feeds) == 0 {
		for _, feedPath := range commonFeedPaths {
			feedURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, feedPath)
			if seen[feedURL] {
				continue
			}

			req, err := http.NewRequest("HEAD", feedURL, nil)
			if err != nil {
				continue
			}
			headResp, err := client.Do(req)
			if err != nil || headResp.StatusCode >= 400 {
				if headResp != nil {
					headResp.Body.Close()
				}
				continue
			}
			headResp.Body.Close()

			feeds = append(feeds, DiscoverFeed{
				URL:  feedURL,
				Type: guessFeedType(feedPath),
			})
			seen[feedURL] = true
		}
	}

	data := map[string]any{
		"site_url": siteURL,
		"feeds":    feeds,
	}
	output.PrintSuccess(data, nil)
	return nil
}

// resolveURL resolves a potentially relative URL against a base URL.
func resolveURL(base *url.URL, href string) (string, error) {
	ref, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(ref).String(), nil
}

// guessFeedType guesses the feed type from a file path extension.
func guessFeedType(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".json"):
		return "json"
	case strings.Contains(lower, "atom"):
		return "atom"
	default:
		return "rss"
	}
}
