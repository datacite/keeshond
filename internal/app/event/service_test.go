package event

import (
	"fmt"
	"testing"

	"github.com/datacite/keeshond/internal/app"
)

func buildEventService(dataCiteUrl string, validateDoiExistence bool, validateDoiUrl bool) *EventService {
	config := &app.Config{
		DataCite: struct {
			Url          string
			JWT          string
			JWTPublicKey string
		}{
			Url: dataCiteUrl,
		},
		Validate: struct {
			DoiExistence bool
			DoiUrl       bool
		}{
			DoiExistence: validateDoiExistence,
			DoiUrl:       validateDoiUrl,
		},
	}

	eventService := &EventService{
		config: config,
	}

	return eventService
}

func TestValidateDoiUrl(t *testing.T) {
	if !validateDoiUrl("http://www.example.com/url/?foo=bar&foo=baz#this_is_fragment", "https://www.example.com/url?foo=bar&foo=baz#this_is_fragment") {
		t.Errorf("validateDoiUrl should return false")
	}
}

// Test Validate
func TestValidateSuccessWithBothDoiExistenceCheckAndDoiUrlCheck(t *testing.T) {
	eventService := buildEventService("https://api.stage.datacite.org", true, true)

	eventRequest := &EventRequest{
		Name: "view",
		Pid:  "10.70102/mdc.jeopardy",
		Url:  "https://demorepo.stage.datacite.org/datasets/10.70102/mdc.jeopardy",
	}

	err := eventService.Validate(eventRequest)

	if err != nil {
		t.Errorf("Validate should return nil")
	}
}

func TestValidateSuccessWithOnlyDoiExistenceCheck(t *testing.T) {
	eventService := buildEventService("https://api.stage.datacite.org", true, false)

	eventRequest := &EventRequest{
		Name: "view",
		Pid:  "10.70102/mdc.jeopardy",
		// Set Url to a value that should fail.
		// Since we are skipping the Url check, this test should pass i.e. the existence check should be successful.
		Url: "https://demorepo.stage.datacite.org/datasets/10.70102/mdc.jeopardy.no.bueno",
	}

	err := eventService.Validate(eventRequest)

	if err != nil {
		t.Errorf("Validate should return nil")
	}
}

func TestValidateSuccessWithOnlyDoiUrlCheck(t *testing.T) {
	eventService := buildEventService("https://api.stage.datacite.org", false, true)

	eventRequest := &EventRequest{
		Name: "view",
		Pid:  "10.70102/mdc.jeopardy",
		Url:  "https://demorepo.stage.datacite.org/datasets/10.70102/mdc.jeopardy",
	}

	err := eventService.Validate(eventRequest)

	if err != nil {
		t.Errorf("Validate should return nil")
	}
}

func TestValidateSuccessWithNeitherDoiExistenceCheckAndDoiUrlCheck(t *testing.T) {
	eventService := buildEventService("https://api.stage.datacite.org", false, false)

	eventRequest := &EventRequest{
		Name: "view",
		Pid:  "10.70102/mdc.jeopardy",
		Url:  "https://demorepo.stage.datacite.org/datasets/10.70102/mdc.jeopardy",
	}

	err := eventService.Validate(eventRequest)

	if err != nil {
		t.Errorf("Validate should return nil")
	}
}

func TestValidateSuccessWithEventRequestNameNotView(t *testing.T) {
	eventService := buildEventService("https://api.stage.datacite.org", true, true)

	eventRequest := &EventRequest{
		Name: "not_view",
		// Set PID to a value that would not resolve to a DOI.
		Pid: "10.70102/mdc.jeopardy.no.bueno",
		// Set URL to a value that would not match the PID.
		Url: "https://demorepo.stage.datacite.org/datasets/10.70102/mdc.jeopardy.sin coincidencia",
	}

	err := eventService.Validate(eventRequest)

	if err != nil {
		t.Errorf("Validate should return nil")
	}
}

func TestValidateFailureWhenDoiDoesNotExist(t *testing.T) {
	eventService := buildEventService("https://api.stage.datacite.org", true, true)

	eventRequest := &EventRequest{
		Name: "view",
		Pid:  "10.70102/mdc.jeopardy.no.bueno",
		Url:  "https://demorepo.stage.datacite.org/datasets/10.70102/mdc.jeopardy",
	}

	err := eventService.Validate(eventRequest)

	const actualErr = "this DOI doesn't exist in DataCite"

	if err == nil {
		t.Errorf("Validate should return an error")
	}

	if err.Error() != actualErr {
		t.Errorf("Validate should return error: %v", actualErr)
	}
}

func TestValidateFailureWhenDoiUrlCheckIsUnsuccessful(t *testing.T) {
	eventService := buildEventService("https://api.stage.datacite.org", true, true)

	eventRequest := &EventRequest{
		Name: "view",
		Pid:  "10.70102/mdc.jeopardy",
		Url:  "https://demorepo.stage.datacite.org/datasets/10.70102/mdc.jeopardy.no.bueno",
	}

	err := eventService.Validate(eventRequest)

	const actualErr = "this DOI doesn't match this URL"

	if err == nil {
		t.Errorf("Validate should return an error")
	}

	if err.Error() != actualErr {
		t.Errorf("Validate should return error: %v", actualErr)
	}
}

func TestValidateFailureWhenCannotAccessDataCiteApi(t *testing.T) {
	// We provide an incorrect DataCite URL in order to generate a failed response.
	eventService := buildEventService("https://api.stage.datamight.com", true, true)

	eventRequest := &EventRequest{
		Name: "view",
		Pid:  "10.70102/mdc.jeopardy",
		Url:  "https://demorepo.stage.datacite.org/datasets/10.70102/mdc.jeopardy",
	}

	err := eventService.Validate(eventRequest)

	fmt.Printf("err: %v\n", err)

	if err == nil {
		t.Errorf("Validate should return an error")
	}
}
