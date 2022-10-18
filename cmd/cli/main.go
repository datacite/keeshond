package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/datacite/keeshond/internal/app"
	"github.com/datacite/keeshond/internal/app/db"
	"github.com/datacite/keeshond/internal/app/event"
	"github.com/datacite/keeshond/internal/app/session"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

func main() {
    app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "event",
				Usage: "Add an event to the analytics service",
				Action: func(cCtx *cli.Context) error {
					// Parse event json request from first cli argument
					eventRequest := event.EventRequest{}
					if err := json.Unmarshal([]byte(cCtx.Args().First()), &eventRequest); err != nil {
						return err
					}

					// Get configuration from environment variables.
					var config = app.GetConfigFromEnv()
					config.ValidateDoi = false

					// Setup database connection
					conn := createDB(config)

					// Create repositories and services
					sessionRepository := session.NewSessionRepository(conn, config)
					sessionService := session.NewSessionService(sessionRepository, config)

					eventRepository := event.NewEventRepository(conn, config)
					eventService := event.NewEventService(eventRepository, sessionService, config)

					// Send the request to the service for storing
					eventService.CreateEvent(&eventRequest)

					return nil
				},
			},
		},
    }

    if err := app.Run(os.Args); err != nil {
        log.Fatal(err)
    }

}

// Function to setup database connection
func createDB(config *app.Config) *gorm.DB {

	// Create database connection.
	dsn := db.CreateClickhouseDSN(
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Dbname,
		config.Database.Password,
	)
	conn, err := db.NewGormClickhouseConnection(dsn)

	if err != nil {
		// Log fatal
		log.Fatal(err)
	}

	// Test database connection.
	if err := db.TestConnection(conn); err != nil {
		// Log fatal
		log.Fatal(err)
	} else {
		log.Println("Database connection successful.")
	}

	return conn
}

// Function used to do database migrations
func migrateDB(conn *gorm.DB) {
	// Migrations.
	if err := db.AutoMigrate(conn); err != nil {
		log.Println(err)
	}
}