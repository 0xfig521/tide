package repo

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/0xfig-labs/tide/internal/db"
	"github.com/0xfig-labs/tide/internal/models"
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

// TestEntryRepo_ListEntries_SortByRelevance verifies that SortBy string values
// are accepted and produce correct ordering. With no keyword, SortBy=relevance
// falls back to published_at DESC like the default.
func TestEntryRepo_ListEntries_SortByRelevance(t *testing.T) {
	database := setupTestDB(t)
	feed := createTestFeed(t, database, "https://example.com/relevance.xml")
	repo := NewEntryRepo(database)

	now := time.Now()

	offsets := []int{3, 1, 2}
	for i, off := range offsets {
		pt := now.Add(-time.Duration(off) * time.Hour)
		entry := &models.Entry{
			FeedID:      feed.ID,
			Title:       "Entry " + string(rune('A'+i)),
			URL:         "https://example.com/sr/" + string(rune('0'+i)),
			GUID:        "sort-rel-guid-" + string(rune('0'+i)),
			PublishedAt: &pt,
			Hash:        "sort-rel-hash-" + string(rune('0'+i)),
		}
		if err := repo.InsertOrIgnore(entry); err != nil {
			t.Fatalf("InsertOrIgnore %d failed: %v", i, err)
		}
	}

	// Default sort = published_at DESC.
	def, err := repo.ListEntries(EntryQuery{PageSize: 10})
	if err != nil {
		t.Fatalf("ListEntries (default) failed: %v", err)
	}
	if len(def) != 3 {
		t.Fatalf("default: expected 3, got %d", len(def))
	}

	// SortBy=published gives same order.
	pub, err := repo.ListEntries(EntryQuery{SortBy: "published", PageSize: 10})
	if err != nil {
		t.Fatalf("ListEntries (published) failed: %v", err)
	}
	for i := range def {
		if pub[i].ID != def[i].ID {
			t.Errorf("published[%d]=%d != default[%d]=%d", i, pub[i].ID, i, def[i].ID)
		}
	}

	// SortBy=relevance without keyword falls back to published_at DESC.
	rel, err := repo.ListEntries(EntryQuery{SortBy: "relevance", PageSize: 10})
	if err != nil {
		t.Fatalf("ListEntries (relevance, no keyword) failed: %v", err)
	}
	if len(rel) != 3 {
		t.Fatalf("relevance: expected 3, got %d", len(rel))
	}
	for i := range def {
		if rel[i].ID != def[i].ID {
			t.Errorf("relevance[%d]=%d != default[%d]=%d", i, rel[i].ID, i, def[i].ID)
		}
	}
}

// TestEntryRepo_ListEntries_SortByPublished verifies that SortBy=published
// (or the default empty string) returns entries ordered by published_at DESC.
func TestEntryRepo_ListEntries_SortByPublished(t *testing.T) {
	database := setupTestDB(t)
	feed := createTestFeed(t, database, "https://example.com/sort-published.xml")
	repo := NewEntryRepo(database)

	now := time.Now()

	// Insert 3 entries in reverse chronological order: oldest → middle → newest.
	for i, offset := range []int{3, 2, 1} {
		pt := now.Add(-time.Duration(offset) * time.Hour)
		entry := &models.Entry{
			FeedID:      feed.ID,
			Title:       "Entry " + string(rune('A'+i)),
			URL:         "https://example.com/sp/" + string(rune('0'+i)),
			GUID:        "sort-pub-guid-" + string(rune('0'+i)),
			Content:     "sort by published test",
			Description: "Sort by published test",
			PublishedAt: &pt,
			Hash:        "sort-pub-hash-" + string(rune('0'+i)),
		}
		if err := repo.InsertOrIgnore(entry); err != nil {
			t.Fatalf("InsertOrIgnore %d failed: %v", i, err)
		}
	}

	// Default sort (empty string) = published_at DESC.
	results, err := repo.ListEntries(EntryQuery{PageSize: 10})
	if err != nil {
		t.Fatalf("ListEntries failed: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3, got %d", len(results))
	}
	expectedGUIDs := []string{"sort-pub-guid-2", "sort-pub-guid-1", "sort-pub-guid-0"}
	for i, exp := range expectedGUIDs {
		if results[i].GUID != exp {
			t.Errorf("published sort [%d]: got GUID %s, want %s", i, results[i].GUID, exp)
		}
	}

	// Explicit SortBy=published should give the same order.
	results2, err := repo.ListEntries(EntryQuery{SortBy: "published", PageSize: 10})
	if err != nil {
		t.Fatalf("ListEntries (explicit published) failed: %v", err)
	}
	for i := range expectedGUIDs {
		if results2[i].GUID != results[i].GUID {
			t.Errorf("explicit published sort [%d] differs from default: %s vs %s",
				i, results2[i].GUID, results[i].GUID)
		}
	}
}

// TestEntryRepo_ListEntries_StateFilter verifies that ListEntries with a State
// filter only returns entries whose state matches, via LEFT JOIN on entry_states.
func TestEntryRepo_ListEntries_StateFilter(t *testing.T) {
	database := setupTestDB(t)
	feed := createTestFeed(t, database, "https://example.com/state-filter.xml")
	repo := NewEntryRepo(database)
	stateRepo := NewEntryStateRepo(database)

	now := time.Now()

	// Insert 3 entries.
	for i := 0; i < 3; i++ {
		pt := now.Add(-time.Duration(i) * time.Hour)
		entry := &models.Entry{
			FeedID:      feed.ID,
			Title:       "State Entry " + string(rune('A'+i)),
			URL:         "https://example.com/sf/" + string(rune('0'+i)),
			GUID:        "state-filter-guid-" + string(rune('0'+i)),
			Content:     "state filter test content",
			Description: "State filter test",
			PublishedAt: &pt,
			Hash:        "state-filter-hash-" + string(rune('0'+i)),
		}
		if err := repo.InsertOrIgnore(entry); err != nil {
			t.Fatalf("InsertOrIgnore %d failed: %v", i, err)
		}
	}

	// List all entries to get their IDs.
	allEntries, err := repo.ListEntries(EntryQuery{PageSize: 10})
	if err != nil {
		t.Fatalf("ListEntries failed: %v", err)
	}
	if len(allEntries) != 3 {
		t.Fatalf("expected 3, got %d", len(allEntries))
	}

	// Set state "processed" on entries 0 and 2. Entry 1 stays without any state.
	if err := stateRepo.SetState(allEntries[0].ID, "processed"); err != nil {
		t.Fatalf("SetState on entry 0 failed: %v", err)
	}
	if err := stateRepo.SetState(allEntries[2].ID, "processed"); err != nil {
		t.Fatalf("SetState on entry 2 failed: %v", err)
	}

	// Filter by state="processed": should return entries 0 and 2 only.
	processed, err := repo.ListEntries(EntryQuery{State: "processed", PageSize: 10})
	if err != nil {
		t.Fatalf("ListEntries (state=processed) failed: %v", err)
	}
	if len(processed) != 2 {
		t.Fatalf("expected 2 processed entries, got %d", len(processed))
	}

	// Verify correct entries returned.
	seen := make(map[int64]bool)
	for _, e := range processed {
		seen[e.ID] = true
	}
	if !seen[allEntries[0].ID] {
		t.Error("entry 0 (processed) not in state-filtered results")
	}
	if !seen[allEntries[2].ID] {
		t.Error("entry 2 (processed) not in state-filtered results")
	}
	if seen[allEntries[1].ID] {
		t.Error("entry 1 (no state) should not appear in state=processed results")
	}
}

// TestEntryRepo_ListEntries_StateFilter_NoMatches verifies that a state filter
// matching no entries returns an empty list (not an error).
func TestEntryRepo_ListEntries_StateFilter_NoMatches(t *testing.T) {
	database := setupTestDB(t)
	feed := createTestFeed(t, database, "https://example.com/state-no-match.xml")
	repo := NewEntryRepo(database)
	stateRepo := NewEntryStateRepo(database)

	now := time.Now()

	// Insert 2 entries and set state "seen" on both.
	for i := 0; i < 2; i++ {
		pt := now.Add(-time.Duration(i) * time.Hour)
		entry := &models.Entry{
			FeedID:      feed.ID,
			Title:       "NoMatch Entry " + string(rune('A'+i)),
			URL:         "https://example.com/snm/" + string(rune('0'+i)),
			GUID:        "state-nomatch-guid-" + string(rune('0'+i)),
			Content:     "nomatch content",
			Description: "No match test",
			PublishedAt: &pt,
			Hash:        "state-nomatch-hash-" + string(rune('0'+i)),
		}
		if err := repo.InsertOrIgnore(entry); err != nil {
			t.Fatalf("InsertOrIgnore %d failed: %v", i, err)
		}
	}

	allEntries, err := repo.ListEntries(EntryQuery{PageSize: 10})
	if err != nil {
		t.Fatalf("ListEntries failed: %v", err)
	}
	for _, e := range allEntries {
		if err := stateRepo.SetState(e.ID, "seen"); err != nil {
			t.Fatalf("SetState failed: %v", err)
		}
	}

	// Query for state="processed" — none have this state.
	results, err := repo.ListEntries(EntryQuery{State: "processed", PageSize: 10})
	if err != nil {
		t.Fatalf("ListEntries (state=processed, no matches) failed: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// TestEntryRepo_CountEntries_StateFilter verifies that CountEntries respects
// the State filter and returns the correct count.
func TestEntryRepo_CountEntries_StateFilter(t *testing.T) {
	database := setupTestDB(t)
	feed := createTestFeed(t, database, "https://example.com/count-state.xml")
	repo := NewEntryRepo(database)
	stateRepo := NewEntryStateRepo(database)

	now := time.Now()

	// Insert 5 entries.
	for i := 0; i < 5; i++ {
		pt := now.Add(-time.Duration(i) * time.Hour)
		entry := &models.Entry{
			FeedID:      feed.ID,
			Title:       "Count Entry " + string(rune('A'+i)),
			URL:         "https://example.com/cs/" + string(rune('0'+i)),
			GUID:        "count-state-guid-" + string(rune('0'+i)),
			Content:     "count state test content",
			Description: "Count state test",
			PublishedAt: &pt,
			Hash:        "count-state-hash-" + string(rune('0'+i)),
		}
		if err := repo.InsertOrIgnore(entry); err != nil {
			t.Fatalf("InsertOrIgnore %d failed: %v", i, err)
		}
	}

	allEntries, err := repo.ListEntries(EntryQuery{PageSize: 20})
	if err != nil {
		t.Fatalf("ListEntries failed: %v", err)
	}

	// Set "processed" on entries 0, 2, 4 (3 entries). Others get "seen".
	for i, e := range allEntries {
		if i%2 == 0 {
			if err := stateRepo.SetState(e.ID, "processed"); err != nil {
				t.Fatalf("SetState processed on %d failed: %v", i, err)
			}
		} else {
			if err := stateRepo.SetState(e.ID, "seen"); err != nil {
				t.Fatalf("SetState seen on %d failed: %v", i, err)
			}
		}
	}

	// Count without state filter: all 5.
	total, err := repo.CountEntries(EntryQuery{})
	if err != nil {
		t.Fatalf("CountEntries (no filter) failed: %v", err)
	}
	if total != 5 {
		t.Errorf("total count: expected 5, got %d", total)
	}

	// Count only processed: 3 (entries 0, 2, 4).
	processedCount, err := repo.CountEntries(EntryQuery{State: "processed"})
	if err != nil {
		t.Fatalf("CountEntries (state=processed) failed: %v", err)
	}
	if processedCount != 3 {
		t.Errorf("processed count: expected 3, got %d", processedCount)
	}

	// Count only seen: 2.
	seenCount, err := repo.CountEntries(EntryQuery{State: "seen"})
	if err != nil {
		t.Fatalf("CountEntries (state=seen) failed: %v", err)
	}
	if seenCount != 2 {
		t.Errorf("seen count: expected 2, got %d", seenCount)
	}

	// State filter for a state not used: 0.
	zeroCount, err := repo.CountEntries(EntryQuery{State: "failed"})
	if err != nil {
		t.Fatalf("CountEntries (state=failed) failed: %v", err)
	}
	if zeroCount != 0 {
		t.Errorf("failed count: expected 0, got %d", zeroCount)
	}
}

func TestEntryRepo_BatchInsertEntries(t *testing.T) {
	database := setupTestDB(t)
	feed := createTestFeed(t, database, "https://example.com/batch.xml")
	repo := NewEntryRepo(database)

	now := time.Now()

	// Create 5 entries
	entries := make([]*models.Entry, 5)
	for i := 0; i < 5; i++ {
		pt := now.Add(-time.Duration(i) * time.Hour)
		entries[i] = &models.Entry{
			FeedID:      feed.ID,
			Title:       "Batch Entry " + string(rune('A'+i)),
			URL:         "https://example.com/b/" + string(rune('0'+i)),
			GUID:        "batch-guid-" + string(rune('0'+i)),
			Content:     "Batch content " + string(rune('0'+i)),
			Description: "Batch description " + string(rune('0'+i)),
			PublishedAt: &pt,
			Hash:        "batch-hash-" + string(rune('0'+i)),
		}
	}

	// Batch insert all 5
	count, err := repo.BatchInsertEntries(entries)
	if err != nil {
		t.Fatalf("BatchInsertEntries failed: %v", err)
	}
	if count != 5 {
		t.Errorf("expected 5 new entries, got %d", count)
	}

	// Verify all 5 persisted
	results, err := repo.ListEntries(EntryQuery{PageSize: 100})
	if err != nil {
		t.Fatalf("ListEntries failed: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("expected 5 entries in DB, got %d", len(results))
	}

	// Insert the same entries again — should insert 0 (all duplicates)
	count, err = repo.BatchInsertEntries(entries)
	if err != nil {
		t.Fatalf("BatchInsertEntries (dup) failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 new entries for duplicates, got %d", count)
	}

	// Verify still only 5 entries
	results, err = repo.ListEntries(EntryQuery{PageSize: 100})
	if err != nil {
		t.Fatalf("ListEntries (after dup) failed: %v", err)
	}
	if len(results) != 5 {
		t.Errorf("expected still 5 entries after dup, got %d", len(results))
	}

	// Verify GetByID returns full content
	got, err := repo.GetByID(results[0].ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Content == "" {
		t.Error("GetByID should return full content, but Content is empty")
	}

	// Verify ListEntries does NOT return content (light scan)
	if results[0].Content != "" {
		t.Error("ListEntries should NOT return content (light scan), but Content is non-empty")
	}
}

func TestEntryRepo_ListEntries_ContentNotReturned(t *testing.T) {
	database := setupTestDB(t)
	feed := createTestFeed(t, database, "https://example.com/content-check.xml")
	repo := NewEntryRepo(database)

	now := time.Now()
	entry := &models.Entry{
		FeedID:      feed.ID,
		Title:       "Content Check",
		URL:         "https://example.com/content-check",
		GUID:        "content-check-guid",
		Content:     "This is the full article content that should NOT appear in list results",
		Description: "Short description",
		PublishedAt: &now,
		Hash:        "content-check-hash",
	}

	if err := repo.InsertOrIgnore(entry); err != nil {
		t.Fatalf("InsertOrIgnore failed: %v", err)
	}

	// ListEntries should not include content
	results, err := repo.ListEntries(EntryQuery{PageSize: 100})
	if err != nil {
		t.Fatalf("ListEntries failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 entry")
	}
	if results[0].Content != "" {
		t.Error("ListEntries should NOT return content field")
	}

	// GetByID should include full content
	got, err := repo.GetByID(results[0].ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if got.Content != "This is the full article content that should NOT appear in list results" {
		t.Errorf("GetByID should return full content, got: %q", got.Content)
	}
}

// TestEntryRepo_ListByFeed_ContentNotReturned verifies that ListByFeed
// does not include the content field (uses entryListCols).
func TestEntryRepo_ListByFeed_ContentNotReturned(t *testing.T) {
	database := setupTestDB(t)
	feed := createTestFeed(t, database, "https://example.com/lbf-content.xml")
	repo := NewEntryRepo(database)

	now := time.Now()
	entry := &models.Entry{
		FeedID:      feed.ID,
		Title:       "LBF Content Check",
		URL:         "https://example.com/lbf-content",
		GUID:        "lbf-content-guid",
		Content:     "ListByFeed should not return this content",
		PublishedAt: &now,
		Hash:        "lbf-content-hash",
	}

	if err := repo.InsertOrIgnore(entry); err != nil {
		t.Fatalf("InsertOrIgnore failed: %v", err)
	}

	results, err := repo.ListByFeed(feed.ID, 10, 0)
	if err != nil {
		t.Fatalf("ListByFeed failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least 1 entry")
	}
	if results[0].Content != "" {
		t.Error("ListByFeed should NOT return content field")
	}
}

func TestEntryRepo_DeleteOlderThan(t *testing.T) {
	database := setupTestDB(t)
	feed := createTestFeed(t, database, "https://example.com/prune.xml")
	repo := NewEntryRepo(database)

	// Insert entries with explicit created_at times via raw SQL.
	// "old1" and "old2" are 10 days old; "new" is recent.
	type insert struct {
		guid      string
		createdAt string
		hash      string
	}
	entries := []insert{
		{"prune-old-1", time.Now().Add(-10 * 24 * time.Hour).Format("2006-01-02 15:04:05"), "prune-hash-old1"},
		{"prune-old-2", time.Now().Add(-10 * 24 * time.Hour).Format("2006-01-02 15:04:05"), "prune-hash-old2"},
		{"prune-new", time.Now().Format("2006-01-02 15:04:05"), "prune-hash-new"},
	}

	for _, e := range entries {
		_, err := database.Conn.Exec(`
			INSERT INTO entries (feed_id, title, url, guid, hash, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, feed.ID, "Prune Test", "https://example.com/prune/"+e.guid, e.guid, e.hash, e.createdAt)
		if err != nil {
			t.Fatalf("insert entry %s failed: %v", e.guid, err)
		}
	}

	// Before pruning: 3 entries
	all, err := repo.ListEntries(EntryQuery{PageSize: 100})
	if err != nil {
		t.Fatalf("ListEntries failed: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 entries before prune, got %d", len(all))
	}

	// Prune entries older than 7 days: should delete "old1" and "old2" (10 days old)
	deleted, err := repo.DeleteOlderThan(7)
	if err != nil {
		t.Fatalf("DeleteOlderThan failed: %v", err)
	}
	if deleted != 2 {
		t.Errorf("expected 2 deleted entries, got %d", deleted)
	}

	// Verify only the new entry remains
	remaining, err := repo.ListEntries(EntryQuery{PageSize: 100})
	if err != nil {
		t.Fatalf("ListEntries failed: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected 1 entry after prune, got %d", len(remaining))
	}
	if remaining[0].GUID != "prune-new" {
		t.Errorf("expected remaining entry GUID 'prune-new', got %q", remaining[0].GUID)
	}

	// Prune with 0 days: should not error but return 0 (no entries older than "now")
	// Since --days must be >= 1 in the CLI, but 0 should be handled safely
	zero, err := repo.DeleteOlderThan(0)
	if err != nil {
		t.Fatalf("DeleteOlderThan(0) failed: %v", err)
	}
	if zero != 0 {
		t.Errorf("expected 0 deletions for 0-day retention, got %d", zero)
	}
}
