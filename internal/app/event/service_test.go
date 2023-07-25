package event

import (
	"net/http"
	"testing"
)

func TestValidateDoiUrl(t *testing.T) {
	if(!validateDoiUrl("http://www.example.com/url/?foo=bar&foo=baz#this_is_fragment", "https://www.example.com/url?foo=bar&foo=baz#this_is_fragment")) {
		t.Errorf("validateDoiUrl should return false")
	}
}

// Test checkDoiExistsInDataCite
func TestCheckDoiExistsInDataCite(t *testing.T) {
	client := &http.Client{}

	// Check if the DOI exists in DataCite
	err := checkDoiExistsInDataCite("10.70102/mdc.jeopardy", "https://demorepo.stage.datacite.org/datasets/10.70102/mdc.jeopardy", "https://api.stage.datacite.org", client)


	// Check if the error is nil
	if err != nil {
		t.Errorf("checkDoiExistsInDataCite should return nil")
	}
}