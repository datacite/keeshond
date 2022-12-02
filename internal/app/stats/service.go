package stats

import (
	"errors"
	"strings"
	"time"
)

type StatsService struct {
	repository StatsRepositoryReader
}

// NewStatsService creates a new stats service
func NewStatsService(repository StatsRepositoryReader) *StatsService {
	return &StatsService{
		repository: repository,
	}
}

func (service *StatsService) Aggregate(repoId string, query Query) AggregateResult {
	return service.repository.Aggregate(repoId, query)
}

func (service *StatsService) Timeseries(repoId string, query Query) []TimeseriesResult {
	return service.repository.Timeseries(repoId, query)
}

func (service *StatsService) BreakdownByPID(repoId string, query Query, page int, pageSize int) []BreakdownResult {
	return service.repository.BreakdownByPID(repoId, query, page, pageSize)
}

func (service *StatsService) CountUniquePID(repoId string, query Query) int64 {
	return service.repository.CountUniquePID(repoId, query)
}


// Function to parse a period string into start and end time ranges relative to date
func ParsePeriodString(period string, date string) (time.Time, time.Time, error) {
	// Set default start and end times

	var startTime time.Time
	var endTime time.Time

	var relativeDate time.Time
	// If date is empty set it to start of today
	if date == "" {
		today := time.Now()
		relativeDate = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	} else {
		relativeDate, _ = time.Parse("2006-01-02", date)
	}
	// Set end date to always be end of the day to include all events
	endTime = relativeDate

	// Parse the period query
	switch period {
	case "day":
		// Start and end time are a full day based on the relative date
		startTime = time.Date(relativeDate.Year(), relativeDate.Month(), relativeDate.Day(), 0, 0, 0, 0, relativeDate.Location())
	case "7d":
		startTime = relativeDate.AddDate(0, 0, -6)
	case "custom":
		// Parse date range string
		if date != "" {
			// Split date string into start and end date
			dateRange := strings.Split(date, ",")
			if len(dateRange) != 2 {
				return startTime, endTime, errors.New("invalid date range")
			}

			var err error
			// Parse start date
			startTime, err = time.Parse("2006-01-02", dateRange[0])
			if err != nil {
				return startTime, endTime, errors.New("invalid start date")
			}

			// Parse end date
			endTime, err = time.Parse("2006-01-02", dateRange[1])
			if err != nil {
				return startTime, endTime, errors.New("invalid end date")
			}
		} else {
			return startTime, endTime, errors.New("no date specified for custom period")
		}
	case "30d":
		fallthrough
	default:
		startTime = relativeDate.AddDate(0, 0, -29)
	}

	endTime = endTime.AddDate(0, 0, 1)

	return startTime, endTime, nil
}