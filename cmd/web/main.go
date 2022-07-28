package main

import (
	"log"

	"github.com/datacite/keeshond/internal/app"
	"github.com/datacite/keeshond/internal/app/net"
)

func main() {
	// Get configuration from environment variables.
	var config = app.GetConfigFromEnv()

	// Run with configuration.
	if err := run(config); err != nil {
		log.Fatal(err)
	}
}

func run(config *app.Config) error {
	server := net.NewHttpServer(config)

	// Open the server.
	if err := server.Open(); err != nil {
		return err
	}

	return nil
}
