package models

import (
	"time"
)

type Screening struct {
	// ScrapeID is supposed to identify the scrape that generated this screening
	// information. All screenings within a scrape share this ID.
	//
	// For now, the UNIX epoch from the begining of the scrape process is taken.
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
