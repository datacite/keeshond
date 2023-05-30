package event

import (
	"testing"
)

// Test checkDoiExistsInDataCite
func TestValidateDoiUrl(t *testing.T) {
	if(!validateDoiUrl("http://www.example.com/url?foo=bar&foo=baz#this_is_fragment", "https://www.example.com/url?foo=bar&foo=baz#this_is_fragment")) {
		t.Errorf("validateDoiUrl should return false")
	}
}