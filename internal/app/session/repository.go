package session

import (
	"github.com/datacite/keeshond/internal/app"
	"gorm.io/gorm"
)

type RepositoryReader interface {
	Create(salt *Salt) error
	Get() (Salt, error)
}

type Repository struct {
	db 		*gorm.DB
	config 	*app.Config
}

func NewRepository(db *gorm.DB, config *app.Config) *Repository {
	return &Repository{
		db: db,
		config: config,
	}
}

func (repository *Repository) Create(salt *Salt) error {
	return repository.db.Create(salt).Error
}

func (repository *Repository) Get() (Salt, error) {
	var salt Salt
	if err := repository.db.First(&salt).Error; err != nil {
		return salt, err
	}
	return salt, nil
}
