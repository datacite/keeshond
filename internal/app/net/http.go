package net

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

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

	eventServiceDB *event.EventService
	eventServicePlausible *event.EventService

	statsService *stats.StatsService
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
	s.router.Get("/api/stats/aggregate/{repoId}", s.getStatistics)

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

func (s *Http) getStatistics(w http.ResponseWriter, r *http.Request) {
	repoId := chi.URLParam(r, "repoId")

	// Get total views for a repository in last 30 days
	totalViewsLast30Days := s.statsService.GetTotalsByPidInLast30Days("view", repoId)

	// Set json response headers
	w.Header().Set("Content-Type", "application/json")

	// Serialise to json if not null
	if totalViewsLast30Days != nil {
		json.NewEncoder(w).Encode(totalViewsLast30Days)
	} else {
		// Return empty array if no data
		json.NewEncoder(w).Encode([]int{})
	}
}