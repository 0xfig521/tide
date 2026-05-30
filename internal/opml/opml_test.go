package opml

import (
	"encoding/xml"
	"strings"
	"testing"
)

const fixtureFlat = `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>Test Feeds</title></head>
  <body>
    <outline text="Go Blog" title="Go Blog" type="rss" xmlUrl="https://go.dev/blog/feed.atom" htmlUrl="https://go.dev/blog/"/>
    <outline text="Rust Blog" title="Rust Blog" type="rss" xmlUrl="https://blog.rust-lang.org/feed.xml"/>
  </body>
</opml>`

const fixtureNested = `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>My Feeds</title></head>
  <body>
    <outline text="Tech" title="Tech">
      <outline text="Go Blog" title="Go Blog" type="rss" xmlUrl="https://go.dev/blog/feed.atom" htmlUrl="https://go.dev/blog/"/>
      <outline text="Rust Blog" title="Rust Blog" type="rss" xmlUrl="https://blog.rust-lang.org/feed.xml"/>
    </outline>
    <outline text="News" title="News">
      <outline text="HN" title="Hacker News" type="rss" xmlUrl="https://hnrss.org/frontpage" htmlUrl="https://news.ycombinator.com/"/>
    </outline>
    <outline text="No Cat" title="No Category Feed" type="rss" xmlUrl="https://example.com/feed.xml"/>
  </body>
</opml>`

const fixtureDeepNested = `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>Deep</title></head>
  <body>
    <outline text="Tech" title="Tech">
      <outline text="Languages" title="Languages">
        <outline text="Go Blog" type="rss" xmlUrl="https://go.dev/blog/feed.atom"/>
      </outline>
    </outline>
  </body>
</opml>`

const fixtureEmpty = `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>Empty</title></head>
  <body></body>
</opml>`

const fixtureNoXMLURL = `<?xml version="1.0" encoding="UTF-8"?>
<opml version="2.0">
  <head><title>Test</title></head>
  <body>
    <outline text="Group Only" title="Group">
      <outline text="Subgroup" title="Sub"/>
    </outline>
  </body>
</opml>`

func TestParseFlat(t *testing.T) {
	feeds, err := Parse([]byte(fixtureFlat))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(feeds) != 2 {
		t.Fatalf("expected 2 feeds, got %d", len(feeds))
	}
	if feeds[0].XmlURL != "https://go.dev/blog/feed.atom" {
		t.Errorf("unexpected XmlURL: %s", feeds[0].XmlURL)
	}
	if feeds[0].HtmlURL != "https://go.dev/blog/" {
		t.Errorf("unexpected HtmlURL: %s", feeds[0].HtmlURL)
	}
	if len(feeds[0].Categories) != 0 {
		t.Errorf("expected no categories, got %v", feeds[0].Categories)
	}
	if feeds[1].XmlURL != "https://blog.rust-lang.org/feed.xml" {
		t.Errorf("unexpected XmlURL: %s", feeds[1].XmlURL)
	}
}

func TestParseNested(t *testing.T) {
	feeds, err := Parse([]byte(fixtureNested))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(feeds) != 4 {
		t.Fatalf("expected 4 feeds, got %d", len(feeds))
	}

	// Go Blog should have category "Tech"
	goFeed := findFeed(feeds, "https://go.dev/blog/feed.atom")
	if goFeed == nil {
		t.Fatal("Go Blog feed not found")
	}
	if len(goFeed.Categories) != 1 || goFeed.Categories[0] != "Tech" {
		t.Errorf("expected [Tech], got %v", goFeed.Categories)
	}

	// HN should have category "News"
	hnFeed := findFeed(feeds, "https://hnrss.org/frontpage")
	if hnFeed == nil {
		t.Fatal("HN feed not found")
	}
	if len(hnFeed.Categories) != 1 || hnFeed.Categories[0] != "News" {
		t.Errorf("expected [News], got %v", hnFeed.Categories)
	}

	// No-category feed
	noCat := findFeed(feeds, "https://example.com/feed.xml")
	if noCat == nil {
		t.Fatal("No-category feed not found")
	}
	if len(noCat.Categories) != 0 {
		t.Errorf("expected no categories, got %v", noCat.Categories)
	}
}

func TestParseDeepNested(t *testing.T) {
	feeds, err := Parse([]byte(fixtureDeepNested))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(feeds) != 1 {
		t.Fatalf("expected 1 feed, got %d", len(feeds))
	}
	// Should have category path ["Tech", "Languages"]
	if len(feeds[0].Categories) != 2 {
		t.Errorf("expected 2 categories, got %d: %v", len(feeds[0].Categories), feeds[0].Categories)
	}
	if feeds[0].Categories[0] != "Tech" || feeds[0].Categories[1] != "Languages" {
		t.Errorf("expected [Tech Languages], got %v", feeds[0].Categories)
	}
}

func TestParseEmpty(t *testing.T) {
	feeds, err := Parse([]byte(fixtureEmpty))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(feeds) != 0 {
		t.Errorf("expected 0 feeds, got %d", len(feeds))
	}
}

func TestParseNoXMLURL(t *testing.T) {
	feeds, err := Parse([]byte(fixtureNoXMLURL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(feeds) != 0 {
		t.Errorf("expected 0 feeds, got %d", len(feeds))
	}
}

func TestParseInvalidXML(t *testing.T) {
	_, err := Parse([]byte("not xml"))
	if err == nil {
		t.Error("expected error for invalid XML")
	}
}

func TestGenerate(t *testing.T) {
	groups := []FeedGroup{
		{
			Category: "Tech",
			Feeds: []OpmlFeed{
				{Title: "Go Blog", XmlURL: "https://go.dev/blog/feed.atom", HtmlURL: "https://go.dev/blog/"},
				{Title: "Rust Blog", XmlURL: "https://blog.rust-lang.org/feed.xml"},
			},
		},
		{
			Category: "",
			Feeds: []OpmlFeed{
				{Title: "No Cat", XmlURL: "https://example.com/feed.xml"},
			},
		},
	}

	data, err := Generate("Test Export", groups)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's valid XML with OPML root
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var rootName string
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		if se, ok := tok.(xml.StartElement); ok {
			rootName = se.Name.Local
			break
		}
	}
	if rootName != "opml" {
		t.Errorf("expected root <opml>, got <%s>", rootName)
	}

	// Verify it contains expected URLs
	s := string(data)
	if !strings.Contains(s, "https://go.dev/blog/feed.atom") {
		t.Error("missing Go Blog URL")
	}
	if !strings.Contains(s, "https://example.com/feed.xml") {
		t.Error("missing uncategorized feed URL")
	}
	if !strings.Contains(s, `version="2.0"`) {
		t.Error("missing OPML version")
	}
	if !strings.Contains(s, "<title>Test Export</title>") {
		t.Error("missing title")
	}
}

func TestGenerateRoundtrip(t *testing.T) {
	// Parse, then generate, then parse again — should preserve feed structure
	feeds, err := Parse([]byte(fixtureNested))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Build groups from parsed feeds
	catMap := map[string][]OpmlFeed{}
	var uncat []OpmlFeed
	for _, f := range feeds {
		if len(f.Categories) == 0 {
			uncat = append(uncat, f)
		} else {
			catMap[f.Categories[0]] = append(catMap[f.Categories[0]], f)
		}
	}

	var groups []FeedGroup
	for cat, fds := range catMap {
		groups = append(groups, FeedGroup{Category: cat, Feeds: fds})
	}
	if len(uncat) > 0 {
		groups = append(groups, FeedGroup{Category: "", Feeds: uncat})
	}

	data, err := Generate("Roundtrip", groups)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	// Re-parse
	feeds2, err := Parse(data)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}

	if len(feeds2) != len(feeds) {
		t.Errorf("roundtrip: expected %d feeds, got %d", len(feeds), len(feeds2))
	}

	// Verify all URLs present
	urls := map[string]bool{}
	for _, f := range feeds2 {
		urls[f.XmlURL] = true
	}
	for _, f := range feeds {
		if !urls[f.XmlURL] {
			t.Errorf("roundtrip: missing URL %s", f.XmlURL)
		}
	}
}

func TestGenerateDefaultTitle(t *testing.T) {
	data, err := Generate("", []FeedGroup{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(data), "<title>Tide Feeds</title>") {
		t.Error("expected default title")
	}
}

func findFeed(feeds []OpmlFeed, xmlURL string) *OpmlFeed {
	for i := range feeds {
		if feeds[i].XmlURL == xmlURL {
			return &feeds[i]
		}
	}
	return nil
}
