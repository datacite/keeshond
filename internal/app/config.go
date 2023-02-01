package app

import (
	"os"
)

type Config struct {
	HTTP struct {
		Addr string
	}

	AnalyticsDatabase struct {
		Host string
		Port string
		User string
		Dbname string
		Password string
		Sslmode string
	}

	Plausible struct {
		Url string
	}

	DataCite struct {
		Url string
		JWT string
	}

	ValidateDoi bool
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GetConfigFromEnv() *Config {
	// Get configuration from environment variables.
	config := Config{}
	config.HTTP.Addr = getEnv("HTTP_ADDR", ":8081")
	config.Plausible.Url = getEnv("PLAUSIBLE_URL", "https://analytics.stage.datacite.org")
	config.DataCite.Url = getEnv("DATACITE_API_URL", "https://api.stage.datacite.org")
	config.DataCite.JWT = getEnv("DATACITE_JWT", "")

	// Database config
	config.AnalyticsDatabase.Host = getEnv("ANALYTICS_DATABASE_HOST", "localhost")
	config.AnalyticsDatabase.Port = getEnv("ANALYTICS_DATABASE_PORT", "9000")
	config.AnalyticsDatabase.User = getEnv("ANALYTICS_DATABASE_USER", "keeshond")
	config.AnalyticsDatabase.Dbname = getEnv("ANALYTICS_DATABASE_DBNAME", "keeshond")
	config.AnalyticsDatabase.Password = getEnv("ANALYTICS_DATABASE_PASSWORD", "keeshond")

	// Validate DOI
	config.ValidateDoi = getEnv("VALIDATE_DOI", "true") == "true"

	return &config
}
