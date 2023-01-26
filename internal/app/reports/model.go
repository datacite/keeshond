package reports

import "time"

type ReportingPeriod struct {
	BeginDate time.Time `json:"begin-date"`
	EndDate   time.Time `json:"end-date"`
}

type Exception struct {
	Code string `json:"code"`
	Severity string `json:"severity"`
	Message string `json:"message"`
	HelpUrl string `json:"help-url"`
	Data string `json:"data"`
}

type CounterIdentifier struct {
	Type string `json:"type"`
	Value string `json:"value"`
}

type CounterDatasetInstance struct {
	MetricType string `json:"metric-type"`
	Count int `json:"count"`
	AccessMethod string `json:"access-method"`
}

type CounterDatasetPerformance struct {
	Instance CounterDatasetInstance `json:"instance"`
	Period ReportingPeriod `json:"period"`
}

// SUSHI report header struct
type ReportHeader struct {
	ReportName string `json:"report-name"`
	ReportId   string `json:"report-id"`
	Release string `json:"release"`
	Created   string `json:"created"`
	CreatedBy string `json:"created-by"`
	ReportingPeriod ReportingPeriod `json:"reporting-period"`
	ReportFilters []string `json:"report-filters"`
	ReportAttributes []string `json:"report-attributes"`
	Exceptions []Exception `json:"exceptions"`
}

// COUNTER report dataset usage struct
type CounterDatasetUsage struct {
	DatasetTitle string `json:"dataset-title"`
	DatasetId CounterIdentifier `json:"dataset-id"`
	Platform string `json:"platform"`
	Publisher string `json:"publisher"`
	PublisherId CounterIdentifier `json:"publisher-id"`
	DataType string `json:"data-type"`
	Performance []CounterDatasetPerformance `json:"performance"`
}

type CounterDatasetReport struct {
	ReportHeader ReportHeader `json:"report-header"`
	ReportDatasets []CounterDatasetUsage `json:"report-datasets"`
}