package keeshond

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type GetUrlResponse struct {
	Url string `json:"url"`
}

func checkExistsInDataCite(doi string, url string, dataciteApiUrl string, client *http.Client) error {
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
