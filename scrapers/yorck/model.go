package yorck

import "time"

type FilmsYorck struct {
	Props Props `json:"props"`
}

type Props struct {
	PageProps PageProps `json:"pageProps"`
}

type PageProps struct {
	Films []Films `json:"films"`
}

type Films struct {
	Fields FieldsFilms `json:"fields"`
}

type FieldsFilms struct {
	Title     string     `json:"title"`
	Runtime   int        `json:"runtime"`
	Sessions  []Sessions `json:"sessions"`
	HeroImage HeroImage  `json:"heroImage"`
	Slug      string     `json:"slug"`
}

type Sessions struct {
	Fields FieldsSessions `json:"fields"`
}

type FieldsSessions struct {
	StartTime time.Time `json:"startTime"`
	Cinema    Cinema    `json:"cinema"`
}

type Cinema struct {
	Fields FieldsCinema `json:"fields"`
}

type FieldsCinema struct {
	Name string `json:"name"`
}

type HeroImage struct {
	Fields FieldsHeroImage `json:"fields"`
}

type FieldsHeroImage struct {
	Image ImageFields `json:"image"`
}

type ImageFields struct {
	FieldsImage FieldsImage `json:"fields"`
}

type FieldsImage struct {
	File File `json:"file"`
}

type File struct {
	URL string `json:"url"`
}
