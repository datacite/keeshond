package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/datacite/keeshond/internal/app"
)

type Service struct {
	eventRepository RepositoryReader
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
func NewService(repository RepositoryReader, config *app.Config) *Service {
	return &Service{
		eventRepository: repository,
		config:     config,
	}
}

func (service *Service) CreateEvent(eventRequest *Request) (Event, error) {

	now := time.Now()

	// Build a salted integer user id
	// hash(daily_salt + website_domain + ip_address + user_agent)
	// userId := service.config.Salt + eventRequest.repoId
	// salt, user_agent <> ip_address <> domain <> root_domain)

	// // Construct a session id based timestamp date + hour time slice + user id
	// sessionId := now.Format("2006-01-02") + "|" + now.Format("15") + "|" + userId

	var userId int64 = 0
	var sessionId int64 = 0

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
	err := service.eventRepository.Create(&event)
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
	api_url := fmt.Sprintf("%s/dois/%s/get-url", dataciteApiUrl, doi)

	// Post json to url
	resp, _ := http.Get(api_url)

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