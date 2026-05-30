package fetcher

import (
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/0xfig521/tide/internal/db"
	"github.com/0xfig521/tide/internal/models"
	"github.com/0xfig521/tide/internal/repo"
)

// Scheduler periodically pulls feeds that are due for checking.
type Scheduler struct {
	db           *db.DB
	pool         *Pool
	cfg          Config
	batchSize    int
	limitPerHost int
	stopCh       chan struct{}
	stopped      bool
	mu           sync.Mutex
}

// NewScheduler creates a new scheduler.
func NewScheduler(db *db.DB, pool *Pool, cfg Config, batchSize, limitPerHost int) *Scheduler {
	return &Scheduler{
		db:           db,
		pool:         pool,
		cfg:          cfg,
		batchSize:    batchSize,
		limitPerHost: limitPerHost,
		stopCh:       make(chan struct{}),
	}
}

// Run starts the scheduler loop. Runs until Stop() is called.
func (s *Scheduler) Run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("[scheduler] started, batch=%d, limit_per_host=%d, interval=%s",
		s.batchSize, s.limitPerHost, interval)

	// Run immediately on start
	s.tick()

	for {
		select {
		case <-ticker.C:
			s.tick()
		case <-s.stopCh:
			log.Printf("[scheduler] stopped")
			return
		}
	}
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.stopped {
		s.stopped = true
		close(s.stopCh)
	}
}

func (s *Scheduler) tick() {
	feedRepo := repo.NewFeedRepo(s.db)

	feeds, err := feedRepo.GetDueFeeds(s.batchSize)
	if err != nil {
		log.Printf("[scheduler] failed to get due feeds: %v", err)
		return
	}

	if len(feeds) == 0 {
		return
	}

	// Apply per-host rate limiting
	jobs := s.buildBatch(feeds)
	if len(jobs) == 0 {
		return
	}

	log.Printf("[scheduler] dispatching %d jobs (%d feeds skipped by host limit)",
		len(jobs), len(feeds)-len(jobs))
	s.pool.Push(jobs)
}

func (s *Scheduler) buildBatch(feeds []*models.Feed) []FetchJob {
	hosts := make(map[string]int)
	var jobs []FetchJob

	for _, f := range feeds {
		hostname := extractHostname(f.FeedURL)
		if hosts[hostname] >= s.limitPerHost {
			continue
		}
		hosts[hostname]++

		lastFetchedAt := ""
		if f.LastFetchedAt != nil {
			lastFetchedAt = f.LastFetchedAt.Format("2006-01-02 15:04:05")
		}

		jobs = append(jobs, FetchJob{
			FeedID:        f.ID,
			FeedURL:       f.FeedURL,
			ETag:          f.ETagHeader,
			LastModified:  f.LastModifiedHeader,
			LastFetchedAt: lastFetchedAt,
		})
	}
	return jobs
}

func extractHostname(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Hostname()
}

// BatchSize returns the configured batch size.
func (s *Scheduler) BatchSize() int {
	return s.batchSize
}

// LimitPerHost returns the configured host limit.
func (s *Scheduler) LimitPerHost() int {
	return s.limitPerHost
}

// SetBatchSize updates the batch size.
func (s *Scheduler) SetBatchSize(n int) {
	s.batchSize = n
}

// SetLimitPerHost updates the per-host limit.
func (s *Scheduler) SetLimitPerHost(n int) {
	s.limitPerHost = n
}
