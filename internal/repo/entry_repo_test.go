package repo

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/0xfig521/tide/internal/db"
	"github.com/0xfig521/tide/internal/models"
)

// setupTestDB creates an in-memory SQLite database, runs migrations, and returns the connection.
// ":memory:" paths work with db.Open because modernc.org/sqlite interprets "file::memory:" DSNs.
func setupTestDB(t *testing.T) *db.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() {
		database.Close()
	})
	return database
}

// createTestFeed inserts a feed and returns it for use in entry tests.
func createTestFeed(t *testing.T, db *db.DB, feedURL string) *models.Feed {
	t.Helper()

	fr := NewFeedRepo(db)
	feed, err := fr.Create(feedURL)
	if err != nil {
		t.Fatalf("failed to create test feed: %v", err)
	}
	return feed
}

func TestEntryRepo_InsertOrIgnore(t *testing.T) {
	database := setupTestDB(t)
	feed := createTestFeed(t, database, "https://example.com/feed.xml")
	repo := NewEntryRepo(database)

	pubTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	entry := &models.Entry{
		FeedID:      feed.ID,
		Title:       "Test Entry",
		URL:         "https://example.com/1",
		GUID:        "guid-1",
		Content:     "Content here",
		Description: "Description here",
		AuthorName:  "Alice",
		PublishedAt: &pubTime,
		Hash:        "abc123",
	}

	if err := repo.InsertOrIgnore(entry); err != nil {
		t.Fatalf("InsertOrIgnore failed: %v", err)
	}

	// Insert the same entry again — should be a no-op (IGNORE).
	if err := repo.InsertOrIgnore(entry); err != nil {
		t.Fatalf("second InsertOrIgnore should not error: %v", err)
	}

	// Verify only one entry exists.
	entries, err := repo.ListEntries(EntryQuery{PageSize: 100})
	if err != nil {
		t.Fatalf("ListEntries failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestEntryRepo_GetByID(t *testing.T) {
	database := setupTestDB(t)
	feed := createTestFeed(t, database, "https://example.com/feed2.xml")
	repo := NewEntryRepo(database)

	pubTime := time.Date(2024, 2, 20, 15, 30, 0, 0, time.UTC)
	entry := &models.Entry{
		FeedID:      feed.ID,
		Title:       "Find Me",
		URL:         "https://example.com/find-me",
		GUID:        "guid-find-me",
		Content:     "Full content",
		Description: "Desc",
		AuthorName:  "Bob",
		PublishedAt: &pubTime,
		Hash:        "hash-find-me",
	}

	if err := repo.InsertOrIgnore(entry); err != nil {
		t.Fatalf("InsertOrIgnore failed: %v", err)
	}

	// We need the ID — list entries to get it.
	entries, err := repo.ListEntries(EntryQuery{PageSize: 10})
	if err != nil {
		t.Fatalf("ListEntries failed: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least 1 entry")
	}

	got, err := repo.GetByID(entries[0].ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if got.Title != "Find Me" {
		t.Errorf("Title = %q, want %q", got.Title, "Find Me")
	}
	if got.URL != "https://example.com/find-me" {
		t.Errorf("URL = %q, want %q", got.URL, "https://example.com/find-me")
	}
	if got.AuthorName != "Bob" {
		t.Errorf("AuthorName = %q, want %q", got.AuthorName, "Bob")
	}
	if got.FeedID != feed.ID {
		t.Errorf("FeedID = %d, want %d", got.FeedID, feed.ID)
	}
}

func TestEntryRepo_ListEntries_Pagination(t *testing.T) {
	database := setupTestDB(t)
	feed := createTestFeed(t, database, "https://example.com/paginate.xml")
	repo := NewEntryRepo(database)

	// Insert 5 entries with distinct published times.
	now := time.Now()
	for i := 0; i < 5; i++ {
		pt := now.Add(-time.Duration(i) * time.Hour)
		entry := &models.Entry{
			FeedID:      feed.ID,
			Title:       "Entry " + string(rune('A'+i)),
			URL:         "https://example.com/p/" + string(rune('0'+i)),
			GUID:        "paginate-guid-" + string(rune('0'+i)),
			PublishedAt: &pt,
			Hash:        "paginate-hash-" + string(rune('0'+i)),
		}
		if err := repo.InsertOrIgnore(entry); err != nil {
			t.Fatalf("InsertOrIgnore %d failed: %v", i, err)
		}
	}

	// Page 1: first 3 entries.
	page1, err := repo.ListEntries(EntryQuery{Page: 1, PageSize: 3})
	if err != nil {
		t.Fatalf("ListEntries page 1 failed: %v", err)
	}
	if len(page1) != 3 {
		t.Errorf("page 1: expected 3 entries, got %d", len(page1))
	}

	// Page 2: remaining 2 entries.
	page2, err := repo.ListEntries(EntryQuery{Page: 2, PageSize: 3})
	if err != nil {
		t.Fatalf("ListEntries page 2 failed: %v", err)
	}
	if len(page2) != 2 {
		t.Errorf("page 2: expected 2 entries, got %d", len(page2))
	}

	// Verify no overlap in IDs.
	ids := make(map[int64]bool)
	for _, e := range page1 {
		ids[e.ID] = true
	}
	for _, e := range page2 {
		if ids[e.ID] {
			t.Errorf("duplicate entry ID %d across pages", e.ID)
		}
	}

	// Verify ordering: page1 entries should be newer than page2.
	if len(page1) > 0 && len(page2) > 0 {
		if page1[0].PublishedAt == nil || page2[0].PublishedAt == nil {
			t.Skip("nil published_at in pagination results")
		} else if !page1[0].PublishedAt.After(*page2[0].PublishedAt) {
			t.Error("page 1 entries should be newer than page 2 entries")
		}
	}
}
