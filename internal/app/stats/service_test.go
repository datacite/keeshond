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
	conn   *gorm.DB
	config *app.Config
}

func createMockEvents() []event.Event {
	events := []event.Event{}

	events = append(events, event.CreateMockEvent("view", "example.com", "10.1234/1", 123, time.Date(2022, 01, 01, 00, 00, 00, 000, time.Local)))
	events = append(events, event.CreateMockEvent("view", "example.com", "10.1234/1", 123, time.Date(2022, 01, 01, 00, 00, 15, 000, time.Local)))
	events = append(events, event.CreateMockEvent("view", "example.com", "10.1234/1", 123, time.Date(2022, 01, 01, 00, 00, 30, 000, time.Local)))
	events = append(events, event.CreateMockEvent("view", "example.com", "10.1234/1", 123, time.Date(2022, 01, 01, 00, 10, 30, 000, time.Local)))
	events = append(events, event.CreateMockEvent("download", "example.com", "10.1234/1", 123, time.Date(2022, 01, 01, 00, 00, 30, 000, time.Local)))
	events = append(events, event.CreateMockEvent("download", "example.com", "10.1234/1", 123, time.Date(2022, 01, 01, 00, 10, 30, 000, time.Local)))
	events = append(events, event.CreateMockEvent("view", "example.com", "10.1234/1", 124, time.Date(2022, 01, 01, 00, 11, 30, 000, time.Local)))
	events = append(events, event.CreateMockEvent("view", "example.com", "10.1234/1", 124, time.Date(2022, 01, 01, 01, 00, 30, 000, time.Local)))
	events = append(events, event.CreateMockEvent("view", "example.com", "10.1234/1", 124, time.Date(2022, 01, 01, 02, 00, 30, 000, time.Local)))
	events = append(events, event.CreateMockEvent("download", "example.com", "10.1234/1", 124, time.Date(2022, 01, 01, 02, 00, 30, 000, time.Local)))

	events = append(events, event.CreateMockEvent("view", "example.com", "10.1234/2", 123, time.Date(2022, 01, 01, 00, 00, 00, 000, time.Local)))
	events = append(events, event.CreateMockEvent("view", "example.com", "10.1234/2", 123, time.Date(2022, 01, 01, 00, 00, 29, 000, time.Local)))
	events = append(events, event.CreateMockEvent("view", "example.com", "10.1234/2", 124, time.Date(2022, 01, 01, 00, 00, 30, 000, time.Local)))
	events = append(events, event.CreateMockEvent("view", "example.com", "10.1234/2", 123, time.Date(2022, 01, 01, 00, 10, 00, 000, time.Local)))
	events = append(events, event.CreateMockEvent("download", "example.com", "10.1234/2", 123, time.Date(2022, 01, 01, 00, 00, 00, 000, time.Local)))
	events = append(events, event.CreateMockEvent("download", "example.com", "10.1234/2", 124, time.Date(2022, 01, 01, 00, 00, 29, 000, time.Local)))
	events = append(events, event.CreateMockEvent("download", "example.com", "10.1234/2", 124, time.Date(2022, 01, 01, 00, 00, 30, 000, time.Local)))
	events = append(events, event.CreateMockEvent("download", "example.com", "10.1234/2", 123, time.Date(2022, 01, 01, 00, 10, 00, 000, time.Local)))
	events = append(events, event.CreateMockEvent("view", "example.com", "10.1234/2", 123, time.Date(2022, 01, 01, 00, 11, 30, 000, time.Local)))
	events = append(events, event.CreateMockEvent("view", "example.com", "10.1234/2", 123, time.Date(2022, 01, 01, 01, 00, 30, 000, time.Local)))
	events = append(events, event.CreateMockEvent("view", "example.com", "10.1234/2", 123, time.Date(2022, 01, 01, 02, 00, 30, 000, time.Local)))
	events = append(events, event.CreateMockEvent("download", "example.com", "10.1234/2", 124, time.Date(2022, 01, 01, 02, 00, 30, 000, time.Local)))

	return events
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

func setup() TestState {
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
		conn:   conn,
		config: config,
	}

	mockEvents := createMockEvents()

	// Setup repository and services
	sessionRepository := session.NewSessionRepository(conn, config)
	sessionService := session.NewSessionService(sessionRepository, config)
	eventRepository := event.NewEventRepository(conn, config)
	eventService := event.NewEventService(eventRepository, sessionService, config)

	// Insert mock events
	for _, event := range mockEvents {
		eventService.CreateRaw(event)
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
	start := time.Date(2022, 01, 01, 00, 00, 00, 000, time.Local)

	// End of day
	end := start.Add(24 * time.Hour)

	// Construct query
	query := Query{
		Start:  start,
		End:    end,
		Period: "custom",
	}

	statsRepository := NewStatsRepository(conn)
	statsService := NewStatsService(statsRepository)

	// Get stats
	result := statsService.Aggregate("example.com", query)

	if result.TotalDownloads != 7 {
		t.Errorf("TotalDownloads is not 7 but got %d", result.TotalDownloads)
	}

	if result.TotalViews != 12 {
		t.Errorf("TotalViews is not 12 but got %d", result.TotalViews)
	}

	if result.UniqueViews != 6 {
		t.Errorf("UniqueViews is not 6 but got %d", result.UniqueViews)
	}

	if result.UniqueDownloads != 3 {
		t.Errorf("UniqueDownloads is not 3 but got %d", result.UniqueDownloads)
	}
}

func TestStatsService_Timeseries(t *testing.T) {
	// Test config
	config := app.GetConfigFromEnv()
	config.ValidateDoi = false
	config.Database.Dbname = "keeshond_test"

	conn, err := setupTestDB(config)
	if err != nil {
		// Fail
		t.Errorf("Error connecting to test database: %s", err)
	}

	statsRepository := NewStatsRepository(conn)
	statsService := NewStatsService(statsRepository)

	// Start of today
	start := time.Date(2022, 01, 01, 00, 00, 00, 000, time.Local)

	// End of the day
	end := start.Add(24 * time.Hour)

	// Construct query for timeseries by hour
	query := Query{
		Start:    start,
		End:      end,
		Period:   "custom",
		Interval: "hour",
	}

	// Get stats
	result := statsService.Timeseries("example.com", query)

	if len(result) != 24 {
		t.Errorf("Timeseries should have 24 rows to represent 24 hours")
	}

	// Get value representing current hour
	testHour := time.Date(2022, 01, 01, 00, 00, 00, 000, time.Local).Hour()
	// Look through results and match current hour to date
	for _, row := range result {
		if row.Date.Hour() == testHour {
			if row.TotalDownloads != 5 {
				t.Errorf("Downloads for current hour should be 5 but got %d", row.TotalDownloads)
			}
			if row.TotalViews != 8 {
				t.Errorf("Views for current hour should be 8 but got %d", row.TotalViews)
			}
			if row.UniqueDownloads != 2 {
				t.Errorf("Unique downloads for current hour should be 2 but got %d", row.UniqueDownloads)
			}
			if row.UniqueViews != 2 {
				t.Errorf("Unique views for current hour should be 2 but got %d", row.UniqueViews)
			}
		}
	}
}

func TestStatsService_BreakdownByPID(t *testing.T) {
	// Test config
	config := app.GetConfigFromEnv()
	config.ValidateDoi = false
	config.Database.Dbname = "keeshond_test"

	conn, err := setupTestDB(config)
	if err != nil {
		// Fail
		t.Errorf("Error connecting to test database: %s", err)
	}

	statsRepository := NewStatsRepository(conn)
	statsService := NewStatsService(statsRepository)

	// Start of today
	start := time.Date(2022, 01, 01, 00, 00, 00, 000, time.Local)

	// End of the day
	end := start.Add(24 * time.Hour)

	// Construct query for timeseries by hour
	query := Query{
		Start:  start,
		End:    end,
		Period: "custom",
	}

	// Get stats
	result := statsService.BreakdownByPID("example.com", query, 1, 100)

	if len(result) != 2 {
		t.Errorf("BreakdownByPID should have 2 rows but got %d", len(result))
	}
}

func TestStatsService_CountBreakdownByPID(t *testing.T) {
	// Test config
	config := app.GetConfigFromEnv()
	config.ValidateDoi = false
	config.Database.Dbname = "keeshond_test"

	conn, err := setupTestDB(config)
	if err != nil {
		// Fail
		t.Errorf("Error connecting to test database: %s", err)
	}

	statsRepository := NewStatsRepository(conn)
	statsService := NewStatsService(statsRepository)

	// Start of today
	start := time.Date(2022, 01, 01, 00, 00, 00, 000, time.Local)

	// End of the day
	end := start.Add(24 * time.Hour)

	// Construct query for timeseries by hour
	query := Query{
		Start:  start,
		End:    end,
		Period: "custom",
	}

	// Get stats
	result := statsService.CountBreakdownByPID("example.com", query)

	if result != 2 {
		t.Errorf("CountBreakdownByPID should have returned 2 but got %d", result)
	}
}