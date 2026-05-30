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
	Categories  []string `json:"categories,omitempty"`
	EntryCount  int      `json:"entry_count"`
	UnreadCount int      `json:"unread_count"`
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
	UnreadCount int      `json:"unread_count"`
	LastFetched string   `json:"last_fetched_at,omitempty"`
	IsActive    bool     `json:"is_active"`
}
