package net

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/datacite/keeshond/internal/app"
	"github.com/datacite/keeshond/internal/app/event"
	"github.com/datacite/keeshond/internal/app/session"
	"github.com/datacite/keeshond/internal/app/stats"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gorm.io/gorm"
)

type Http struct {
	server *http.Server
	router *chi.Mux
	config *app.Config
	db     *gorm.DB

	eventServiceDB        *event.EventService
	eventServicePlausible *event.EventService

	statsService *stats.StatsService
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewHttpServer(config *app.Config, db *gorm.DB) *Http {
	// Create a new server that wraps the net/http server & add a router.
	s := &Http{
		server: &http.Server{},
		router: chi.NewRouter(),
		config: config,
		db:     db,
	}

	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)

	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Register repositories and services
	eventRepositoryPlausible := event.NewRepositoryPlausible(config)
	eventRepositoryDB := event.NewEventRepository(s.db, config)

	sessionRepository := session.NewSessionRepository(s.db, config)
	sessionService := session.NewSessionService(sessionRepository, config)

	eventServicePlausible := event.NewEventService(eventRepositoryPlausible, sessionService, config)
	eventServiceDB := event.NewEventService(eventRepositoryDB, sessionService, config)

	statsRepository := stats.NewStatsRepository(s.db)
	statsService := stats.NewStatsService(statsRepository)
	s.statsService = statsService

	s.eventServicePlausible = eventServicePlausible
	s.eventServiceDB = eventServiceDB

	// Register routes.
	s.router.Get("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	s.router.Post("/api/metric", s.createMetric)

	// Create API routes for getting aggregate for metric statistics
	s.router.Get("/api/stats/aggregate/{repoId}", s.getAggregate)

	s.server.Handler = s.router

	return s
}

// Open validates the server options and begins listening on the bind address.
func (s *Http) Open() (err error) {
	s.server.Addr = s.config.HTTP.Addr

	log.Println("Server starting, listening on", s.config.HTTP.Addr)
	return s.server.ListenAndServe()
}

// Get remote IP Address
func getRemoteAddr(r *http.Request) string {
	// X-Forwarded-For: client, proxy1, proxy2, ...
	remoteAddr := strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
	if remoteAddr == "" {
		remoteAddr = r.RemoteAddr
	}

	remoteAddr = strings.Split(remoteAddr, ":")[0]
	return remoteAddr
}

type MetricRequest struct {
	Name   string `json:"n"`
	RepoId string `json:"i"`
	Url    string `json:"u"`
	Pid    string `json:"p"`
}

func (s *Http) createMetric(w http.ResponseWriter, r *http.Request) {
	// Metric request is different to a eventRequest as only some data comes
	// from the json body
	var metricRequest MetricRequest

	// Marshal json to metric
	if err := json.NewDecoder(r.Body).Decode(&metricRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get potential IP from request
	clientIp := getRemoteAddr(r)

	// Create event request from the metric request
	eventRequest := event.EventRequest{
		Name:      metricRequest.Name,
		RepoId:    metricRequest.RepoId,
		Url:       metricRequest.Url,
		Useragent: r.UserAgent(),
		ClientIp:  clientIp,
		Pid:       metricRequest.Pid,
	}

	// Validate Event Request
	if err := s.eventServiceDB.Validate(&eventRequest); err != nil {
		// Format error message
		errorMessage := fmt.Sprintf("%s - %s, Usage stats cannot be processed", eventRequest.Pid, err.Error())

		http.Error(w, errorMessage, http.StatusBadRequest)

		return
	}

	// Create event plausible
	if _, err := s.eventServicePlausible.CreateEvent(&eventRequest); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Create event db
	if _, err := s.eventServiceDB.CreateEvent(&eventRequest); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Function to parse the period query into start and end time ranges
func parsePeriodQuery(period string, date string) (time.Time, time.Time, error) {
	// Set default start and end times

	var startTime time.Time
	var endTime time.Time

	var relativeDate time.Time
	// If date is empty set it to start of today
	if date == "" {
		today := time.Now()
		relativeDate = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	} else {
		relativeDate, _ = time.Parse("2006-01-02", date)
	}
	// Set end date to always be end of the day to include all events
	endTime = relativeDate

	// Parse the period query
	if period != "" {
		switch period {
		case "day":
			// Start and end time are a full day based on the relative date
			startTime = time.Date(relativeDate.Year(), relativeDate.Month(), relativeDate.Day(), 0, 0, 0, 0, relativeDate.Location())
		case "7d":
			startTime = relativeDate.AddDate(0, 0, -6)
		case "30d":
			startTime = relativeDate.AddDate(0, 0, -29)
		case "custom":
			// Parse date range string
			if date != "" {
				// Split date string into start and end date
				dateRange := strings.Split(date, ",")
				if len(dateRange) != 2 {
					return startTime, endTime, errors.New("invalid date range")
				}

				var err error
				// Parse start date
				startTime, err = time.Parse("2006-01-02", dateRange[0])
				if err != nil {
					return startTime, endTime, errors.New("invalid start date")
				}

				// Parse end date
				endTime, err = time.Parse("2006-01-02", dateRange[1])
				if err != nil {
					return startTime, endTime, errors.New("invalid end date")
				}
			} else {
				return startTime, endTime, errors.New("no date specified for custom period")
			}
		default:
			return startTime, endTime, errors.New("invalid period query")
		}
	}

	endTime = endTime.AddDate(0, 0, 1).Add(-time.Second)

	return startTime, endTime, nil
}

// Take an error and return a json response
func errorResponse(w http.ResponseWriter, err error) {
	// Create error response
	errorResponse := ErrorResponse{
		Error: err.Error(),
	}

	// Marshal error response to json
	jsonResponse, _ := json.Marshal(errorResponse)

	// Write error response to response writer
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(jsonResponse)
}

func (s *Http) getAggregate(w http.ResponseWriter, r *http.Request) {
	repoId := chi.URLParam(r, "repoId")
	period := r.URL.Query().Get("period")
	date := r.URL.Query().Get("date")

	startDate, endDate, err := parsePeriodQuery(period, date)
	// Print start and end date
	log.Println(startDate, endDate)

	if err != nil {
		errorResponse(w, err)
		return
	}

	query := stats.Query{
		Start:  startDate,
		End:    endDate,
		Period: period,
	}

	// Get total views for a repository in last 30 days
	results := s.statsService.Aggregate(repoId, query)

	// Put results inside results object
	data := make(map[string]interface{})
	data["results"] = results

	// Set json response headers
	w.Header().Set("Content-Type", "application/json")

	// Serialise results but put inside a json object
	json.NewEncoder(w).Encode(data)
}