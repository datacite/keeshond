package db

import (
	"fmt"
	"log"
	"os"
	"time"

	extraClausePlugin "github.com/WinterYukky/gorm-extra-clause-plugin"
	"github.com/datacite/keeshond/internal/app/event"
	"github.com/datacite/keeshond/internal/app/session"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Format a clickhouse dsn from seperate config fields
func CreateClickhouseDSN(host, port, user, password, dbname string) string {
    return fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s", user, password, host, port, dbname)
}

func NewGormClickhouseConnection(dsn string) (*gorm.DB, error) {

    // Setup a custom logger for gorm
    newLogger := logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
        logger.Config{
          SlowThreshold:              time.Second,   // Slow SQL threshold
          LogLevel:                   logger.Error, // Log level
          IgnoreRecordNotFoundError: true,           // Ignore ErrRecordNotFound error for logger
          Colorful:                  false,          // Disable color
        },
      )

    // Open the connection with the custom logger
    db, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{
        Logger: newLogger,
    })

    if err != nil {
        return db, err
    }

    // Gives us special access to do a common table expression with gorm i.e. "with"
    db.Use(extraClausePlugin.New())

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
    var err error

    err = db.Set("gorm:table_options", event.TABLE_OPTIONS).AutoMigrate(&event.Event{})

    if err != nil {
        return err
    }

    err = db.AutoMigrate (
        &session.Salt{},
	)
    return err
}