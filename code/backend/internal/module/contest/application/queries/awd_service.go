package queries

import contestports "ctf-platform/internal/module/contest/ports"

type AWDService struct {
	repo        contestports.AWDQueryRepository
	contestRepo contestports.ContestLookupRepository
}

func NewAWDService(repo contestports.AWDQueryRepository, contestRepo contestports.ContestLookupRepository) *AWDService {
	return &AWDService{
		repo:        repo,
		contestRepo: contestRepo,
	}
}
