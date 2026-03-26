package queries

import runtimeports "ctf-platform/internal/module/runtime/ports"

type CountRunningService struct {
	repo runtimeports.CountRunningRepository
}

func NewCountRunningService(repo runtimeports.CountRunningRepository) *CountRunningService {
	return &CountRunningService{repo: repo}
}

func (s *CountRunningService) CountRunning() (int64, error) {
	if s == nil || s.repo == nil {
		return 0, nil
	}
	return s.repo.CountRunning()
}
