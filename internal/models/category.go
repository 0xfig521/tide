package models

import "time"

// Category represents a feed category.
type Category struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// Runtime fields
	FeedCount int `json:"feed_count"`
}
