package stats

type Service struct {
	repository RepositoryReader
}

// NewService creates a new stats service
func NewService(repository RepositoryReader) *Service {
	return &Service{
		repository: repository,
	}
}

func (service *Service) GetTotalInToday(metricName string, repoid string, pid string) int64 {
	return service.repository.GetTotalInToday(metricName, repoid, pid)
}

func (service *Service) GetTotalInLast7Days(metricName string, repoid string, pid string) int64 {
	return service.repository.GetTotalInLast7Days(metricName, repoid, pid)
}

func (service *Service) GetTotalInLast30Days(metricName string, repoid string, pid string) int64 {
	return service.repository.GetTotalInLast30Days(metricName, repoid, pid)
}
