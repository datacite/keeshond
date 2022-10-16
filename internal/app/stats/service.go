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

func (service *StatsService) GetTotalsByPidInToday(metricName string, repoid string) []struct{Pid string; Count int64} {
	return service.repository.GetTotalsByPidInToday(metricName, repoid)
}

func (service *StatsService) GetTotalsByPidInLast7Days(metricName string, repoid string) []struct{Pid string; Count int64} {
	return service.repository.GetTotalsByPidInLast7Days(metricName, repoid)
}

func (service *StatsService) GetTotalsByPidInLast30Days(metricName string, repoid string) []struct{Pid string; Count int64} {
	return service.repository.GetTotalsByPidInLast30Days(metricName, repoid)
}