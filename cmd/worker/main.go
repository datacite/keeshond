package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/datacite/keeshond/internal/app"
	"github.com/datacite/keeshond/internal/app/db"
	"github.com/datacite/keeshond/internal/app/reports"
	"github.com/datacite/keeshond/internal/app/stats"
	"gorm.io/gorm"
)

func report_job(repoId string, beginDate time.Time, endDate time.Time, platform string, publisher string, publisherId string) error {
	addCompressedHeader := true

	// Get keeshond configuration from environment variables.
	var config = app.GetConfigFromEnv()

	// config contains the datacite URL, construct reports API URL
	reportsAPIURL := config.DataCite.Url + "/reports"

	// Setup database connection
	conn := createDB(config)

	statsRepository := stats.NewStatsRepository(conn)
	statsService := stats.NewStatsService(statsRepository)

	reportsService := reports.NewReportsService(statsService)

	// Create shared data used for all datasets
	sharedData := reports.SharedData {
		Platform: platform,
		Publisher: publisher,
		PublisherId: publisherId,
	}

	// Generate report
	generateReport, err := reportsService.GenerateDatasetUsageReport(repoId, beginDate, endDate, sharedData, addCompressedHeader)

	if err != nil {
		return err
	}

	// Keep calling the generateReport function until it returns nil
	index := 0
	for {
		index++
		report, err := generateReport()

		if err != nil {
			return err
		}
		if report == nil {
			// This is the end of report generation
			break
		}

		// Serialize report to json
		reportJson, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}

		// Gzip json
		compressedJson, _ := gzipData(reportJson)

		// Send to Reports API
		err = sendReportToAPI(reportsAPIURL, compressedJson, config.DataCite.JWT)

		if err != nil {
			return err
		}
	}

	return nil
}

func sendReportToAPI(reportsAPIURL string, compressedJson []byte, jwt string) error {
	// Make a POST request to Reports API

	bodyReader := bytes.NewReader(compressedJson)
	req, err := http.NewRequest(http.MethodPost, reportsAPIURL, bodyReader)
	if err != nil {
		return err
	}
	// Set various headers as required by Reports API
	req.Header.Set("Content-Type", "application/gzip")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept", "gzip")

	// Add JWT token to request
	req.Header.Set("Authorization", "Bearer " + jwt)

	client := http.Client{
		Timeout: 100 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	// Check response code
	switch res.StatusCode {
	case http.StatusCreated:
		log.Default().Println("Report sent to Reports API")
	case http.StatusUnauthorized:
		return errors.New("unauthorized, JWT token is missing or invalid")
	case http.StatusForbidden:
		return errors.New("forbidden, JWT is expired or invalid")
	case http.StatusUnsupportedMediaType:
		return errors.New("did not include correct Content-Type header")
	case http.StatusUnprocessableEntity:
		return errors.New("invalid report provided")
	default:
		return errors.New("Error sending report to Reports API: " + res.Status)
	}

	return nil
}

func main() {
	// Get repoId from environment variable
	repoId, ok := os.LookupEnv("REPO_ID")
	if !ok {
		log.Fatal("REPO_ID environment variable not set")
		return
	}

	// Get beginDate from environment variable
	beginDateStr, ok := os.LookupEnv("BEGIN_DATE")
	if !ok {
		log.Fatal("BEGIN_DATE environment variable not set")
		return
	}

	// Get endDate from environment variable
	endDateStr, ok := os.LookupEnv("END_DATE")
	if !ok {
		log.Fatal("END_DATE environment variable not set")
		return
	}

	// Parse beginDate
	beginDate, err := time.Parse("2006-01-02", beginDateStr)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Parse endDate
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Get platform from environment variable
	platform, ok := os.LookupEnv("PLATFORM")
	if !ok {
		log.Fatal("PLATFORM environment variable not set")
		return
	}

	// Get publisher from environment variable
	publisher, ok := os.LookupEnv("PUBLISHER")
	if !ok {
		log.Fatal("PUBLISHER environment variable not set")
		return
	}

	// Get publisherId from environment variable
	publisherId, ok := os.LookupEnv("PUBLISHER_ID")
	if !ok {
		log.Fatal("PUBLISHER_ID environment variable not set")
		return
	}

	// Output details of report we're generating
	log.Printf("Starting generation of report for repoId: %s, beginDate: %s, endDate: %s, platform: %s, publisher: %s, publisherId: %s", repoId, beginDate, endDate, platform, publisher, publisherId)

	if err := report_job(repoId, beginDate, endDate, platform, publisher, publisherId); err != nil {
		log.Fatal(err)
	}

	// Success message
	log.Println("Report generation completed successfully")
}

// Function to gzip data
func gzipData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Flush(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Function to setup database connection
func createDB(config *app.Config) *gorm.DB {

	// Create database connection.
	dsn := db.CreateClickhouseDSN(
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Dbname,
	)
	conn, err := db.NewGormClickhouseConnection(dsn)

	if err != nil {
		// Log fatal
		log.Fatal(err)
	}

	// Test database connection.
	if err := db.TestConnection(conn); err != nil {
		// Log fatal
		log.Fatal(err)
	}

	return conn
}
