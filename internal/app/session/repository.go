package session

import (
	"github.com/datacite/keeshond/internal/app"
	"gorm.io/gorm"
)

type SessionRepositoryReader interface {
	Create(salt *Salt) error
	Get() (Salt, error)
}

type SessionRepository struct {
	db     *gorm.DB
	config *app.Config
}

func NewSessionRepository(db *gorm.DB, config *app.Config) *SessionRepository {
	return &SessionRepository{
		db:     db,
		config: config,
	}
}

func (repository *SessionRepository) Create(salt *Salt) error {
	return repository.db.Create(salt).Error
}

func (repository *SessionRepository) Get() (Salt, error) {
	var salt Salt
	if err := repository.db.First(&salt).Error; err != nil {
		return salt, err
	}
	return salt, nil
}
