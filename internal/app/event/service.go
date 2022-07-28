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
	repository RepositoryReader
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
		repository: repository,
		config:     config,
	}
}

func (service *Service) CreateEvent(eventRequest *Request) (Event, error) {
	event := Event{
		Timestamp: time.Now(),
		Name:      eventRequest.Name,
		RepoId:    eventRequest.RepoId,
		Url:       eventRequest.Url,
		Useragent: eventRequest.Useragent,
		ClientIp:  eventRequest.ClientIp,
		Pid:       eventRequest.Pid,
	}
	err := service.repository.Create(&event)
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

type GetUrlResponse struct {
	Url string `json:"url"`
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