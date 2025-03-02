package screening

import (
	"fmt"
	"log/slog"
	"os"
	"sort"
	"time"

	"github.com/PhilippReinke/scrapers/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SQLite struct {
	db *gorm.DB
}

var _ Repo = (*SQLite)(nil) // interface guard

func NewSQLiteRepo(dbPath string) (*SQLite, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		slog.Error("failed to open SQLite database.", "err", err)
		os.Exit(1)
	}

	if err := db.AutoMigrate(&models.Screening{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema")
	}

	return &SQLite{
		db: db,
	}, nil
}

func (s *SQLite) QueryAll() ([]models.Screening, error) {
	var screenings []models.Screening
	result := s.db.Find(&screenings)
	return screenings, result.Error
}

func (s *SQLite) QueryWithFilter(filter models.Filter) ([]models.Screening, error) {
	var screenings []models.Screening

	if filter.ScrapeID == "" {
		// If no scrape id has been selected we return the results of the latest
		// scrape by cinema.
		return s.queryLatestByCinema(filter)
	}

	req := s.db.Where("scrape_id = ?", filter.ScrapeID)

	// only results from today onwards
	req = req.Where("DATE(date) >= ?", time.Now().Format(time.DateOnly))

	if filter.Date != "" {
		req = req.Where("DATE(date) = ?", filter.Date)
	}
	if filter.Cinema != "" {
		req = req.Where("cinema = ?", filter.Cinema)
	}
	result := req.Find(&screenings)

	return screenings, result.Error
}

func (s *SQLite) queryLatestByCinema(filter models.Filter) ([]models.Screening, error) {
	var cinemasToQuery []string
	if filter.Cinema == "" {
		// We wanna query all cinemas in this case.
		s.db.Model(&models.Screening{}).
			Distinct("cinema").
			Pluck("cinema", &cinemasToQuery) // TODO: error handling.
	} else {
		cinemasToQuery = append(cinemasToQuery, filter.Cinema)
	}

	var screenings []models.Screening
	for _, cinema := range cinemasToQuery {
		// find latest scrape id for cinema
		var latest models.Screening
		s.db.Where("cinema = ?", cinema).Last(&latest) // TODO: error handling.
		latestScrapeID := latest.ScrapeID

		var screeningsTemp []models.Screening
		req := s.db.
			Where("scrape_id = ?", latestScrapeID).
			Where("cinema = ?", cinema)
		if filter.Date != "" {
			req = req.Where("DATE(date) = ?", filter.Date)
		}

		// only results from today onwards
		req = req.Where("DATE(date) >= ?", time.Now().Format(time.DateOnly))

		result := req.Order("date asc").Find(&screeningsTemp)
		_ = result // TODO: error handling.
		screenings = append(screenings, screeningsTemp...)
	}

	sort.Slice(screenings, func(i, j int) bool {
		return screenings[i].Date.Unix() < screenings[j].Date.Unix()
	})

	return screenings, nil
}

func (s *SQLite) Insert(screening models.Screening) error {
	result := s.db.Create(&screening)
	return result.Error
}

func (s *SQLite) FilterOptions() models.FilterOptions {
	nowDate := time.Now().Format(time.DateOnly)

	var scrapeIDs []int64
	s.db.Model(&models.Screening{}).
		Distinct("scrape_id").
		Where("DATE(date) >= ?", nowDate).
		Order("scrape_id desc").
		Pluck("scrape_id", &scrapeIDs)

	var unqiueDates []time.Time
	s.db.Model(&models.Screening{}).
		Distinct("date").
		Where("DATE(date) >= ?", nowDate).
		Order("date asc").
		Pluck("date", &unqiueDates)

	var uniqueCinemas []string
	s.db.Model(&models.Screening{}).
		Distinct("cinema").
		Where("DATE(date) >= ?", nowDate).
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
