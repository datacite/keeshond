package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/datacite/keeshond/internal/app"
	"github.com/datacite/keeshond/internal/app/db"
	"github.com/datacite/keeshond/internal/app/event"
	"github.com/datacite/keeshond/internal/app/reports"
	"github.com/datacite/keeshond/internal/app/session"
	"github.com/datacite/keeshond/internal/app/stats"
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
			{
				Name:  "report",
				Usage: "Generate a report",
				Action: func(cCtx *cli.Context) error {
					// go run cmd/cli/main.go report example.com 2022-01-01 2022-12-31

					// Parse repoId from first cli argument
					repoId := cCtx.Args().First()

					// Parse beginDate from third cli argument
					beginDate, err := time.Parse("2006-01-02", cCtx.Args().Get(1))
					if err != nil {
						return err
					}
					// Parse endDate from fourth cli argument
					endDate, err := time.Parse("2006-01-02", cCtx.Args().Get(2))
					if err != nil {
						return err
					}

					// Get configuration from environment variables.
					var config = app.GetConfigFromEnv()
					config.ValidateDoi = false

					// Setup database connection
					conn := createDB(config)

					statsRepository := stats.NewStatsRepository(conn)
					statsService := stats.NewStatsService(statsRepository)

					reportsService := reports.NewReportsService(statsService)

					// Generate report
					generateReport, err := reportsService.GenerateDatasetUsageReport(repoId, beginDate, endDate)

					if err != nil {
						return err
					}

					// Keep calling the generateReport function until it returns nil
					for {
						report, err := generateReport()

						if err != nil {
							break
						}
						if report == nil {
							break
						}

						// Serialize report to json
						reportJson, err := json.MarshalIndent(report, "", "  ")
						if err != nil {
							return err
						}

						filename := fmt.Sprintf("%s-%s-%s.json", repoId, beginDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
						f, err := os.Create(filename)
						if err != nil {
							return err
						}

						_, err = f.Write(reportJson)

						if err != nil {
							return err
						}
					}

					return err
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
		config.Database.Password,
		config.Database.Dbname,
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