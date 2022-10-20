package stats

type PidStat struct {
	Metric 		string `json:"metric"`
	Pid    	  	string `json:"pid"`
	Count  	  	int64  `json:"count"`
}