package fetcher

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/0xfig-labs/tide/internal/db"
	"github.com/0xfig-labs/tide/internal/models"
	"github.com/0xfig-labs/tide/internal/repo"
	"github.com/0xfig-labs/tide/pkg"
)

// Worker processes fetch jobs.
type Worker struct {
	db     *db.DB
	cfg    Config
	parser *Parser
}

func (w *Worker) run(jobs <-chan FetchJob, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		if err := w.processJob(job); err != nil {
			log.Printf("[worker] feed %d fetch error: %v", job.FeedID, err)
		}
	}
}

func (w *Worker) processJob(job FetchJob) error {
	feedRepo := repo.NewFeedRepo(w.db)
	entryRepo := repo.NewEntryRepo(w.db)

	// Check cache: skip if within min fetch interval (unless force refresh)
	if !w.cfg.ForceRefresh && job.LastFetchedAt != "" {
		lastFetch, err := time.Parse("2006-01-02 15:04:05", job.LastFetchedAt)
		if err == nil && time.Since(lastFetch) < w.cfg.MinFetchInterval {
			log.Printf("[worker] feed %d skipped (cache): last fetched %s ago", job.FeedID, time.Since(lastFetch).Round(time.Second))
			return nil
		}
	}

	// Fetch feed
	feed, etag, lastModified, statusCode, err := w.parser.Fetch(job.FeedURL, job.ETag, job.LastModified)
	if err != nil {
		log.Printf("[worker] feed %d (%s) fetch error: %v", job.FeedID, job.FeedURL, err)
		if recordErr := feedRepo.UpdateFetchError(job.FeedID, err.Error(), statusCode); recordErr != nil {
			log.Printf("[worker] feed %d failed to record error: %v", job.FeedID, recordErr)
		}
		return err
	}

	now := time.Now()
	nextCheck := now.Add(w.cfg.CheckInterval)

	// Handle 304 Not Modified
	if statusCode == http.StatusNotModified {
		if err := feedRepo.UpdateFetchResult(job.FeedID, etag, lastModified, statusCode, now, nextCheck); err != nil {
			log.Printf("[worker] feed %d failed to update fetch result: %v", job.FeedID, err)
		}
		log.Printf("[worker] feed %d not modified (304)", job.FeedID)
		return nil
	}

	// Handle nil feed (shouldn't happen with OK status, but be safe)
	if feed == nil {
		return nil
	}

	// Update feed metadata
	if err := feedRepo.UpdateMeta(job.FeedID, feed.Title, feed.Description, feed.Link, imageURL(feed), feed.Language, feed.FeedType); err != nil {
		log.Printf("[worker] feed %d failed to update meta: %v", job.FeedID, err)
	}

	// Update fetch result
	if err := feedRepo.UpdateFetchResult(job.FeedID, etag, lastModified, statusCode, now, nextCheck); err != nil {
		log.Printf("[worker] feed %d failed to update fetch result: %v", job.FeedID, err)
	}

	// Batch insert entries with dedup (single transaction for performance)
	batch := make([]*models.Entry, 0, len(feed.Items))
	for _, item := range feed.Items {
		batch = append(batch, convertEntry(job.FeedID, item))
	}
	newCount, err := entryRepo.BatchInsertEntries(batch)
	if err != nil {
		log.Printf("[worker] feed %d batch insert error: %v", job.FeedID, err)
	}

	if newCount > 0 {
		log.Printf("[worker] feed %d (%s): %d new entries", job.FeedID, truncateStr(feed.Title, 30), newCount)
	}
	return nil
}

func convertEntry(feedID int64, item *gofeed.Item) *models.Entry {
	return ConvertEntry(feedID, item)
}

// ConvertEntry converts a gofeed.Item to a models.Entry with dedup hash.
func ConvertEntry(feedID int64, item *gofeed.Item) *models.Entry {
	e := &models.Entry{
		FeedID:      feedID,
		Title:       item.Title,
		URL:         item.Link,
		GUID:        item.GUID,
		Content:     item.Content,
		Description: item.Description,
		ImageURL:    imageURLFromItem(item),
	}

	if item.PublishedParsed != nil {
		e.PublishedAt = item.PublishedParsed
	} else if item.UpdatedParsed != nil {
		e.PublishedAt = item.UpdatedParsed
	}

	if len(item.Authors) > 0 {
		e.AuthorName = item.Authors[0].Name
	} else if item.Author != nil {
		e.AuthorName = item.Author.Name
	}

	if len(item.Categories) > 0 {
		e.Categories = strings.Join(item.Categories, ",")
	}

	e.Hash = pkg.EntryHash(feedID, e.GUID)
	return e
}

func imageURL(feed *gofeed.Feed) string {
	return ImageURL(feed)
}

// ImageURL extracts the image URL from a gofeed.Feed.
func ImageURL(feed *gofeed.Feed) string {
	if feed.Image != nil && feed.Image.URL != "" {
		return feed.Image.URL
	}
	return ""
}

func imageURLFromItem(item *gofeed.Item) string {
	if item.Image != nil && item.Image.URL != "" {
		return item.Image.URL
	}
	return ""
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
