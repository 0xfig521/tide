package models

import "time"

// Entry represents a single article/item from a feed.
type Entry struct {
	ID          int64      `json:"id"`
	FeedID      int64      `json:"feed_id"`
	Title       string     `json:"title"`
	URL         string     `json:"url"`
	GUID        string     `json:"guid"`
	Content     string     `json:"content,omitempty"`
	Description string     `json:"description,omitempty"`
	AuthorName  string     `json:"author_name,omitempty"`
	ImageURL    string     `json:"image_url,omitempty"`
	Categories  string     `json:"categories,omitempty"` // JSON array string
	PublishedAt *time.Time `json:"published_at"`
	Hash        string     `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	// Runtime fields
	FeedTitle string `json:"feed_title,omitempty"`
}

// EntryState tracks an agent's processing state for a single entry.
// This replaces the read/star model with a more flexible state machine
// designed for agent workflows: new → seen → processed / ignored / failed.
type EntryState struct {
	EntryID     int64      `json:"entry_id"`
	State       string     `json:"state"`
	Tags        string     `json:"tags,omitempty"`
	Note        string     `json:"note,omitempty"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// EntryOutput is the JSON-serializable view of an entry for CLI output.
type EntryOutput struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Author      string `json:"author"`
	PublishedAt string `json:"published_at"`
	FeedTitle   string `json:"feed_title,omitempty"`
	FeedID      int64  `json:"feed_id"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Categories  string `json:"categories"`
	GUID        string `json:"guid,omitempty"`
}
