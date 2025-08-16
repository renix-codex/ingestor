package models

import "time"

type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

type EnrichedPost struct {
	UserID     int       `json:"userId"`
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	IngestedAt time.Time `json:"ingested_at"`
	Source     string    `json:"source"`
}
