package fetcher

import (
	"testing"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/0xfig-labs/tide/pkg"
)

func TestConvertEntry_BasicMapping(t *testing.T) {
	pubTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	item := &gofeed.Item{
		Title:           "Test Title",
		Link:            "https://example.com/article/1",
		GUID:            "guid-abc-123",
		Content:         "<p>Full article content here.</p>",
		Description:     "Short description.",
		PublishedParsed: &pubTime,
		Authors: []*gofeed.Person{
			{Name: "Alice"},
		},
		Categories: []string{"tech", "go", "testing"},
	}
	item.Image = &gofeed.Image{URL: "https://example.com/thumb.jpg"}

	entry := ConvertEntry(42, item)

	if entry.FeedID != 42 {
		t.Errorf("FeedID = %d, want 42", entry.FeedID)
	}
	if entry.Title != "Test Title" {
		t.Errorf("Title = %q, want %q", entry.Title, "Test Title")
	}
	if entry.URL != "https://example.com/article/1" {
		t.Errorf("URL = %q, want %q", entry.URL, "https://example.com/article/1")
	}
	if entry.GUID != "guid-abc-123" {
		t.Errorf("GUID = %q, want %q", entry.GUID, "guid-abc-123")
	}
	if entry.Content != "<p>Full article content here.</p>" {
		t.Errorf("Content = %q, want %q", entry.Content, "<p>Full article content here.</p>")
	}
	if entry.Description != "Short description." {
		t.Errorf("Description = %q, want %q", entry.Description, "Short description.")
	}
	if entry.AuthorName != "Alice" {
		t.Errorf("AuthorName = %q, want %q", entry.AuthorName, "Alice")
	}
	if entry.Categories != "tech,go,testing" {
		t.Errorf("Categories = %q, want %q", entry.Categories, "tech,go,testing")
	}
	if entry.ImageURL != "https://example.com/thumb.jpg" {
		t.Errorf("ImageURL = %q, want %q", entry.ImageURL, "https://example.com/thumb.jpg")
	}
	if entry.PublishedAt == nil || !entry.PublishedAt.Equal(pubTime) {
		t.Errorf("PublishedAt = %v, want %v", entry.PublishedAt, pubTime)
	}
	expectedHash := pkg.EntryHash(42, "guid-abc-123")
	if entry.Hash != expectedHash {
		t.Errorf("Hash = %q, want %q", entry.Hash, expectedHash)
	}
}

func TestConvertEntry_PublishedAtFallback(t *testing.T) {
	updatedTime := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	item := &gofeed.Item{
		Title:         "No Published Date",
		Link:          "https://example.com/fallback",
		GUID:          "guid-fallback",
		UpdatedParsed: &updatedTime,
		// PublishedParsed intentionally nil
	}

	entry := ConvertEntry(1, item)

	if entry.PublishedAt == nil {
		t.Fatal("PublishedAt should not be nil, expected fallback to UpdatedParsed")
	}
	if !entry.PublishedAt.Equal(updatedTime) {
		t.Errorf("PublishedAt = %v, want %v (UpdatedParsed fallback)", entry.PublishedAt, updatedTime)
	}
}

func TestConvertEntry_PublishedParsedPriority(t *testing.T) {
	pubTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedTime := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	item := &gofeed.Item{
		Title:           "Both Dates Set",
		Link:            "https://example.com/both",
		GUID:            "guid-both",
		PublishedParsed: &pubTime,
		UpdatedParsed:   &updatedTime,
	}

	entry := ConvertEntry(1, item)

	if entry.PublishedAt == nil {
		t.Fatal("PublishedAt should not be nil")
	}
	if !entry.PublishedAt.Equal(pubTime) {
		t.Errorf("PublishedAt = %v, want %v (should prefer PublishedParsed)", entry.PublishedAt, pubTime)
	}
}

func TestConvertEntry_HashStability(t *testing.T) {
	// Same feedID + GUID must always produce the same hash.
	hash1 := ConvertEntry(99, &gofeed.Item{
		Title: "Article A",
		Link:  "https://example.com/a",
		GUID:  "stable-guid",
	}).Hash

	hash2 := ConvertEntry(99, &gofeed.Item{
		Title: "Article A - different title should not matter",
		Link:  "https://example.com/different-url",
		GUID:  "stable-guid",
	}).Hash

	if hash1 != hash2 {
		t.Errorf("hash mismatch: %q vs %q — same (feedID, GUID) must produce same hash", hash1, hash2)
	}

	// Different feedID should produce different hash.
	hash3 := ConvertEntry(100, &gofeed.Item{
		GUID: "stable-guid",
	}).Hash
	if hash3 == hash1 {
		t.Errorf("different feedID should produce different hash, got %q for both", hash1)
	}

	// Different GUID should produce different hash.
	hash4 := ConvertEntry(99, &gofeed.Item{
		GUID: "different-guid",
	}).Hash
	if hash4 == hash1 {
		t.Errorf("different GUID should produce different hash, got %q for both", hash1)
	}
}

func TestConvertEntry_AuthorFallback(t *testing.T) {
	// When Authors slice is empty, fall back to item.Author
	item := &gofeed.Item{
		Title:  "Author Fallback",
		Link:   "https://example.com/author",
		GUID:   "guid-author",
		Author: &gofeed.Person{Name: "LegacyAuthor"},
		// Authors intentionally empty/nil
	}

	entry := ConvertEntry(1, item)

	if entry.AuthorName != "LegacyAuthor" {
		t.Errorf("AuthorName = %q, want %q", entry.AuthorName, "LegacyAuthor")
	}
}

func TestConvertEntry_EmptyCategories(t *testing.T) {
	item := &gofeed.Item{
		Title: "No Categories",
		Link:  "https://example.com/no-cat",
		GUID:  "guid-nocat",
	}

	entry := ConvertEntry(1, item)

	if entry.Categories != "" {
		t.Errorf("Categories = %q, want empty string", entry.Categories)
	}
}

func TestConvertEntry_NoDates(t *testing.T) {
	item := &gofeed.Item{
		Title: "No Dates At All",
		Link:  "https://example.com/nodate",
		GUID:  "guid-nodate",
		// Both PublishedParsed and UpdatedParsed nil
	}

	entry := ConvertEntry(1, item)

	if entry.PublishedAt != nil {
		t.Errorf("PublishedAt = %v, want nil when no dates available", entry.PublishedAt)
	}
}
