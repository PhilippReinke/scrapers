package screening

import "github.com/PhilippReinke/scrapers/models"

type Repo interface {
	QueryAll() ([]models.Screening, error)
	QueryWithFilter(models.Filter) ([]models.Screening, error)
	Insert(screening models.Screening) error
	FilterOptions() models.FilterOptions
}
