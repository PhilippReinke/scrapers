package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/PhilippReinke/scrapers/models"
	"github.com/PhilippReinke/scrapers/repositories/screening"

	"github.com/labstack/echo/v4"
)

func main() {
	dbPath := flag.String("db", "assets/data.db", "path to sqlite db file")
	staticPath := flag.String("static", "web/static", "path to static website folder")
	templatePath := flag.String("template", "web/templates", "path to templates folder")
	flag.Parse()

	screeningRepo, err := loadScreeningRepo(*dbPath)
	if err != nil {
		slog.Error("Failed to load screenings repo.", "err", err)
		os.Exit(1)
	}

	e := setupEchoServer(screeningRepo, *staticPath, *templatePath)
	slog.Info("Serving webpage...", "url", ":8081")
	if err := e.Start(":8081"); err != nil {
		slog.Error("Echo server failed.", "err", err)
		os.Exit(1)
	}
}

func loadScreeningRepo(dbPath string) (screening.Repo, error) {
	repo, err := screening.NewSQLiteRepo(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite: %v", err)
	}

	return repo, nil
}

func setupEchoServer(screeningRepo screening.Repo, staticPath, templatePath string) *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.HidePort = true

	e.Use(slogMiddleware(slog.Default()))

	// setup template renderer
	templateRegex := filepath.Join(templatePath, "*.html")
	templates := &Template{
		templates: template.Must(template.ParseGlob(templateRegex)),
	}
	e.Renderer = templates

	// disable caching
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
			return next(c)
		}
	})

	// /api/selects endpoint
	e.GET("/api/selects", func(c echo.Context) error {
		filterOptions := screeningRepo.FilterOptions()
		return c.Render(http.StatusOK, "selects", filterOptions)
	})

	// /api/screenings endpoint
	e.POST("/api/screenings", func(c echo.Context) error {
		filter := models.Filter{
			ScrapeID: c.FormValue("scrape-ids"),
			Date:     c.FormValue("dates"),
			Cinema:   c.FormValue("cinemas"),
		}
		screenings, err := screeningRepo.QueryWithFilter(filter)
		if err != nil {
			return c.Render(http.StatusInternalServerError, "error", err.Error())
		}
		return c.Render(http.StatusOK, "screenings", screenings)
	})

	// serve content of static folder
	e.Static("/", staticPath)

	return e
}
