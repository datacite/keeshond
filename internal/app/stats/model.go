package stats

import "time"

type PidStat struct {
	Metric 		string `json:"metric"`
	Pid    	  	string `json:"pid"`
	Count  	  	int64  `json:"count"`
}

type Query struct {
	Start 		time.Time // Beginning of the query period
	End 		time.Time // End of the query period
	Period 		string // Date range to query over e.g. "day", "7d", "30d", "custom"
	Interval 	string // Interval to break the results into e.g. "date", "month", "hour"
}