package models

import "time"

type Screening struct {
	ID            string `gorm:"primaryKey"`
	ScrapeID      int64
	Title         string
	Date          time.Time
	Duration      int
	Cinema        string
	ThumbnailLink string
	Description   string
	Link          string
}

type FilterOptions struct {
	ScrapeIDs []int64
	Dates     []time.Time
	Cinemas   []string
}

type Filter struct {
	ScrapeID string
	Date     string
	Cinema   string
}
