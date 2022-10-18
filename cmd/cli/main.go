package main

import (
	"log"

	"github.com/datacite/keeshond/internal/app"
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
	log.Println("Not implemented")

	// // Create database connection.
	// dsn := db.CreateClickhouseDSN(
	// 	config.Database.Host,
	// 	config.Database.Port,
	// 	config.Database.User,
	// 	config.Database.Dbname,
	// 	config.Database.Password,
	// )
	// conn, err := db.NewGormClickhouseConnection(dsn)

	// if err != nil {
	// 	return err
	// }

	// // Test database connection.
	// if err := db.TestConnection(conn); err != nil {
	// 	return err
	// } else {
	// 	log.Println("Database connection successful.")
	// }

	// // Migrations.
	// if err := db.AutoMigrate(conn); err != nil {
	// 	log.Println(err)
	// }


	// Register repositories and services
	// sessionRepository := session.NewSessionRepository(conn, config)
	// sessionService := session.NewSessionService(sessionRepository, config)

	// eventRepository := event.NewEventRepository(conn, config)
	// eventService := event.NewEventService(eventRepository, sessionService, config)

	// statsRepository := stats.NewStatsRepository(conn)
	// statsService := stats.NewStatsService(statsRepository)

	// Build test request event
	// eventRequest := event.EventRequest{
	// 	Name:      "Test",
	// 	RepoId:    "example.com",
	// 	Url:       "http://example.com/page/10.5072/12345",
	// 	Useragent: "Mozilla/5.0 (compatible; FakeUser/1.0; +http://www.example.com/bot.html)",
	// 	ClientIp:  "127.0.0.1",
	// 	Pid:       "10.5072/54321",
	// }

	// eventService.CreateEvent(&eventRequest)

	// totalToday := statsService.GetTotalInToday("Test", "example.com", "10.5072/12345")
	// log.Println(totalToday)

	// totalLast7Days := statsService.GetTotalInLast7Days("Test", "example.com", "10.5072/12345")
	// log.Println(totalLast7Days)

	// totalLast30Days := statsService.GetTotalInLast30Days("Test", "example.com", "10.5072/12345")
	// log.Println(totalLast30Days)

	// totalsByPID := statsService.GetTotalsByPidInLast30Days("Test", "example.com")
	// log.Println(totalsByPID)

	// uniqueToday := statsService.GetUniqueInToday("Test", "example.com", "10.5072/12345")
	// log.Println(uniqueToday)

	// uniqueLast7Days := statsService.GetUniqueInLast7Days("Test", "example.com", "10.5072/12345")
	// log.Println(uniqueLast7Days)

	// uniqueLast30Days := statsService.GetUniqueInLast30Days("Test", "example.com", "10.5072/12345")
	// log.Println(uniqueLast30Days)

	// uniquesByPID := statsService.GetUniquesByPidInLast30Days("Test", "example.com")
	// log.Println(uniquesByPID)

	return nil
}
