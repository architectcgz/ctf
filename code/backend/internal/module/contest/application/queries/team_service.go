package queries

import (
	contestports "ctf-platform/internal/module/contest/ports"
)

type TeamService struct {
	teamRepo    contestports.ContestTeamRepository
	contestRepo contestports.ContestLookupRepository
}

func NewTeamService(teamRepo contestports.ContestTeamRepository, contestRepo contestports.ContestLookupRepository) *TeamService {
	return &TeamService{
		teamRepo:    teamRepo,
		contestRepo: contestRepo,
	}
}
