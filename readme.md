# Keeshond - DataCite Usage Analytics

This is part of the DataCite Usage Analytics service.

## Event Tracking

This has a public API that is the main endpoint for the [DataCite Tracker](https://github.com/datacite/datacite-tracker)
Events are stored within a Clickhouse database and then statistics according to COUNTER can be calculated.

### Setup

1. Install the tracking script [DataCite Tracker](https://github.com/datacite/datacite-tracker)
2. Configure using appropriate details
3. Results should be sent to the /api/metric end point
4. You can use the check api endpoint /api/check/{repo_id} to see if results are being recorded, it returns 200 and the timestamp of the last event if successful.

## COUNTER Usage Report Generation

Based on the data stored in Clickhouse and statistics that can be generated, usage reports in the format of SUSHI Json can be generated.
This can then be sent through to the [DataCite Reports API](https://support.datacite.org/docs/usage-reports-api-guide) for storage and processing into [DataCite Event Data](https://support.datacite.org/docs/eventdata-guide)


## Development

Requirements:

* Go 1.19

### Event Tracking Web Server

#### Running locally

```bash
# Start the http server
go run cmd/web/main.go
```

### Docker

```bash
# Build the Docker image
$ docker build -f ./docker/web/Dockerfile -t keeshondweb .
# and you can run the image with the following command
$ docker run -p 8081:8081 --rm -ti keeshondweb
```

### Usage Report Generation - Worker

This is triggered via a worker script, note that this will automatically submit the usage report to the Usage Reports API.

#### Config

The variables needed for the report generation are taken from Environment variables

REPO_ID - The unique tracking id for a repository, this is used for which stats to collect. This is assigned by DataCite.
BEGIN_DATE - The reporting period start date, typically this will be the start of a month.
END_DATE - The reporting perioid end date, typically this will be the end of a month.
PLATFORM - The name or identifier of the platform that the usage is from.
PUBLISHER - The name of publisher of the dataset
PUBLISHER_ID - The identifier of publisher of the dataset

In addition a valid DataCite JWT will need to be supplied for authentication and submission to the Usage Reports API.

DATACITE_JWT - Valid JWT with correct permissions. This is assigned by DataCite.

#### Running Locally

A report can be triggered using the worker version of the application.

e.g.

```bash
REPO_ID=datacite.demo BEGIN_DATE=2022-01-01 END_DATE=2022-12-31 PLATFORM=datacite PUBLISHER="datacite demo" PUBLISHER_ID=datacite.demo go run cmd/worker/main.go
```

#### Running via docker container

```bash
# Build worker image
docker build -f ./docker/worker/Dockerfile -t keeshondworker .

# Run docker with env vars
docker run --network="host" --env REPO_ID=datacite.demo --env BEGIN_DATE=2022-01-01 --env END_DATE=2022-12-31 --env PLATFORM=datacite --env PUBLISHER="datacite demo" --env PUBLISHER_ID=datacite.demo keeshondworker

```