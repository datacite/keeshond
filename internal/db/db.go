package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Format a postgresql dsn from seperate config fields
func CreatePostgresDSN(host, port, user, dbname, password, sslmode string) string {
    	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)
}

func NewGormPostgresConnection(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
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