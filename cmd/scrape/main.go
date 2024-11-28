package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/PhilippReinke/scrapers/scrapers/babylon"
	"github.com/PhilippReinke/scrapers/scrapers/yorck"
	"github.com/PhilippReinke/scrapers/storage/screening"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	dbPath := flag.String("db", "assets/data.db", "path to sqlite db file")
	flag.Parse()

	db, err := gorm.Open(sqlite.Open(*dbPath), &gorm.Config{})
	if err != nil {
		slog.Error("Could not open SQLite database.", "err", err)
		os.Exit(1)
	}

	repo, err := screening.NewSQLiteRepo(db)
	if err != nil {
		slog.Error("Could not create screening repository.", "err", err)
		os.Exit(1)
	}

	babylonScraper := babylon.New(repo, slog.Default())
	if err := babylonScraper.Run(); err != nil {
		slog.Error("Scraping Kino Babylon failed.", "err", err)
	} else {
		slog.Info("Scraping Kino Babylon succeeded.")
	}

	yorckScraper := yorck.New(repo, slog.Default())
	if err := yorckScraper.Run(); err != nil {
		slog.Error("Scraping Yorck Kino failed.", "err", err)
	} else {
		slog.Info("Scraping Yorck Kino succeeded.")
	}
}
