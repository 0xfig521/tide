package models

import "time"

// Feed represents an RSS/Atom feed subscription.
type Feed struct {
	ID                 int64      `json:"id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	SiteURL            string     `json:"site_url"`
	FeedURL            string     `json:"feed_url"`
	ImageURL           string     `json:"image_url,omitempty"`
	Language           string     `json:"language,omitempty"`
	FeedType           string     `json:"feed_type,omitempty"`
	ETagHeader         string     `json:"-"`
	LastModifiedHeader string     `json:"-"`
	CheckedAt          *time.Time `json:"-"`
	LastFetchedAt      *time.Time `json:"last_fetched_at"`
	HTTPStatusCode     int        `json:"-"`
	NextCheckAt        time.Time  `json:"next_check_at"`
	ParsingErrorCount  int        `json:"parsing_error_count"`
	ParsingErrorMsg    string     `json:"-"`
	IsActive           bool       `json:"is_active"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	// Runtime fields (not from DB)
	Categories []string `json:"categories,omitempty"`
	EntryCount int      `json:"entry_count"`
}

// FeedOutput is the JSON-serializable view of a feed for CLI output.
type FeedOutput struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	FeedURL     string   `json:"feed_url"`
	SiteURL     string   `json:"site_url"`
	Description string   `json:"description,omitempty"`
	ImageURL    string   `json:"image_url,omitempty"`
	Categories  []string `json:"categories"`
	EntryCount  int      `json:"entry_count"`
	LastFetched string   `json:"last_fetched_at,omitempty"`
	IsActive    bool     `json:"is_active"`
}

// FailureType classifies the root cause of a feed fetch failure.
type FailureType string

const (
	FailureHTTP4xx FailureType = "http_4xx"
	FailureHTTP5xx FailureType = "http_5xx"
	FailureTimeout FailureType = "timeout"
	FailureDNS     FailureType = "dns"
	FailureTLS     FailureType = "tls"
	FailureParse   FailureType = "parse"
	FailureUnknown FailureType = "unknown"
)

// ValidFailureTypes is the canonical whitelist for FailureType values.
var ValidFailureTypes = map[FailureType]bool{
	FailureHTTP4xx: true,
	FailureHTTP5xx: true,
	FailureTimeout: true,
	FailureDNS:     true,
	FailureTLS:     true,
	FailureParse:   true,
	FailureUnknown: true,
}

// FeedFailure is one recorded fetch failure for a feed source.
type FeedFailure struct {
	ID           int64       `json:"id"`
	FeedID       int64       `json:"feed_id"`
	ErrorType    FailureType `json:"error_type"`
	ErrorMessage string      `json:"error_message"`
	HTTPStatus   int         `json:"http_status"`
	OccurredAt   string      `json:"occurred_at"`
}
