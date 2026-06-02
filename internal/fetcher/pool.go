package fetcher

import (
	"sync"

	"github.com/0xfig521/tide/internal/db"
	"github.com/0xfig521/tide/internal/repo"
)

// FetchJob represents a feed to be fetched.
type FetchJob struct {
	FeedID        int64
	FeedURL       string
	ETag          string
	LastModified  string
	LastFetchedAt string
}

// Pool is a worker pool for concurrent feed fetching.
type Pool struct {
	jobs chan FetchJob
	wg   sync.WaitGroup
	db   *db.DB
	cfg  Config
}

// NewPool creates a new worker pool.
func NewPool(db *db.DB, cfg Config) *Pool {
	return &Pool{
		jobs: make(chan FetchJob, 1000),
		db:   db,
		cfg:  cfg,
	}
}

// Start launches N worker goroutines.
func (p *Pool) Start(n int) {
	for range n {
		p.wg.Add(1)
		w := &Worker{
			db:     p.db,
			cfg:    p.cfg,
			parser: p.cfg.NewParser(),
		}
		go w.run(p.jobs, &p.wg)
	}
}

// Push adds jobs to the pool. Blocks if the channel is full (backpressure).
func (p *Pool) Push(jobs []FetchJob) {
	for _, j := range jobs {
		p.jobs <- j
	}
}

// Close stops the pool gracefully. Closes the job channel and waits for workers.
func (p *Pool) Close() {
	close(p.jobs)
}

// Wait blocks until all workers finish.
func (p *Pool) Wait() {
	p.wg.Wait()
}

// Shutdown closes the pool and waits for all workers to finish.
func (p *Pool) Shutdown() {
	p.Close()
	p.Wait()
}

// FeedRepo returns a feed repository instance.
func (p *Pool) FeedRepo() *repo.FeedRepo {
	return repo.NewFeedRepo(p.db)
}

// EntryRepo returns an entry repository instance.
func (p *Pool) EntryRepo() *repo.EntryRepo {
	return repo.NewEntryRepo(p.db)
}
