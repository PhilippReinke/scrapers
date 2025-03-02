package babylon

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/PhilippReinke/scrapers/models"
	"github.com/PhilippReinke/scrapers/repositories/screening"

	"github.com/gocolly/colly/v2"
)

const (
	babylonAdress = "https://babylonberlin.eu"
)

type Scraper struct {
	c    *colly.Collector
	l    *slog.Logger
	repo screening.Repo
}

func New(repo screening.Repo, logger *slog.Logger) *Scraper {
	return &Scraper{
		c:    colly.NewCollector(),
		l:    logger,
		repo: repo,
	}
}

func (s *Scraper) Run() error {
	var lastMonth, yearOffset int

	s.c.OnHTML("#regridart-207", func(e *colly.HTMLElement) {
		now := time.Now()

		e.ForEach("li", func(n int, e *colly.HTMLElement) {
			titles := e.ChildTexts("h3")
			if len(titles) <= 2 {
				return
			}

			date, err := parseDate(e.ChildTexts(".mix-date")[0], &lastMonth, &yearOffset)
			if err != nil {
				s.l.Error("Could not parse date.", "err", err)
			}

			duration, err := parseDuration(e.ChildTexts(".runtime")[0])
			if err != nil {
				s.l.Error("Could not parse duration for %v.", "err", err)
			}

			link := babylonAdress + e.ChildAttr(".mix-title", "href")

			s.repo.Insert(models.Screening{
				ScrapeID:      now.Unix(),
				Title:         e.ChildTexts("h3")[2],
				Date:          date,
				Duration:      duration,
				Cinema:        models.KinoBabylon,
				ThumbnailLink: e.ChildAttr(".fancybox", "href"),
				Description:   "todo",
				Link:          link,
			})
		})
	})
	s.c.OnRequest(func(r *colly.Request) {
		s.l.Debug("Running scraper...", "url", r.URL.String())
	})

	return s.c.Visit(babylonAdress + "/programm")
}

func parseDate(dateString string, lastMonth, yearOffset *int) (time.Time, error) {
	var dayAbbr string
	var day, month, hour, minute int

	_, err := fmt.Sscanf(dateString, "%3s %2d.%2d. %2d:%2d", &dayAbbr, &day, &month, &hour, &minute)
	if err != nil {
		return time.Time{}, err
	}

	// workaround
	year := time.Now().Year()
	if *lastMonth > month {
		*yearOffset++
	}
	*lastMonth = month

	tz, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to create timezone: %w", err)
	}

	return time.Date(year+*yearOffset, time.Month(month), day, hour, minute, 0, 0, tz), nil
}

func parseDuration(durationString string) (int, error) {
	var duration int

	_, err := fmt.Sscanf(durationString, "%d min.", &duration)
	if err != nil {
		return 0, err
	}

	return duration, nil
}
