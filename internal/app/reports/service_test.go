package reports

import (
	"testing"
	"time"

	"github.com/datacite/keeshond/internal/app/stats"
)

type MockReportsRepositoryReader struct {

}

type MockStatsService struct {

}

// Mock breakdown
func (m *MockStatsService) BreakdownByPID(repoId string, query stats.Query, page int, pageSize int) []stats.BreakdownResult {
	// Fake paging
	if page == 1 {
		return []stats.BreakdownResult{
			{
				Pid: "10.1234/1",
				TotalViews: 100,
				UniqueViews: 50,
				TotalDownloads: 50,
				UniqueDownloads: 25,
			},
			{
				Pid: "10.1234/2",
				TotalViews: 100,
				UniqueViews: 50,
				TotalDownloads: 50,
				UniqueDownloads: 25,
			},
		}
	} else if page == 2 {
		return []stats.BreakdownResult{
			{
				Pid: "10.1234/3",
				TotalViews: 100,
				UniqueViews: 50,
				TotalDownloads: 50,
				UniqueDownloads: 25,
			},
			{
				Pid: "10.1234/4",
				TotalViews: 100,
				UniqueViews: 50,
				TotalDownloads: 50,
				UniqueDownloads: 25,
			},
		}
	} else {
		return []stats.BreakdownResult{}
	}
}

// Mock count unique
func (m *MockStatsService) CountUniquePID(repoId string, query stats.Query) int64 {
	return 4
}

// Mock timeseries
func (m *MockStatsService) Timeseries(repoId string, query stats.Query) []stats.TimeseriesResult {
	return []stats.TimeseriesResult{
		{
			Date: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			TotalViews: 100,
			UniqueViews: 50,
			TotalDownloads: 50,
			UniqueDownloads: 25,
		},
		{
			Date: time.Date(2018, 1, 2, 0, 0, 0, 0, time.UTC),
			TotalViews: 200,
			UniqueViews: 100,
			TotalDownloads: 100,
			UniqueDownloads: 50,
		},
	}
}

// Mock aggregate
func (m *MockStatsService) Aggregate(repoId string, query stats.Query) stats.AggregateResult {
	return stats.AggregateResult{
		TotalViews: 100,
		UniqueViews: 50,
		TotalDownloads: 50,
		UniqueDownloads: 25,
	}
}

// Test that the service can generate a dataset usage report
func TestGenerateDatasetUsageReport(t *testing.T) {
	// Create a mock stats service
	mockStatsService := &MockStatsService{}

	// Create a service
	service := NewReportsService(mockStatsService)

	beginDate := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2018, 12, 31, 0, 0, 0, 0, time.UTC)

	// Create fake shared data
	sharedData := SharedData {
		Platform: "datacite",
		Publisher: "datacite",
		PublisherId: "datacite.test",
	}

	// Generate the report, returns a function that can be called to get the report
	generateReport, err := service.GenerateDatasetUsageReport("datacite", beginDate, endDate, sharedData, false)

	if err != nil {
		t.Error(err)
	}

	report, _ := generateReport()

	// Check that the report header is correct
	if report.ReportHeader.ReportId != "dsr" {
		t.Errorf("ReportId is not correct got %s", report.ReportHeader.ReportId)
	}

	if report.ReportHeader.ReportName != "Dataset Master Report" {
		t.Errorf("ReportName is not correct got %s", report.ReportHeader.ReportName)
	}

	if report.ReportHeader.Release != "rd1" {
		t.Errorf("Release is not correct got %s", report.ReportHeader.Release)
	}

	if report.ReportHeader.Created != beginDate.Format(time.RFC3339) {
		t.Errorf("Created is not correct got %s", report.ReportHeader.Created)
	}

	if report.ReportHeader.ReportingPeriod.BeginDate.Format("2006-01-02") != "2018-01-01" {
		t.Errorf("BeginDate is not correct got %s", report.ReportHeader.ReportingPeriod.BeginDate.Format("2006-01-02"))
	}

	if report.ReportHeader.ReportingPeriod.EndDate.Format("2006-01-02") != "2018-12-31" {
		t.Errorf("EndDate is not correct got %s", report.ReportHeader.ReportingPeriod.EndDate.Format("2006-01-02"))
	}

	// Check that the report metrics are correct

	datasets := report.ReportDatasets

	if len(datasets) != 4 {
		t.Errorf("ReportDatasets length is not correct got %d", len(report.ReportDatasets))
	}

	// check first dataset
	first_dataset := datasets[0]

	if first_dataset.DatasetId.Value != "10.1234/1" {
		t.Errorf("DatasetId is not correct got %s", first_dataset.DatasetId.Value)
	}

	if first_dataset.DatasetTitle != "" {
		t.Errorf("DatasetTitle is not correct got %s", first_dataset.DatasetTitle)
	}

	if first_dataset.Publisher != "datacite" {
		t.Errorf("Publisher is not correct got %s", first_dataset.Publisher)
	}

	if first_dataset.PublisherId.Value != "datacite.test" {
		t.Errorf("PublisherId is not correct got %s", first_dataset.PublisherId.Value)
	}

	// Check performance metrics
	if first_dataset.Performance[0].Instance[0].Count != 50 {
		t.Errorf("Total Dataset Requests is not correct got %d", first_dataset.Performance[0].Instance[0].Count)
	}

	if first_dataset.Performance[0].Instance[1].Count != 25 {
		t.Errorf("Unique Dataset Requests is not correct got %d", first_dataset.Performance[0].Instance[1].Count)
	}

	if first_dataset.Performance[0].Instance[2].Count != 100 {
		t.Errorf("Total Dataset Investigations is not correct got %d", first_dataset.Performance[0].Instance[2].Count)
	}

	if first_dataset.Performance[0].Instance[3].Count != 50 {
		t.Errorf("Total Dataset Investigations is not correct got %d", first_dataset.Performance[0].Instance[3].Count)
	}

}
