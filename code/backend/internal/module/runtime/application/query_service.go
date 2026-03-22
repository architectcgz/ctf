package application

type QueryService struct {
	repo CountRunningRepository
}

func NewQueryService(repo CountRunningRepository) *QueryService {
	return &QueryService{repo: repo}
}

func (s *QueryService) CountRunning() (int64, error) {
	if s == nil || s.repo == nil {
		return 0, nil
	}
	return s.repo.CountRunning()
}
