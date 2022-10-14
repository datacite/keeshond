package main

import (
	"log"

	"github.com/datacite/keeshond/internal/app"
	"github.com/datacite/keeshond/internal/app/db"
	"github.com/datacite/keeshond/internal/app/event"
	"github.com/datacite/keeshond/internal/app/session"
)

func main() {
	// Get configuration from environment variables.
	var config = app.GetConfigFromEnv()

	config.ValidateDoi = false

	// Run with configuration.
	if err := run(config); err != nil {
		log.Fatal(err)
	}
}

func run(config *app.Config) error {
	// Create database connection.
	conn, err := db.NewGormPostgresConnection(db.CreatePostgresDSN(
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Dbname,
		config.Database.Password,
		config.Database.Sslmode,
	))

	if err != nil {
		return err
	}

	// Test database connection.
	if err := db.TestConnection(conn); err != nil {
		return err
	} else {
		log.Println("Database connection successful.")
	}

	// Migrations.
	if err := db.AutoMigrate(conn); err != nil {
		log.Println(err)
	}


	// Register repositories and services
	sessionRepository := session.NewRepository(conn, config)
	sessionService := session.NewService(sessionRepository, config)

	eventRepository := event.NewRepositoryDB(conn, config)
	eventService := event.NewService(eventRepository, sessionService, config)

	// Build test request event
	eventRequest := event.Request{
		Name:      "Test",
		RepoId:    "example.com",
		Url:       "http://example.com/page/10.5072/12345",
		Useragent: "Mozilla/5.0 (compatible; FakeUser/1.0; +http://www.example.com/bot.html)",
		ClientIp:  "127.0.0.1",
		Pid:       "10.5072/12345",
	}

	eventService.CreateEvent(&eventRequest)

	return nil
}
