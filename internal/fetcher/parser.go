package fetcher

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

// Config holds fetch-related configuration.
type Config struct {
	Timeout          time.Duration
	MaxIdleConns     int
	MaxConnsPerHost  int
	UserAgent        string
	MinFetchInterval time.Duration
	ForceRefresh     bool
	CheckInterval    time.Duration // how often next_check_at is set after success
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Timeout:          30 * time.Second,
		MaxIdleConns:     100,
		MaxConnsPerHost:  10,
		UserAgent:        "GoRSS/1.0",
		MinFetchInterval: 30 * time.Minute,
		ForceRefresh:     false,
		CheckInterval:    60 * time.Minute,
	}
}

// Parser wraps gofeed.Parser with a shared HTTP client.
// Each goroutine creates its own Parser (gofeed.Parser is not thread-safe),
// but they all share the same *http.Client for connection pooling.
type Parser struct {
	HTTP      *http.Client
	UserAgent string
	Timeout   time.Duration
}

// NewParser creates a new Parser with the given HTTP client.
func (c Config) NewParser() *Parser {
	return &Parser{
		HTTP:      newHTTPClient(c),
		UserAgent: c.UserAgent,
		Timeout:   c.Timeout,
	}
}

// Fetch fetches and parses a feed from the given URL.
func (p *Parser) Fetch(feedURL, etag, lastModified string) (*gofeed.Feed, string, string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), p.Timeout)
	defer cancel()

	fp := gofeed.NewParser()
	fp.Client = p.HTTP
	fp.UserAgent = p.UserAgent

	// Use custom request to set conditional headers
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, "", "", 0, fmt.Errorf("create request: %w", err)
	}
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}
	req.Header.Set("User-Agent", p.UserAgent)
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml, application/feed+json, application/json")

	resp, err := p.HTTP.Do(req)
	if err != nil {
		return nil, "", "", 0, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode

	// 304 Not Modified - feed unchanged
	if statusCode == http.StatusNotModified {
		return nil, "", "", statusCode, nil
	}

	if statusCode != http.StatusOK {
		return nil, "", "", statusCode, fmt.Errorf("http status %d", statusCode)
	}

	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, "", "", statusCode, fmt.Errorf("parse feed: %w", err)
	}

	newETag := resp.Header.Get("ETag")
	newLastModified := resp.Header.Get("Last-Modified")

	return feed, newETag, newLastModified, statusCode, nil
}

func newHTTPClient(c Config) *http.Client {
	return &http.Client{
		Timeout: c.Timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        c.MaxIdleConns,
			MaxIdleConnsPerHost: c.MaxConnsPerHost,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}
