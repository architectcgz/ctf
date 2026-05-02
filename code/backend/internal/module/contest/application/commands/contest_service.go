package commands

import (
	"go.uber.org/zap"

	contestports "ctf-platform/internal/module/contest/ports"
)

type ContestService struct {
	repo    contestCommandRepository
	awdRepo contestports.AWDReadinessQuery
	log     *zap.Logger
}

type contestCommandRepository interface {
	contestports.ContestLookupRepository
	contestports.ContestWriteRepository
}

func NewContestService(repo contestCommandRepository, awdRepo contestports.AWDReadinessQuery, log *zap.Logger) *ContestService {
	if log == nil {
		log = zap.NewNop()
	}
	return &ContestService{repo: repo, awdRepo: awdRepo, log: log}
}
