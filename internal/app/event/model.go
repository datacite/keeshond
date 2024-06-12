package event

import (
	"time"
)

type Event struct {
	Timestamp time.Time `json:"timestamp"`
	Name      string    `json:"name"`
	RepoId    string    `json:"repoId"`
	UserID    uint64    `json:"userId"`
	SessionID uint64    `json:"sessionId"`
	Url       string    `json:"url"`
	Pid       string    `json:"pid"`

	// The following are excluded from being stored, this is part of preventing
	// user identifable information being available to be leaked.
	// They just exist for initial processing and discarded after.
	ClientIp  string `gorm:"-:all" json:"clientIp"`
	Useragent string `gorm:"-:all" json:"useragent"`
}

const TABLE_OPTIONS = "ENGINE=MergeTree PARTITION BY toYYYYMM(timestamp) ORDER BY (repo_id, toDate(timestamp), user_id) SAMPLE BY user_id"
