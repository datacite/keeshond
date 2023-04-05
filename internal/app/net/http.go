package net

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/datacite/keeshond/internal/app"
	"github.com/datacite/keeshond/internal/app/auth"
	"github.com/datacite/keeshond/internal/app/event"
	"github.com/datacite/keeshond/internal/app/session"
	"github.com/datacite/keeshond/internal/app/stats"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"gorm.io/gorm"
)

type Http struct {
	server *http.Server
	router *chi.Mux
	config *app.Config
	db     *gorm.DB
	tokenAuth *jwtauth.JWTAuth

	eventServiceDB        *event.EventService

	statsService *stats.StatsService
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewHttpServer(config *app.Config, db *gorm.DB) *Http {

	tokenAuth := auth.GetAuthToken(config)

	// Create a new server that wraps the net/http server & add a router.
	s := &Http{
		server: &http.Server{},
		router: chi.NewRouter(),
		config: config,
		db:     db,
		tokenAuth: tokenAuth,
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
	eventRepositoryDB := event.NewEventRepository(s.db, config)

	sessionRepository := session.NewSessionRepository(s.db, config)
	sessionService := session.NewSessionService(sessionRepository, config)

	eventServiceDB := event.NewEventService(eventRepositoryDB, sessionService, config)

	statsRepository := stats.NewStatsRepository(s.db)
	statsService := stats.NewStatsService(statsRepository)
	s.statsService = statsService

	s.eventServiceDB = eventServiceDB

	// Register routes.
	s.router.Get("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	s.router.Get("/api/check/{repoId}", s.check)

	s.router.Post("/api/metric", s.createMetric)

	// Protected routes
	s.router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(s.tokenAuth))
		r.Use(jwtauth.Authenticator)

		r.Get("/api/stats/aggregate/{repoId}", s.getAggregate)
		r.Get("/api/stats/timeseries/{repoId}", s.getTimeseries)
		r.Get("/api/stats/breakdown/{repoId}", s.getBreakdown)
	})

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

func (s *Http) check(w http.ResponseWriter, r *http.Request) {
	repoId := chi.URLParam(r, "repoId")

	// Get last event for repoId
	result, hasResult := s.statsService.LastEvent(repoId)

	if !hasResult {
		http.Error(w, "No events found", http.StatusNotFound)
		return
	}

	// Return timestamp from last event in iso8601 format
	w.Write([]byte(result.Timestamp.Format("2006-01-02T15:04:05Z")))
}


// Function to check if useragent is a bot
func isBot(userAgent string) bool {
	// Read file with known bots
	bots, err := ioutil.ReadFile("data/COUNTER_Robots_list.json")

	if err != nil {
		log.Fatal(err)
	}

	// Read json file which is a list of objects containing pattern and last changed date
	var botsList []map[string]interface{}
	json.Unmarshal(bots, &botsList)

	// Loop through list of bots and check if useragent matches pattern
	for _, bot := range botsList {
		pattern := bot["pattern"].(string)
		regex := regexp.MustCompile("(?i)"+pattern)

		if regex.MatchString(userAgent) {
			return true
		}
	}

	return false
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

	// Print useragent
	log.Println(r.UserAgent())

	// Return a bad request if useragent is a bot
	if isBot(r.UserAgent()) {
		http.Error(w, "Event request denied due to known bot", http.StatusBadRequest)
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

	// Create event db
	if _, err := s.eventServiceDB.CreateEvent(&eventRequest); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	startDate, endDate, err := stats.ParsePeriodString(period, date)

	if err != nil {
		errorResponse(w, err)
		return
	}

	query := stats.Query{
		Start:  startDate,
		End:    endDate,
	}

	// Get total views for a repository in query period
	results := s.statsService.Aggregate(repoId, query)

	// Put results inside results object
	data := make(map[string]interface{})
	data["results"] = results

	// Set json response headers
	w.Header().Set("Content-Type", "application/json")

	// Serialise results but put inside a json object
	json.NewEncoder(w).Encode(data)
}

func (s *Http) getTimeseries(w http.ResponseWriter, r *http.Request) {
	repoId := chi.URLParam(r, "repoId")
	period := r.URL.Query().Get("period")
	date := r.URL.Query().Get("date")
	interval := r.URL.Query().Get("interval")

	startDate, endDate, err := stats.ParsePeriodString(period, date)

	if err != nil {
		errorResponse(w, err)
		return
	}

	query := stats.Query{
		Start:  startDate,
		End:    endDate,
		Interval: interval,
	}

	// Get total views for a repository in query period
	results := s.statsService.Timeseries(repoId, query)

	// Put results inside results object
	data := make(map[string]interface{})
	data["results"] = results

	// Set json response headers
	w.Header().Set("Content-Type", "application/json")

	// Serialise results but put inside a json object
	json.NewEncoder(w).Encode(data)
}

func (s *Http) getBreakdown(w http.ResponseWriter, r *http.Request) {
	repoId := chi.URLParam(r, "repoId")
	period := r.URL.Query().Get("period")
	date := r.URL.Query().Get("date")

	// Get page and pageSize as integers from query string
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil {
		pageSize = 100
	}

	startDate, endDate, err := stats.ParsePeriodString(period, date)

	if err != nil {
		errorResponse(w, err)
		return
	}

	query := stats.Query{
		Start:  startDate,
		End:    endDate,
	}

	// Get total views for a repository based on query
	results := s.statsService.BreakdownByPID(repoId, query, page, pageSize)

	// Put results inside results object
	data := make(map[string]interface{})
	data["results"] = results

	// Set json response headers
	w.Header().Set("Content-Type", "application/json")

	// Serialise results but put inside a json object
	json.NewEncoder(w).Encode(data)
}