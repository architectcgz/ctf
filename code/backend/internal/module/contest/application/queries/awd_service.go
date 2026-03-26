package queries

import contestports "ctf-platform/internal/module/contest/ports"

type AWDService struct {
	repo        contestports.AWDRepository
	contestRepo contestports.ContestLookupRepository
}

func NewAWDService(repo contestports.AWDRepository, contestRepo contestports.ContestLookupRepository) *AWDService {
	return &AWDService{
		repo:        repo,
		contestRepo: contestRepo,
	}
}
