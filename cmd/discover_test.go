package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// captureStdout runs fn and returns everything written to stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = orig

	out, _ := io.ReadAll(r)
	return string(out)
}

// ---------------------------------------------------------------------------
// resolveURL tests
// ---------------------------------------------------------------------------

func TestResolveURL_Absolute(t *testing.T) {
	base, err := url.Parse("https://example.com/blog/")
	if err != nil {
		t.Fatalf("failed to parse base URL: %v", err)
	}

	t.Run("absolute URL overrides base", func(t *testing.T) {
		result, err := resolveURL(base, "https://other.com/feed.xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "https://other.com/feed.xml" {
			t.Errorf("got %q, want 'https://other.com/feed.xml'", result)
		}
	})

	t.Run("absolute URL with same host", func(t *testing.T) {
		result, err := resolveURL(base, "https://example.com/rss.xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "https://example.com/rss.xml" {
			t.Errorf("got %q, want 'https://example.com/rss.xml'", result)
		}
	})
}

func TestResolveURL_Relative(t *testing.T) {
	base, err := url.Parse("https://example.com/blog/")
	if err != nil {
		t.Fatalf("failed to parse base URL: %v", err)
	}

	t.Run("root-relative path", func(t *testing.T) {
		result, err := resolveURL(base, "/feed.xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "https://example.com/feed.xml" {
			t.Errorf("got %q, want 'https://example.com/feed.xml'", result)
		}
	})

	t.Run("path-relative resolves against base directory", func(t *testing.T) {
		result, err := resolveURL(base, "feed.atom")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "https://example.com/blog/feed.atom" {
			t.Errorf("got %q, want 'https://example.com/blog/feed.atom'", result)
		}
	})

	t.Run("parent traversal", func(t *testing.T) {
		result, err := resolveURL(base, "../feed.xml")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "https://example.com/feed.xml" {
			t.Errorf("got %q, want 'https://example.com/feed.xml'", result)
		}
	})
}

// ---------------------------------------------------------------------------
// guessFeedType tests
// ---------------------------------------------------------------------------

func TestGuessFeedType_RSS(t *testing.T) {
	tests := []string{"/feed.xml", "/rss.xml", "/feed", "/rss", "/index.xml", "/some/path"}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			if got := guessFeedType(path); got != "rss" {
				t.Errorf("guessFeedType(%q) = %q, want 'rss'", path, got)
			}
		})
	}
}

func TestGuessFeedType_Atom(t *testing.T) {
	tests := []string{"/atom.xml", "/feed.atom", "/feed/atom", "/atom/feed"}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			if got := guessFeedType(path); got != "atom" {
				t.Errorf("guessFeedType(%q) = %q, want 'atom'", path, got)
			}
		})
	}
}

func TestGuessFeedType_JSON(t *testing.T) {
	tests := []string{"/feed.json", "/subscriptions.json", "/data.json"}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			if got := guessFeedType(path); got != "json" {
				t.Errorf("guessFeedType(%q) = %q, want 'json'", path, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// discover command tests
// ---------------------------------------------------------------------------

// responseEnvelope is a minimal struct to parse the JSON envelope from discover output.
type responseEnvelope struct {
	OK    bool          `json:"ok"`
	Data  discoverData  `json:"data"`
	Error *errorPayload `json:"error"`
}

type discoverData struct {
	SiteURL string         `json:"site_url"`
	Feeds   []DiscoverFeed `json:"feeds"`
}

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func TestDiscover_InvalidURL(t *testing.T) {
	cmd := &cobra.Command{}

	tests := []struct {
		name string
		url  string
	}{
		{"empty string", ""},
		{"no scheme", "example.com"},
		{"ftp scheme", "ftp://example.com"},
		{"malformed", "://invalid"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output := captureStdout(t, func() {
				_ = runDiscover(cmd, []string{tc.url})
			})

			var resp responseEnvelope
			if err := json.Unmarshal([]byte(output), &resp); err != nil {
				t.Fatalf("invalid JSON output: %v\nraw: %s", err, output)
			}
			if resp.OK {
				t.Error("expected ok=false for invalid URL")
			}
			if resp.Error == nil {
				t.Fatal("expected error to be non-nil")
			}
			if resp.Error.Code != "invalid_args" {
				t.Errorf("error.code = %q, want 'invalid_args'", resp.Error.Code)
			}
		})
	}
}

func TestDiscover_HTMLParsing_MockPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html>
<head>
  <link rel="alternate" type="application/rss+xml" title="RSS Feed" href="/feed.xml">
</head>
<body></body>
</html>`)
	}))
	defer server.Close()

	cmd := &cobra.Command{}

	output := captureStdout(t, func() {
		_ = runDiscover(cmd, []string{server.URL})
	})

	var resp responseEnvelope
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", err, output)
	}
	if !resp.OK {
		t.Error("expected ok=true")
	}
	if resp.Data.SiteURL != server.URL {
		t.Errorf("site_url = %q, want %q", resp.Data.SiteURL, server.URL)
	}
	if len(resp.Data.Feeds) != 1 {
		t.Fatalf("expected 1 feed, got %d", len(resp.Data.Feeds))
	}

	feed := resp.Data.Feeds[0]
	expectedURL := server.URL + "/feed.xml"
	if feed.URL != expectedURL {
		t.Errorf("feed URL = %q, want %q", feed.URL, expectedURL)
	}
	if feed.Type != "rss" {
		t.Errorf("feed type = %q, want 'rss'", feed.Type)
	}
	if feed.Title != "RSS Feed" {
		t.Errorf("feed title = %q, want 'RSS Feed'", feed.Title)
	}
}

func TestDiscover_HTMLParsing_NoFeeds(t *testing.T) {
	// No feed links, and no fallback paths respond — should return empty feeds.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><head></head><body>No feeds here</body></html>`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cmd := &cobra.Command{}

	output := captureStdout(t, func() {
		_ = runDiscover(cmd, []string{server.URL})
	})

	var resp responseEnvelope
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", err, output)
	}
	if !resp.OK {
		t.Errorf("expected ok=true, got ok=%v, error=%+v", resp.OK, resp.Error)
	}
	if len(resp.Data.Feeds) != 0 {
		t.Errorf("expected 0 feeds, got %d: %+v", len(resp.Data.Feeds), resp.Data.Feeds)
	}
}

func TestDiscover_HTMLParsing_MultipleFeeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html>
<head>
  <link rel="alternate" type="application/rss+xml" title="RSS" href="/rss.xml">
  <link rel="alternate" type="application/atom+xml" title="Atom" href="/atom.xml">
  <link rel="alternate" type="application/feed+json" title="JSON Feed" href="/feed.json">
</head>
<body></body>
</html>`)
	}))
	defer server.Close()

	cmd := &cobra.Command{}

	output := captureStdout(t, func() {
		_ = runDiscover(cmd, []string{server.URL})
	})

	var resp responseEnvelope
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", err, output)
	}
	if !resp.OK {
		t.Error("expected ok=true")
	}
	if len(resp.Data.Feeds) != 3 {
		t.Fatalf("expected 3 feeds, got %d", len(resp.Data.Feeds))
	}

	// Verify all three feeds are present with correct types.
	expected := map[string]string{
		server.URL + "/rss.xml":   "rss",
		server.URL + "/atom.xml":  "atom",
		server.URL + "/feed.json": "json",
	}
	found := 0
	for _, feed := range resp.Data.Feeds {
		wantType, ok := expected[feed.URL]
		if !ok {
			t.Errorf("unexpected feed URL: %s", feed.URL)
			continue
		}
		if feed.Type != wantType {
			t.Errorf("feed %q: type = %q, want %q", feed.URL, feed.Type, wantType)
		}
		found++
	}
	if found != 3 {
		t.Errorf("found %d expected feeds, want 3", found)
	}
}

func TestDiscover_FallbackPaths(t *testing.T) {
	// Serve a page with no feed link tags; only certain fallback paths respond.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Main page: no feed links
		if r.URL.Path == "/" || r.URL.Path == "" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><head></head><body>No feeds in link tags</body></html>`)
			return
		}
		// Only /feed.xml and /feed.json exist on this server.
		switch r.URL.Path {
		case "/feed.xml":
			w.Header().Set("Content-Type", "application/rss+xml")
			w.WriteHeader(http.StatusOK)
		case "/feed.json":
			w.Header().Set("Content-Type", "application/feed+json")
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cmd := &cobra.Command{}

	output := captureStdout(t, func() {
		_ = runDiscover(cmd, []string{server.URL})
	})

	var resp responseEnvelope
	if err := json.Unmarshal([]byte(output), &resp); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", err, output)
	}
	if !resp.OK {
		t.Errorf("expected ok=true, got ok=%v", resp.OK)
	}
	if len(resp.Data.Feeds) != 2 {
		t.Fatalf("expected 2 feeds from fallback, got %d: %+v", len(resp.Data.Feeds), resp.Data.Feeds)
	}

	foundRSS := false
	foundJSON := false
	for _, feed := range resp.Data.Feeds {
		if strings.HasSuffix(feed.URL, "/feed.xml") && feed.Type == "rss" {
			foundRSS = true
		}
		if strings.HasSuffix(feed.URL, "/feed.json") && feed.Type == "json" {
			foundJSON = true
		}
	}
	if !foundRSS {
		t.Error("expected RSS feed (/feed.xml) via fallback")
	}
	if !foundJSON {
		t.Error("expected JSON feed (/feed.json) via fallback")
	}
}
