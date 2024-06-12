package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/datacite/keeshond/internal/app"
	"github.com/datacite/keeshond/internal/app/session"
)

type EventService struct {
	eventRepository EventRepositoryReader
	sessionService  *session.SessionService
	config          *app.Config
}

type EventRequest struct {
	Name      string `json:"name"`
	RepoId    string `json:"repoId"`
	Url       string `json:"url"`
	Useragent string `json:"useragent"`
	ClientIp  string `json:"clientIp"`
	Pid       string `json:"pid"`
}

// NewEventService creates a new event service
func NewEventService(repository EventRepositoryReader, sessionService *session.SessionService, config *app.Config) *EventService {
	return &EventService{
		eventRepository: repository,
		sessionService:  sessionService,
		config:          config,
	}
}

func (service *EventService) CreateEvent(eventRequest *EventRequest) (Event, error) {
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

func (service *EventService) CreateRaw(event Event) (Event, error) {
	err := service.eventRepository.Create(&event)

	return event, err
}

func (service *EventService) Validate(eventRequest *EventRequest) error {
	var err error

	if eventRequest.Name != "view" {
		return err
	}

	if !service.shouldValidate() {
		return err
	}

	resp, err := getDoi(eventRequest.Pid, service.config.DataCite.Url)

	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}

	defer resp.Body.Close()

	if service.config.Validate.DoiExistence {
		err = checkDoiExistence(resp)
	}

	if err != nil {
		return err
	}

	if service.config.Validate.DoiUrl {
		err = checkDoiUrl(resp, eventRequest.Url)
	}

	return err
}

func (service *EventService) shouldValidate() bool {
	return service.config.Validate.DoiExistence && service.config.Validate.DoiUrl
}

func getDoi(doi string, dataciteApiUrl string) (*http.Response, error) {
	apiUrl := fmt.Sprintf("%s/dois/%s/get-url", dataciteApiUrl, doi)

	return http.Get(apiUrl)
}

func checkDoiExistence(resp *http.Response) error {
	if resp.StatusCode == 404 {
		return errors.New("this DOI doesn't exist in DataCite")
	}

	return nil
}

func checkDoiUrl(resp *http.Response, url string) error {
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("failed to read body: %v", err)
	}

	var result struct {
		Url string `json:"url"`
	}

	err = json.Unmarshal(body, &result)

	if err != nil {
		return errors.New("can not unmarshal JSON")
	}

	if !validateDoiUrl(result.Url, url) {
		return errors.New("this DOI doesn't match this URL")
	}

	return nil
}

func validateDoiUrl(doiUrl string, urlCompare string) bool {
	// Compare the result with the url but ignore the protocol
	return stripScheme(doiUrl) == stripScheme(urlCompare)
}

// Function to strip the scheme from a URL
func stripScheme(urlToStrip string) string {
	parsedUrl, _ := url.ParseRequestURI(urlToStrip)
	parsedUrl.Scheme = ""
	// Strip trailing slash from path
	parsedUrl.Path = strings.TrimSuffix(parsedUrl.Path, "/")

	return parsedUrl.String()
}

func CreateMockEvent(metricName string, repoId string, doi string, userId uint64, timestamp time.Time) Event {
	// Construct fake URL from doi
	url := "http://" + repoId + "/page/" + doi

	// Generate session id
	sessionId := session.GenerateSessionId(userId, timestamp)

	// View event
	return Event{
		RepoId:    repoId,
		Name:      metricName,
		Useragent: "Mozilla/5.0 (compatible; FakeUser/1.0; +http://www.example.com/bot.html)",
		ClientIp:  "127.0.0.1",
		UserID:    userId,
		SessionID: sessionId,
		Url:       url,
		Pid:       doi,
		Timestamp: timestamp,
	}
}
