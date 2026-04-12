package commands

import (
	"go.uber.org/zap"

	contestports "ctf-platform/internal/module/contest/ports"
)

type ContestService struct {
	repo    contestports.ContestCommandRepository
	awdRepo contestports.AWDRepository
	log     *zap.Logger
}

func NewContestService(repo contestports.ContestCommandRepository, awdRepo contestports.AWDRepository, log *zap.Logger) *ContestService {
	if log == nil {
		log = zap.NewNop()
	}
	return &ContestService{repo: repo, awdRepo: awdRepo, log: log}
}
