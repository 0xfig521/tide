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
	IsRead      bool       `json:"is_read"`
	IsStarred   bool       `json:"is_starred"`
	Hash        string     `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	// Runtime fields
	FeedTitle string `json:"feed_title,omitempty"`
}

// EntryOutput is the JSON-serializable view of an entry for CLI output.
type EntryOutput struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Author      string `json:"author,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
	FeedTitle   string `json:"feed_title,omitempty"`
	FeedID      int64  `json:"feed_id"`
	Description string `json:"description,omitempty"`
	Content     string `json:"content,omitempty"`
	Categories  string `json:"categories,omitempty"`
	GUID        string `json:"guid,omitempty"`
	IsRead      bool   `json:"is_read"`
	IsStarred   bool   `json:"is_starred"`
}
