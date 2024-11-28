package screening

import (
	"fmt"
	"sort"
	"time"

	"github.com/PhilippReinke/scrapers/models"
	"gorm.io/gorm"
)

type SQLite struct {
	db *gorm.DB
}

var _ Repo = (*SQLite)(nil) // interface guard

func NewSQLiteRepo(db *gorm.DB) (*SQLite, error) {
	if err := db.AutoMigrate(&models.Screening{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema")
	}

	return &SQLite{
		db: db,
	}, nil
}

func (sr *SQLite) QueryAll() ([]models.Screening, error) {
	var screenings []models.Screening
	result := sr.db.Find(&screenings)
	return screenings, result.Error
}

func (sr *SQLite) QueryWithFilter(filter models.Filter) ([]models.Screening, error) {
	var screenings []models.Screening

	if filter.ScrapeID == "" {
		// If no scrape id has been selected we return the results of the latest
		// scrape by cinema.
		return sr.queryLatestByCinema(filter)
	}

	req := sr.db.Where("scrape_id = ?", filter.ScrapeID)
	if filter.Date != "" {
		req = req.Where("DATE(date) = ?", filter.Date)
	}
	if filter.Cinema != "" {
		req = req.Where("cinema = ?", filter.Cinema)
	}
	result := req.Find(&screenings)

	return screenings, result.Error
}

func (sr *SQLite) queryLatestByCinema(filter models.Filter) ([]models.Screening, error) {
	var cinemasToQuery []string
	if filter.Cinema == "" {
		// We wanna query all cinemas in this case.
		sr.db.Model(&models.Screening{}).
			Distinct("cinema").
			Pluck("cinema", &cinemasToQuery) // TODO: error handling.
	} else {
		cinemasToQuery = append(cinemasToQuery, filter.Cinema)
	}

	var screenings []models.Screening
	for _, cinema := range cinemasToQuery {
		// find latest scrape id for cinema
		var latest models.Screening
		sr.db.Where("cinema = ?", cinema).Last(&latest) // TODO: error handling.
		latestScrapeID := latest.ScrapeID

		var screeningsTemp []models.Screening
		req := sr.db.
			Where("scrape_id = ?", latestScrapeID).
			Where("cinema = ?", cinema)
		if filter.Date != "" {
			req = req.Where("DATE(date) = ?", filter.Date)
		}
		result := req.Order("date asc").Find(&screeningsTemp)
		_ = result // TODO: error handling.
		screenings = append(screenings, screeningsTemp...)
	}

	sort.Slice(screenings, func(i, j int) bool {
		return screenings[i].Date.Unix() < screenings[j].Date.Unix()
	})

	return screenings, nil
}

func (sr *SQLite) Insert(screening models.Screening) error {
	result := sr.db.Create(&screening)
	return result.Error
}

func (sr *SQLite) FilterOptions() models.FilterOptions {
	var scrapeIDs []int64
	sr.db.Model(&models.Screening{}).
		Distinct("scrape_id").
		Order("scrape_id desc").
		Pluck("scrape_id", &scrapeIDs)

	var unqiueDates []time.Time
	sr.db.Model(&models.Screening{}).
		Distinct("date").
		Order("date asc").
		Pluck("date", &unqiueDates)

	var uniqueCinemas []string
	sr.db.Model(&models.Screening{}).
		Distinct("cinema").
		Pluck("cinema", &uniqueCinemas)

	return models.FilterOptions{
		ScrapeIDs: scrapeIDs,
		Dates:     uniqueByDate(unqiueDates),
		Cinemas:   uniqueCinemas,
	}
}

func uniqueByDate(dates []time.Time) []time.Time {
	var datesUnique []time.Time
	dateMap := make(map[time.Time]struct{})

	for _, date := range dates {
		dateOnly := date.Truncate(24 * time.Hour)
		_, ok := dateMap[dateOnly]
		if !ok {
			dateMap[dateOnly] = struct{}{}
			datesUnique = append(datesUnique, dateOnly)
		}
	}
	return datesUnique
}
