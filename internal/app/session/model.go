package session

import (
	"time"
)

type Salt struct {
	ID      uint `gorm:"primary key;autoIncrement"`
	Salt    []byte
	Created time.Time
}
