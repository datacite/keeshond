package main

import (
	"log"
	"os"

	"github.com/datacite/keeshond"
)

type Config struct {
	HTTP struct {
		Addr string
	}

	Plausible struct {
		Url string
	}

	DataCite struct {
		Url string
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// Get configuration from environment variables.
	config := Config{}
	config.HTTP.Addr = getEnv("HTTP_ADDR", ":8081")
	config.Plausible.Url = getEnv("PLAUSIBLE_URL", "http://localhost:8100/")
	config.DataCite.Url = getEnv("DATACITE_API_URL", "http://api.stage.datacite.org")

	// Run with configuration.
	if err := run(&config); err != nil {
		log.Fatal(err)
	}
}

func run(config *Config) error {
	server := keeshond.NewServer()
	server.Addr = config.HTTP.Addr
	server.PlausibleUrl = config.Plausible.Url
	server.DataCiteApiUrl = config.DataCite.Url

	// Open the server.
	if err := server.Open(); err != nil {
		return err
	}

	return nil
}
