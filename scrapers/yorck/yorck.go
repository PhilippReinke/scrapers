package yorck

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/PhilippReinke/scrapers/models"
	"github.com/PhilippReinke/scrapers/storage/screening"
)

const (
	yorckAdress    = "https://www.yorck.de/filme"
	scriptTagBegin = "<script id=\"__NEXT_DATA__\" type=\"application/json\">"
	scriptTagEnd   = "</script>"
)

type Scraper struct {
	l    *slog.Logger
	repo screening.Repo
}

func New(repo screening.Repo, logger *slog.Logger) *Scraper {
	return &Scraper{
		l:    logger,
		repo: repo,
	}
}

func (s *Scraper) Run() error {
	res, err := http.Get(yorckAdress)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bodyByte, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	body := string(bodyByte)

	begin := strings.Index(body, scriptTagBegin)
	if begin == -1 {
		return fmt.Errorf("failed to find begin of film data")
	}
	begin += len(scriptTagBegin)
	end := strings.Index(body[begin:], scriptTagEnd)
	if end == -1 {
		return fmt.Errorf("failed to find end of film data")
	}

	jsonString := body[begin : begin+end]

	var films FilmsYorck
	if err := json.Unmarshal([]byte(jsonString), &films); err != nil {
		return err
	}

	var cnt int
	now := time.Now()
	for _, film := range films.Props.PageProps.Films {
		for _, session := range film.Fields.Sessions {
			s.repo.Insert(models.Screening{
				ID:            buildID(session.Fields.Cinema.Fields.Name, now, cnt),
				ScrapeID:      now.Unix(),
				Title:         film.Fields.Title,
				Date:          session.Fields.StartTime,
				Duration:      film.Fields.Runtime,
				Cinema:        session.Fields.Cinema.Fields.Name,
				ThumbnailLink: fmt.Sprintf("https:%v?w=480&q=75", film.Fields.HeroImage.Fields.Image.FieldsImage.File.URL),
				Description:   "todo",
				Link:          fmt.Sprintf("%v/%v", yorckAdress, film.Fields.Slug),
			})
			cnt++
		}
	}

	return nil
}

func buildID(cinema string, now time.Time, cnt int) string {
	cinemaNormalised := strings.ToLower(strings.ReplaceAll(cinema, " ", ""))
	return fmt.Sprintf("%v-%v-%v", cinemaNormalised, now.Unix(), cnt)
}
