package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/datacite/keeshond/internal/app"
	"github.com/datacite/keeshond/internal/app/session"
)

type Service struct {
	eventRepository RepositoryReader
	sessionService *session.Service
	config     *app.Config
}

type Request struct {
	Name      string    `json:"name"`
	RepoId    string    `json:"repoId"`
	Url       string    `json:"url"`
	Useragent string    `json:"useragent"`
	ClientIp  string    `json:"clientIp"`
	Pid       string    `json:"pid"`
}

// NewService creates a new event service
func NewService(repository RepositoryReader, sessionService *session.Service, config *app.Config) *Service {
	return &Service{
		eventRepository: repository,
		sessionService:  sessionService,
		config:     config,
	}
}

func (service *Service) CreateEvent(eventRequest *Request) (Event, error) {
	var err error

	now := time.Now()

	salt, err := service.sessionService.GetSalt()
	if err != nil {
		return Event{}, err
	}

	// Get hostname from the url
	url, err := url.Parse(eventRequest.Url)
    if err != nil {
        return Event{}, err
    }
    hostDomain := strings.TrimPrefix(url.Hostname(), "www.")

	// User id is generate conforming to COUNTER rules
	// It's a cryptographic hash of details with a daily salt.
	var userId uint64 = session.GenerateUserId(
		&salt,
		eventRequest.ClientIp,
		eventRequest.Useragent,
		eventRequest.RepoId,
		hostDomain,
	)

	// Session id is hashed session based on the user id and current time
	// Sessions will be different every hour
	var sessionId uint64 = session.GenerateSessionId(
		userId,
		now,
	)

	event := Event{
		Timestamp: now,
		Name:      eventRequest.Name,
		RepoId:    eventRequest.RepoId,
		UserID:    userId,
		SessionID: sessionId,
		Url:       eventRequest.Url,
		Useragent: eventRequest.Useragent,
		ClientIp:  eventRequest.ClientIp,
		Pid:       eventRequest.Pid,
	}
	err = service.eventRepository.Create(&event)
	return event, err
}

func (service *Service) Validate(eventRequest *Request) error {
	// Http client
	client := &http.Client{}

	// Validate PID when server is set to validate and is a view event
	if service.config.ValidateDoi && eventRequest.Name == "view" {
		return checkDoiExistsInDataCite(eventRequest.Pid, eventRequest.Url, service.config.DataCite.Url, client);
	}

	return nil
}

func checkDoiExistsInDataCite(doi string, url string, dataciteApiUrl string, client *http.Client) error {
	// Make API call to DataCite for DOI

	// URL to send the metric to
	apiUrl := fmt.Sprintf("%s/dois/%s/get-url", dataciteApiUrl, doi)

	// Post json to url
	resp, _ := http.Get(apiUrl)

	if resp.StatusCode == 404 {
		return errors.New("This DOI doesn't exist in DataCite")
	}

	// Close response
	defer resp.Body.Close()

	// Get Json result
	body, _ := ioutil.ReadAll(resp.Body)

	type GetUrlResponse struct {
		Url string `json:"url"`
	}

	var result GetUrlResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return errors.New("Can not unmarshal JSON")
	}

	// Compare the result with the url
	if result.Url != url {
		return errors.New("This DOI doesn't match this URL")
	}

	return nil
}