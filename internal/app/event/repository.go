package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/datacite/keeshond/internal/app"
	"gorm.io/gorm"
)

type EventRepositoryReader interface {
	Create(event *Event) error
	// GetAll() ([]Event, error)
	// GetByID(id uint) (Event, error)
}

//
// Database implementation of the event repository
//

type EventRepository struct {
	db 		*gorm.DB
	config 	*app.Config
}

// NewRepository creates a new event repository
func NewEventRepository(db *gorm.DB, config *app.Config) *EventRepository {
	return &EventRepository{
		db: db,
		config: config,
	}
}

// Create a new event
func (repository *EventRepository) Create(event *Event) error {
	return repository.db.Create(event).Error
}


//
// Plausible implementation of the event repository
//

type RepositoryPlausible struct {
	config *app.Config
}

func NewRepositoryPlausible(config *app.Config) *RepositoryPlausible {
	return &RepositoryPlausible{
		config: config,
	}
}

// Create a new event
func (repository *RepositoryPlausible) Create(event *Event) error {

	type PlausibleEvent struct {
		Name   string `json:"name"`
		Domain string `json:"domain"`
		Url    string `json:"url"`
		Props  string `json:"props"`
	}

	// URL to send the metric to
	url := fmt.Sprintf("%s/api/event", repository.config.Plausible.Url)

	// Create plausible event
	plausibleEvent := PlausibleEvent{
		Name:   event.Name,
		Domain: event.RepoId,
		Url:    event.Url,
		Props:  "{\"pid\":\"" + event.Pid + "\"}",
	}

	// Marshal plausible event to json
	jsonData, err := json.Marshal(plausibleEvent)

	// Post json to url
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", event.Useragent)
	req.Header.Set("X-Forwarded-For", event.ClientIp)

	// Http client
	client := &http.Client{}

	// Send request as post
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Close response
	defer resp.Body.Close()

	return nil
}
