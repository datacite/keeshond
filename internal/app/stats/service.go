package stats

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