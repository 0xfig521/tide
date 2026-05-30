package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"

	"github.com/0xfig521/tide/internal/fetcher"
	"github.com/0xfig521/tide/internal/output"
	"github.com/0xfig521/tide/internal/repo"
)

var (
	fetchFeedID      int64
	fetchCategory    string
	fetchConcurrency int
	fetchForce       bool
	fetchDaemon      bool
	fetchInterval    time.Duration
	fetchQuiet       bool
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch articles from RSS feeds",
	Long: `Fetch articles from all or specified RSS feeds.

When run without flags, fetches all due feeds (those with expired next_check_at).
Use --feed to fetch a specific feed, --category for a group of feeds.
Use --daemon to run as a background scheduler.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := fetcher.DefaultConfig()
		cfg.ForceRefresh = fetchForce
		parser := cfg.NewParser()

		if fetchDaemon {
			runDaemon(cfg)
			return nil
		}

		jobs := buildJobList()
		if len(jobs) == 0 {
			output.PrintSuccess(map[string]any{"message": "no feeds to fetch"}, nil)
			return nil
		}

		var bar *progressbar.ProgressBar
		if !fetchQuiet {
			bar = progressbar.NewOptions(len(jobs),
				progressbar.OptionSetDescription("Fetching"),
				progressbar.OptionSetWidth(30),
				progressbar.OptionShowCount(),
				progressbar.OptionShowIts(),
				progressbar.OptionSetItsString("feeds"),
				progressbar.OptionThrottle(100*time.Millisecond),
				progressbar.OptionSetPredictTime(true),
				progressbar.OptionSetWriter(os.Stderr),
			)
		}

		feedRepo := repo.NewFeedRepo(dbConn)
		entryRepo := repo.NewEntryRepo(dbConn)

		jobCh := make(chan fetcher.FetchJob, len(jobs))
		for _, j := range jobs {
			jobCh <- j
		}
		close(jobCh)

		var wg sync.WaitGroup
		var newEntries, failedFeeds, unchanged atomic.Int64

		for range fetchConcurrency {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobCh {
					feed, etag, lastModified, statusCode, fetchErr := parser.Fetch(job.FeedURL, job.ETag, job.LastModified)

					if fetchErr != nil {
						feedRepo.UpdateFetchError(job.FeedID, fetchErr.Error())
						failedFeeds.Add(1)
						if bar != nil {
							bar.Add(1)
						}
						continue
					}

					now := time.Now()
					nextCheck := now.Add(cfg.CheckInterval)

					if statusCode == 304 {
						feedRepo.UpdateFetchResult(job.FeedID, etag, lastModified, statusCode, now, nextCheck)
						unchanged.Add(1)
						if bar != nil {
							bar.Add(1)
						}
						continue
					}

					if feed != nil {
						feedRepo.UpdateMeta(job.FeedID, feed.Title, feed.Description, feed.Link,
							fetcher.ImageURL(feed), feed.Language, feed.FeedType)
						feedRepo.UpdateFetchResult(job.FeedID, etag, lastModified, statusCode, now, nextCheck)

						for _, item := range feed.Items {
							entry := fetcher.ConvertEntry(job.FeedID, item)
							if err := entryRepo.InsertOrIgnore(entry); err == nil {
								newEntries.Add(1)
							}
						}
					}
					if bar != nil {
						bar.Add(1)
					}
				}
			}()
		}
		wg.Wait()
		if bar != nil {
			bar.Finish()
		}

		output.PrintSuccess(map[string]any{
			"feeds_fetched": len(jobs),
			"new_entries":   newEntries.Load(),
			"unchanged":     unchanged.Load(),
			"failed":        failedFeeds.Load(),
		}, nil)
		return nil
	},
}

func buildJobList() []fetcher.FetchJob {
	var jobs []fetcher.FetchJob
	fr := feedRepo()

	if fetchFeedID > 0 {
		f, err := fr.GetByID(fetchFeedID)
		if err != nil {
			return jobs
		}
		lastFetchedAt := ""
		if f.LastFetchedAt != nil {
			lastFetchedAt = f.LastFetchedAt.Format("2006-01-02 15:04:05")
		}
		jobs = append(jobs, fetcher.FetchJob{
			FeedID: f.ID, FeedURL: f.FeedURL,
			ETag: f.ETagHeader, LastModified: f.LastModifiedHeader,
			LastFetchedAt: lastFetchedAt,
		})
	} else if fetchCategory != "" {
		feeds, _ := fr.List(fetchCategory)
		for _, f := range feeds {
			lastFetchedAt := ""
			if f.LastFetchedAt != nil {
				lastFetchedAt = f.LastFetchedAt.Format("2006-01-02 15:04:05")
			}
			jobs = append(jobs, fetcher.FetchJob{
				FeedID: f.ID, FeedURL: f.FeedURL,
				ETag: f.ETagHeader, LastModified: f.LastModifiedHeader,
				LastFetchedAt: lastFetchedAt,
			})
		}
	} else {
		feeds, _ := fr.GetDueFeeds(100)
		for _, f := range feeds {
			lastFetchedAt := ""
			if f.LastFetchedAt != nil {
				lastFetchedAt = f.LastFetchedAt.Format("2006-01-02 15:04:05")
			}
			jobs = append(jobs, fetcher.FetchJob{
				FeedID: f.ID, FeedURL: f.FeedURL,
				ETag: f.ETagHeader, LastModified: f.LastModifiedHeader,
				LastFetchedAt: lastFetchedAt,
			})
		}
	}
	return jobs
}

func runDaemon(cfg fetcher.Config) {
	pool := fetcher.NewPool(dbConn, cfg)
	pool.Start(fetchConcurrency)
	defer pool.Shutdown()

	scheduler := fetcher.NewScheduler(dbConn, pool, cfg, 100, 2)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go scheduler.Run(fetchInterval)

	fmt.Println(output.Success(fmt.Sprintf("Daemon started. Interval: %s, workers: %d. Ctrl+C to stop.", fetchInterval, fetchConcurrency)))
	<-sigCh
	fmt.Println(output.Warn("\nShutting down..."))
	scheduler.Stop()
}

func init() {
	fetchCmd.Flags().Int64Var(&fetchFeedID, "feed", 0, "Fetch specific feed by ID")
	fetchCmd.Flags().StringVarP(&fetchCategory, "category", "c", "", "Fetch feeds in a category")
	fetchCmd.Flags().IntVarP(&fetchConcurrency, "concurrency", "n", 5, "Number of concurrent workers")
	fetchCmd.Flags().BoolVarP(&fetchForce, "force", "f", false, "Force refresh (ignore cache interval)")
	fetchCmd.Flags().BoolVar(&fetchDaemon, "daemon", false, "Run as daemon (continuous scheduler)")
	fetchCmd.Flags().DurationVar(&fetchInterval, "interval", 30*time.Minute, "Daemon fetch interval")
	fetchCmd.Flags().BoolVar(&fetchQuiet, "quiet", false, "Suppress progress bar output")
	rootCmd.AddCommand(fetchCmd)
}
