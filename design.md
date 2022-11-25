# Design

## Code organisation
- Model - The data model
- Repository - The interaction with the database or external services.
- Service - This is 'business' logic that works with the repository defined.
- net - Network services that are built on top i.e. http server for API

Functionality be packaged by what you're dealing with, so we avoid top level models/ services/ repositories/ instead you work with the specific areas i.e. Events, Users etc

## Folder Structure
- internal - Reserved by golang, but is where everything specific to app lives, i.e. nothing third parties could import
- app - Main package for all application code
- cmd - The main access to the functioanlity, different applications can be built depending on usecases using the internal code.


# Statistics API

### Metrics

- total_views - Total count for metric type 'view', duplicated events within 30 seconds removed.
- total_downloads - Total count for metric type 'download', duplicated events within 30 seconds removed.
- unique_views - Unique count for metric type 'view', filtered for unique by session_id
- unique_downloads - Unique count for metric type 'download', filtered for unique by session_id

### Time Periods

Time Periods are relative to a date, the date by default is the current day.

- day - A full day based on date
- 7d - Last 7 days relative to date
- 30d - Last 30 days relative to date
- custom - To provide a custom period using date parameter

Custom date ranges are set by specifying two ISO8601 timestamp on the date parameter
For example the month of January 2022:
?period=custom&date=2022-01-01,2022-01-31.

### Aggregates

Over a time period return an aggregate by metric types.
This is useful to get an overview, defaults to total_views,total_downloads.

#### Params
- repo_id (Required) - The repository identifier that your tracker is recording against.
- period - The period of time you want to aggregate over. Default 30d

#### Example
/api/stats/aggregate?repo_id=example.com&period=7d

{
  "results": {
    "unique_views": 5
    "total_views": 10
    "unique_downloads": 10
    "total_downloads": 40
  }
}

### Timeseries

Over a period of time show a breakdown over a time period.

#### Params
- repo_id (Required) - The repository identifier that your tracker is recording against.
- period - The period of time you want to aggregate over. Default 30d
- interval - Valid interval periods are "day", "month", "hour" - Defaults to day

#### Example

/api/stats/breakdown?repo_id=example.com&period=7d

{
  "results": [
    {
      "date": "2022-01-01",
      "total_views": 10,
      "total_downloads: "5",
      "unique_views: "5"
      "unique_downloads: "5"
    },
    {
      "date": "2022-01-02",
      "total_views": 10,
      "total_downloads: "5",
      "unique_views: "5"
      "unique_downloads: "5"
    },
    {
      "date": "2022-01-03",
      "total_views": 10,
      "total_downloads: "5",
      "unique_views: "5"
      "unique_downloads: "5"
    },
    {
      "date": "2022-01-04",
      "total_views": 10,
      "total_downloads: "5",
      "unique_views: "5"
      "unique_downloads: "5"
    },
    {
      "date": "2022-01-05",
      "total_views": 10,
      "total_downloads: "5",
      "unique_views: "5"
      "unique_downloads: "5"
    },
    {
      "date": "2022-01-06",
      "total_views": 10,
      "total_downloads: "5",
      "unique_views: "5"
      "unique_downloads: "5"
    },
    {
      "date": "2022-01-07",
      "total_views": 10,
      "total_downloads: "5",
      "unique_views: "5"
      "unique_downloads: "5"
    }
  ]
}

### Breakdown

Breakdown of metrics by PID

#### Params
- repo_id (Required) - The repository identifier that your tracker is recording against.
- period - The time range you want to aggregate over. Default 30d
- pageSize - Limit of results to return, maxium 1000. Can be combined with page for pagination of results.
- page - Which page of results to look at, starts at 1.

#### Example

/api/stats/breakdown?repo_id=example.com

{
  "results": [
    {
        "pid": "10.5072/12345"
        "total_views": 2
        "total_downloads": 10
    },
    {
        "pid": "10.5072/56789"
        "total_views": 5
        "total_downloads": 15
    }
  ]
}

