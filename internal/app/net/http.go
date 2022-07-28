package net

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/datacite/keeshond/internal/app"
	"github.com/datacite/keeshond/internal/app/event"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Http struct {
	server *http.Server
	router *chi.Mux
	config *app.Config

	eventService *event.Service
}

func NewHttpServer(config *app.Config) *Http {
	// Create a new server that wraps the net/http server & add a router.
	s := &Http{
		server: &http.Server{},
		router: chi.NewRouter(),
		config: config,
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
	eventRepository := event.NewRepositoryPlausible(config)
	eventService := event.NewService(eventRepository, config)
	s.eventService = eventService

	// Register routes.
	s.router.Get("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	s.router.Post("/api/metric", s.createMetric)

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
	eventRequest := event.Request{
		Name:      metricRequest.Name,
		RepoId:    metricRequest.RepoId,
		Url:       metricRequest.Url,
		Useragent: r.UserAgent(),
		ClientIp:  clientIp,
		Pid:       metricRequest.Pid,
	}

	// Validate Event Request
	if err := s.eventService.Validate(&eventRequest); err != nil {
		// Format error message
		errorMessage := fmt.Sprintf("%s - %s, Usage stats cannot be processed", eventRequest.Pid, err.Error())

		http.Error(w, errorMessage, http.StatusBadRequest)

		return
	}

	// Create event
	if _, err := s.eventService.CreateEvent(&eventRequest); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
