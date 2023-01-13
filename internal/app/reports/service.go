package reports

import (
	"errors"
	"fmt"
	"time"

	"github.com/datacite/keeshond/internal/app/stats"
)

type ReportsService struct {
	statsService stats.StatsServiceInterface
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
func (service *ReportsService) GenerateDatasetUsageReport(repoId string, startDate time.Time, endDate time.Time, addCompressedHeader bool) (func() (*CounterDatasetReport, error), error) {
	// Create stats query object
	query := stats.Query {
		Start: startDate,
		End: endDate,
	}

	// print query
	fmt.Println(query)

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
				datasetUsage := generateDatasetUsage(startDate, endDate, result)

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
		if addCompressedHeader {
			// Add exception that this will be compressed report
			exceptions = []Exception{
				Exception{
					Code: "69",
					Message: "Report is compressed using gzip",
					Severity: "warning",
					HelpUrl: "https://github.com/datacite/sashimi",
					Data: "usage data needs to be uncompressed",
				},
			}
		}

		// Generate report header
		reportHeader := generateReportHeader(startDate, endDate, exceptions)

		// Generate report
		report := CounterDatasetReport{
			ReportHeader: reportHeader,
			ReportDatasets: results,
		}

		return &report, nil
	}

	return generateReportFunc, nil
}

func NewError(s string) {
	panic("unimplemented")
}

// Generate report header
func generateReportHeader(beginDate time.Time, endDate time.Time, exceptions []Exception) ReportHeader {
	var reportHeader ReportHeader

	reportHeader.ReportName = "Dataset Master Report"
	reportHeader.Release = "rd1"
	reportHeader.ReportId = "dsr"
	reportHeader.Created = beginDate.Format(time.RFC3339)
	reportHeader.CreatedBy = "DataCite"
	reportHeader.ReportingPeriod = ReportingPeriod{
		BeginDate: beginDate,
		EndDate: endDate,
	}

	var missingAttributeExceptions []Exception
	// Add missing attribute exceptions for default dataset attributes we know will be missing
	missingAttributeExceptions = append(missingAttributeExceptions, Exception{
		Code: "3071",
		Severity: "warning",
		Message: "dataset-title",
		Data: "dataset-title is unavailable in this report, can be obtained from metadata lookup based on dataset-id",
	}, Exception{
		Code: "3071",
		Severity: "warning",
		Message: "platform",
	}, Exception{
		Code: "3071",
		Severity: "warning",
		Message: "publisher",
		Data: "publisher is unavailable in this report, can be obtained from metadata lookup based on dataset-id",
	}, Exception{
		Code: "3071",
		Severity: "warning",
		Message: "publisher-id",
		Data: "publisher-id is unavailable in this report, can be obtained from metadata lookup based on dataset-id",
	})

	// Combine exceptions
	reportHeader.Exceptions = append(exceptions, missingAttributeExceptions...)

	return reportHeader
}

func generateDatasetUsage(beginDate time.Time, endDate time.Time, result stats.BreakdownResult) CounterDatasetUsage {
	var datasetUsage CounterDatasetUsage

	datasetUsage.DatasetTitle = ""

	datasetUsage.DatasetId = CounterIdentifier{
		Type: "DOI",
		Value: result.Pid,
	}

	datasetUsage.Platform = ""
	datasetUsage.Publisher = ""
	datasetUsage.PublisherId = CounterIdentifier{}

	datasetUsage.DataType = "dataset"
	datasetUsage.Performance = []CounterDatasetPerformance{
		{
			Instance: CounterDatasetInstance{
				MetricType: "total-dataset-requests",
				Count: int(result.TotalDownloads),
			},
			Period: ReportingPeriod{
				BeginDate: beginDate,
				EndDate: endDate,
			},
		},
		{
			Instance: CounterDatasetInstance{
				MetricType: "unique-dataset-requests",
				Count: int(result.UniqueDownloads),
			},
			Period: ReportingPeriod{
				BeginDate: beginDate,
				EndDate: endDate,
			},
		},
		{
			Instance: CounterDatasetInstance{
				MetricType: "total-dataset-invesigations",
				Count: int(result.TotalViews),
			},
			Period: ReportingPeriod{
				BeginDate: beginDate,
				EndDate: endDate,
			},
		},
		{
			Instance: CounterDatasetInstance{
				MetricType: "unique-dataset-investigations",
				Count: int(result.UniqueViews),
			},
			Period: ReportingPeriod{
				BeginDate: beginDate,
				EndDate: endDate,
			},
		},
	}

	return datasetUsage
}