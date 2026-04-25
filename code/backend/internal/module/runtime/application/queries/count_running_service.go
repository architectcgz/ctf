package queries

import "context"

import runtimeports "ctf-platform/internal/module/runtime/ports"

type CountRunningService struct {
	repo runtimeports.CountRunningRepository
}

func NewCountRunningService(repo runtimeports.CountRunningRepository) *CountRunningService {
	return &CountRunningService{repo: repo}
}

func (s *CountRunningService) CountRunning(ctx context.Context) (int64, error) {
	if s == nil || s.repo == nil {
		return 0, nil
	}
	return s.repo.CountRunning(ctx)
}
