package app

import (
	"os"
)

type Config struct {
	HTTP struct {
		Addr string
	}

	Database struct {
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

	// Database config
	config.Database.Host = getEnv("DATABASE_HOST", "localhost")
	config.Database.Port = getEnv("DATABASE_PORT", "5432")
	config.Database.User = getEnv("DATABASE_USER", "postgres")
	config.Database.Dbname = getEnv("DATABASE_DBNAME", "keeshond")
	config.Database.Password = getEnv("DATABASE_PASSWORD", "postgres")
	config.Database.Sslmode = getEnv("DATABASE_SSLMODE", "disable")

	// Validate DOI
	config.ValidateDoi = getEnv("VALIDATE_DOI", "true") == "true"

	return &config
}
