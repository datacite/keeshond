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

func (service *StatsService) GetTotalInToday(metricName string, repoid string, pid string) int64 {
	return service.repository.GetTotalInToday(metricName, repoid, pid)
}

func (service *StatsService) GetTotalInLast7Days(metricName string, repoid string, pid string) int64 {
	return service.repository.GetTotalInLast7Days(metricName, repoid, pid)
}

func (service *StatsService) GetTotalInLast30Days(metricName string, repoid string, pid string) int64 {
	return service.repository.GetTotalInLast30Days(metricName, repoid, pid)
}
