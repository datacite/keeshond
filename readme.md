# Keeshond - DataCite metrics microservice

This is the main endpoint for the [DataCite Tracker](https://github.com/datacite/datacite-tracker)
Events are stored within a Clickhouse database and then statistics according to COUNTER can be calculated.

## Development

Requirements:

* Go 1.19

### Running locally

go run cmd/web/main.go - Starts the local HTTP server

### Docker

```bash
# Build the Docker image
$ docker build -t keeshondweb .
# and you can run the image with the following command
$ docker run -p 8081:8081 --rm -ti keeshondweb
```