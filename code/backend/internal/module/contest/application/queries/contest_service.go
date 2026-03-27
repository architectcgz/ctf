package queries

import (
	"go.uber.org/zap"

	contestports "ctf-platform/internal/module/contest/ports"
)

type ContestService struct {
	repo contestports.ContestListRepository
	log  *zap.Logger
}

func NewContestService(repo contestports.ContestListRepository, log *zap.Logger) *ContestService {
	if log == nil {
		log = zap.NewNop()
	}
	return &ContestService{repo: repo, log: log}
}
