package stats

import "time"

type AggregateResult struct {
	TotalViews int64 `json:"total_views"`
	UniqueViews int64 `json:"unique_views"`
	TotalDownloads int64 `json:"total_downloads"`
	UniqueDownloads int64 `json:"unique_downloads"`
}

type TimeseriesResult struct {
	Date time.Time `json:"date"`
	TotalViews int64 `json:"total_views"`
	UniqueViews int64 `json:"unique_views"`
	TotalDownloads int64 `json:"total_downloads"`
	UniqueDownloads int64 `json:"unique_downloads"`
}

type BreakdownResult struct {
	Pid string `json:"pid"`
	TotalViews int64 `json:"total_views"`
	UniqueViews int64 `json:"unique_views"`
	TotalDownloads int64 `json:"total_downloads"`
	UniqueDownloads int64 `json:"unique_downloads"`
}

type PidStat struct {
	Metric 		string `json:"metric"`
	Pid    	  	string `json:"pid"`
	Count  	  	int64  `json:"count"`
}

type Query struct {
	Start 		time.Time // Beginning of the query period
	End 		time.Time // End of the query period
	Period 		string // Date range to query over e.g. "day", "7d", "30d", "custom"
	Interval 	string // Interval to break the results into e.g. "day", "month", "hour"
}