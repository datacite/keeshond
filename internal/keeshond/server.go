package keeshond

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	server *http.Server
	router *chi.Mux

	Addr string

	PlausibleUrl   string
	DataCiteApiUrl string
}

func NewServer() *Server {
	// Create a new server that wraps the net/http server & add a gorilla router.
	s := &Server{
		server: &http.Server{},
		router: chi.NewRouter(),
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

	// Register routes.
	s.router.Get("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	s.router.Post("/api/metric", s.createMetric)

	s.server.Handler = s.router

	return s
}

// Open validates the server options and begins listening on the bind address.
func (s *Server) Open() (err error) {
	s.server.Addr = s.Addr

	return s.server.ListenAndServe()
}

type MetricRequest struct {
	Name   string `json:"n"`
	RepoId string `json:"i"`
	Url    string `json:"u"`
	Pid    string `json:"p"`
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

func (s *Server) createMetric(w http.ResponseWriter, r *http.Request) {
	// Metric request is different to a metric event as only some data comes
	// from the json body
	var metricRequest MetricRequest

	// Marshal json to metric
	if err := json.NewDecoder(r.Body).Decode(&metricRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get potential IP from request
	clientIp := getRemoteAddr(r)

	// Create metric event from the metric request
	metricEvent := NewMetricEvent(metricRequest.Name, metricRequest.RepoId, time.Now(), metricRequest.Url, r.UserAgent(), clientIp, metricRequest.Pid)

	// Http client
	client := &http.Client{}

	// Validate PID
	if err := checkExistsInDataCite(metricEvent.Pid, metricEvent.Url, s.DataCiteApiUrl, client); err != nil {
		// Format error message
		errorMessage := fmt.Sprintf("%s - %s, Usage stats cannot be processed", metricEvent.Pid, err.Error())

		http.Error(w, errorMessage, http.StatusBadRequest)
	}

	// Save to plausible

	if err := SendMetricEventToPlausible(metricEvent, s.PlausibleUrl, client); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
