package stats

import (
	"os"
	"testing"
	"time"

	"github.com/datacite/keeshond/internal/app"
	"github.com/datacite/keeshond/internal/app/db"
	"github.com/datacite/keeshond/internal/app/event"
	"github.com/datacite/keeshond/internal/app/session"
	"gorm.io/gorm"
)

type TestState struct {
	conn *gorm.DB
	config *app.Config
}

func setupTestDB(config *app.Config) (*gorm.DB, error) {
	// Get clickhouse dsn
	dsn := db.CreateClickhouseDSN(
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Dbname,
	)

	// Setup db connection
	conn, err := db.NewGormClickhouseConnection(dsn)

	if err != nil {
		return nil, err
	}

	// Test database connection.
	if err := db.TestConnection(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

func setup() TestState{
	// Test config
	config := app.GetConfigFromEnv()
	config.ValidateDoi = false
	config.Database.Dbname = "keeshond_test"

	conn, err := setupTestDB(config)
	if err != nil {
		println(err)
	}

	// Migrations.
	if err := db.AutoMigrate(conn); err != nil {
		println(err)
	}


	state := TestState{
		conn: conn,
		config: config,
	}

	// Setup repository and services
	sessionRepository := session.NewSessionRepository(conn, config)
	sessionService := session.NewSessionService(sessionRepository, config)
	eventRepository := event.NewEventRepository(conn, config)
	eventService := event.NewEventService(eventRepository, sessionService, config)

	// Create dummy event request

	// Create array of fake dois
	dois := []string{"10.1234/1", "10.1234/2", "10.1234/3", "10.1234/4", "10.1234/5", "10.1234/6", "10.1234/7", "10.1234/8", "10.1234/9", "10.1234/10"}

	// Loop over dois and create view events
	for _, doi := range dois {
		// Construct fake URL from doi
		url := "http://example.com/page" + doi

		// View event
		eventRequest := event.EventRequest{
			Name: "view",
			RepoId: "example.com",
			Url: url,
			Useragent: "Mozilla/5.0 (compatible; FakeUser/1.0; +http://www.example.com/bot.html)",
			ClientIp: "127.0.0.1",
			Pid: doi,
		}

		eventService.CreateEvent(&eventRequest)

		// Download event
		eventRequest = event.EventRequest{
			Name: "download",
			RepoId: "example.com",
			Url: url,
			Useragent: "Mozilla/5.0 (compatible; FakeUser/1.0; +http://www.example.com/bot.html)",
			ClientIp: "127.0.0.1",
			Pid: doi,
		}

		eventService.CreateEvent(&eventRequest)
	}

	return state
}

func teardown(state *TestState) {
	// Delete from events
	state.conn.Exec("TRUNCATE TABLE events")

	// Delete salts
	state.conn.Exec("TRUNCATE TABLE salts")
}

func TestMain(m *testing.M) {
	state := setup()
	code := m.Run()
	teardown(&state)
	os.Exit(code)
}

func TestStatsService_Aggregate(t *testing.T) {
	// Test config
	config := app.GetConfigFromEnv()
	config.ValidateDoi = false
	config.Database.Dbname = "keeshond_test"

	conn, err := setupTestDB(config)
	if err != nil {
		// Fail
		t.Errorf("Error connecting to test database: %s", err)
	}

	// Start of today
	start := time.Now().Truncate(24 * time.Hour)

	// End of today
	end := start.Add(24 * time.Hour)

	// Construct query
	query := Query{
		Start: start,
		End: end,
		Period: "day",
		Interval: "hour",
	}

	statsRepository := NewStatsRepository(conn)
	statsService := NewStatsService(statsRepository)

	// Get stats
	result := statsService.Aggregate("example.com", query, []string{"view", "download"})

	if result.TotalDownloads != 10 {
		t.Errorf("TotalDownloads is not 10")
	}

	if result.TotalViews != 10 {
		t.Errorf("TotalViews is not 10")
	}

	if result.UniqueViews != 1 {
		t.Errorf("UniqueViews is not 1")
	}

	if result.UniqueDownloads != 1 {
		t.Errorf("UniqueDownloads is not 1")
	}
}