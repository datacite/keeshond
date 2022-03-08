package keeshond

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Metric represents a single metric
type MetricEvent struct {
	Name      string    `json:"name"`
	RepoId    string    `json:"repoId"`
	Timestamp time.Time `json:"timestamp"`
	Url       string    `json:"url"`
	Useragent string    `json:"useragent"`
	ClientIp  string    `json:"clientIp"`
	Pid       string    `json:"pid"`
}

func NewMetricEvent(name string, repoId string, timestamp time.Time, url string, useragent string, clientIp string, pid string) *MetricEvent {
	return &MetricEvent{
		Name:      name,
		RepoId:    repoId,
		Timestamp: timestamp,
		Url:       url,
		Useragent: useragent,
		ClientIp:  clientIp,
		Pid:       pid,
	}
}

type PlausibleEventCustomProps struct {
	Pid string `json:"pid"`
}

type PlausibleEvent struct {
	Name   string                    `json:"name"`
	Domain string                    `json:"domain"`
	Url    string                    `json:"url"`
	Props  PlausibleEventCustomProps `json:"props"`
}

func SendMetricEventToPlausible(metric *MetricEvent, plausibleUrl string, client *http.Client) error {

	// URL to send the metric to
	url := fmt.Sprintf("%s/api/event", plausibleUrl)

	// Create plausible event
	plausibleEvent := PlausibleEvent{
		Name:   metric.Name,
		Domain: metric.RepoId,
		Url:    metric.Url,
		Props: PlausibleEventCustomProps{
			Pid: metric.Pid,
		},
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
	req.Header.Set("User-Agent", metric.Useragent)
	req.Header.Set("X-Forwarded-For", metric.ClientIp)

	// Send request as post
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Close response
	defer resp.Body.Close()

	return nil
}
