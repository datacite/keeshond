package event

import (
	"time"
)

type Event struct {
    ID        uint      `gorm:"primary key;autoIncrement" json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
	RepoId    string    `json:"repoId"`
	Url       string    `json:"url"`
	Useragent string    `json:"useragent"`
	ClientIp  string    `json:"clientIp"`
	Pid       string    `json:"pid"`
}