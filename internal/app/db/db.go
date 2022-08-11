package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/datacite/keeshond/internal/app/event"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Format a postgresql dsn from seperate config fields
func CreatePostgresDSN(host, port, user, dbname, password, sslmode string) string {
    	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

func NewGormPostgresConnection(dsn string) (*gorm.DB, error) {
    newLogger := logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
        logger.Config{
          SlowThreshold:              time.Second,   // Slow SQL threshold
          LogLevel:                   logger.Error, // Log level
          IgnoreRecordNotFoundError: true,           // Ignore ErrRecordNotFound error for logger
          Colorful:                  false,          // Disable color
        },
      )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: newLogger,
    })

    if err != nil {
        return db, err
    }
    return db, nil
}

// Test if the database connection is working
func TestConnection(db *gorm.DB) error {
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }
    err = sqlDB.Ping()
    if err != nil {
        return err
    }
    return nil
}

// Migrate models
func AutoMigrate(db *gorm.DB) error{
    err := db.AutoMigrate (
        &event.Event{},
	)
    return err
}
