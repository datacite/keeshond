package reports

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/datacite/keeshond/internal/app/stats"
)

type ReportsService struct {
	statsService stats.StatsServiceInterface
}

type SharedData struct {
	Platform    string `json:"platform"`
	Publisher   string `json:"publisher"`   // name of the repository
	PublisherId string `json:"publisherId"` // always a client-id
}

func NewReportsService(statsService stats.StatsServiceInterface) *ReportsService {
	return &ReportsService{
		statsService: statsService,
	}
}

// GenerateDatasetUsageReport generates a dataset usage report
// It returns a function to generate part or the full report depending on number of results
// A nil pointer is returned when the callback function is called and there are no results
// If there are more than 50,000 results, the report results should be compressed and an exception is added to the report header to signify this.
func (service *ReportsService) GenerateDatasetUsageReport(repoId string, startDate time.Time, endDate time.Time, sharedData SharedData, addCompressedHeader bool) (func() (*CounterDatasetReport, error), error) {
	// Create stats query object
	query := stats.Query{
		Start: startDate,
		End:   endDate,
	}

	// Hardcoded for now, but possibly configurable in the future.
	reportSize := 50000

	page := 1
	pageSize := 1000

	generateReportFunc := func() (*CounterDatasetReport, error) {
		// Loop through all pages of results until we get empty results
		var results []CounterDatasetUsage
		for {
			// Get results
			breakdownResults := service.statsService.BreakdownByPID(repoId, query, page, pageSize)

			// If we have no results, break
			if len(breakdownResults) == 0 {
				break
			}

			// Loop through results and generate report datasets
			for _, result := range breakdownResults {
				// Generate dataset usage
				datasetUsage := generateDatasetUsage(startDate, endDate, result, sharedData)

				// Add to results
				results = append(results, datasetUsage)
			}

			// Increment page
			page++

			// If we have more than the max number of records, break
			if len(results) >= reportSize {
				break
			}
		}

		var exceptions = []Exception{}
		if len(results) == 0 {
			// return error
			return nil, errors.New("No results found for this query")
		}

		// We never know the dataset-title so we explicitly add an exception for this
		// This information would need to come from our API which would require a lookup based on dataset-id
		// This could provide too much of an overhead for report generation.
		exceptions = append(exceptions, Exception{
			Code:     3071,
			Severity: "warning",
			Message:  "dataset-title",
			Data:     "dataset-title is unavailable in this report, can be obtained from metadata lookup based on dataset-id",
		})

		// Add missing attribute exceptions for potentially missing data
		if sharedData.Platform == "" {
			exceptions = append(exceptions, Exception{
				Code:     3071,
				Severity: "warning",
				Message:  "platform",
			})
		}
		if sharedData.Publisher == "" {
			exceptions = append(exceptions, Exception{
				Code:     3071,
				Severity: "warning",
				Message:  "publisher",
				Data:     "publisher is unavailable in this report, can be obtained from metadata lookup based on dataset-id",
			})
		}
		if sharedData.PublisherId == "" {
			exceptions = append(exceptions, Exception{
				Code:     3071,
				Severity: "warning",
				Message:  "publisher-id",
				Data:     "publisher-id is unavailable in this report, can be obtained from metadata lookup based on dataset-id",
			})
		}

		if addCompressedHeader {
			// Add exception that this will be compressed report
			exceptions = append(exceptions, Exception{
				Code:     69,
				Message:  "Report is compressed using gzip",
				Severity: "warning",
				HelpUrl:  "https://github.com/datacite/sashimi",
				Data:     "usage data needs to be uncompressed",
			})
		}

		// Generate report header
		reportHeader := generateReportHeader(startDate, endDate, sharedData, exceptions)

		// Generate report
		report := CounterDatasetReport{
			ReportHeader:   reportHeader,
			ReportDatasets: results,
		}

		return &report, nil
	}

	return generateReportFunc, nil
}

// Generate report header
func generateReportHeader(beginDate time.Time, endDate time.Time, sharedData SharedData, exceptions []Exception) ReportHeader {
	var reportHeader ReportHeader

	reportHeader.ReportName = "Dataset Master Report"
	reportHeader.Release = "rd1"
	reportHeader.ReportId = "dsr"
	reportHeader.Created = beginDate.Format(time.RFC3339)
	if sharedData.PublisherId != "" {
		reportHeader.CreatedBy = "da_" + sharedData.PublisherId
	} else {
		reportHeader.CreatedBy = "datacite-analytics"
	}
	reportHeader.ReportingPeriod = ReportingPeriod{
		BeginDate: beginDate,
		EndDate:   endDate,
	}
	reportHeader.ReportFilters = []string{}
	reportHeader.ReportAttributes = []string{}

	// Combine exceptions
	reportHeader.Exceptions = exceptions

	return reportHeader
}

func generateDatasetUsage(beginDate time.Time, endDate time.Time, result stats.BreakdownResult, sharedData SharedData) CounterDatasetUsage {
	var datasetUsage CounterDatasetUsage

	datasetUsage.DatasetTitle = ""

	datasetUsage.DatasetId = []CounterIdentifier{{
		Type:  "DOI",
		Value: result.Pid,
	}}

	datasetUsage.Platform = sharedData.Platform
	datasetUsage.Publisher = sharedData.Publisher

	if sharedData.PublisherId != "" {
		datasetUsage.PublisherId = []CounterIdentifier{{
			Type:  "client-id",
			Value: sharedData.PublisherId,
		}}
	} else {
		datasetUsage.PublisherId = []CounterIdentifier{}
	}

	datasetUsage.DataType = "dataset"
	datasetUsage.Performance = []CounterDatasetPerformance{
		{
			Period: ReportingPeriod{
				BeginDate: beginDate,
				EndDate:   endDate,
			},
			Instance: []CounterDatasetInstance{
				{
					MetricType:   "total-dataset-requests",
					Count:        int(result.TotalDownloads),
					AccessMethod: "regular",
				},
				{
					MetricType:   "unique-dataset-requests",
					Count:        int(result.UniqueDownloads),
					AccessMethod: "regular",
				},
				{
					MetricType:   "total-dataset-investigations",
					Count:        int(result.TotalViews),
					AccessMethod: "regular",
				},
				{
					MetricType:   "unique-dataset-investigations",
					Count:        int(result.UniqueViews),
					AccessMethod: "regular",
				},
			},
		},
	}

	return datasetUsage
}

func SendReportToAPI(reportsAPIEndpoint string, compressedJson []byte, jwt string) error {
	// Make a POST request to Reports API
	bodyReader := bytes.NewReader(compressedJson)
	req, err := http.NewRequest(http.MethodPost, reportsAPIEndpoint, bodyReader)
	if err != nil {
		return err
	}

	// Set content type to be gzip as report will be compressed
	req.Header.Set("Content-Type", "application/gzip")
	// Set content encoding to gzip
	req.Header.Set("Content-Encoding", "gzip")

	// Add JWT token to request
	req.Header.Set("Authorization", "Bearer "+jwt)

	client := http.Client{
		Timeout: 100 * time.Second,
		Transport: &http.Transport{
			DisableCompression: true,
		},
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
