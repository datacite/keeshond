package main

import (
	"log"

	"github.com/datacite/keeshond/internal/app"
	"github.com/datacite/keeshond/internal/app/db"
	"github.com/datacite/keeshond/internal/app/net"
)

func main() {
	// Get configuration from environment variables.
	var config = app.GetConfigFromEnv()

	// Run with configuration.
	if err := run(config); err != nil {
		log.Fatal(err)
	}
}

func run(config *app.Config) error {
	// Setup connection to database.
	dsn := db.CreateClickhouseDSN(
		config.AnalyticsDatabase.Host,
		config.AnalyticsDatabase.Port,
		config.AnalyticsDatabase.User,
		config.AnalyticsDatabase.Password,
		config.AnalyticsDatabase.Dbname,
	)
	conn, err := db.NewGormClickhouseConnection(dsn)

	if err != nil {
		// Log error and exit.
		log.Fatal(err)
	}

	// Test database connection.
	if err := db.TestConnection(conn); err != nil {
		// Log error and exit.
		log.Fatal(err)
	} else {
		log.Println("Database connection successful.")
	}

	// Migrations.
	if err := db.AutoMigrate(conn); err != nil {
		log.Println(err)
	}

	server := net.NewHttpServer(config, conn)

	// Open the server.
	if err := server.Open(); err != nil {
		return err
	}

	return nil
}
