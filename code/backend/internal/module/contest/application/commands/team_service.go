package commands

import "ctf-platform/internal/module/contest/ports"

type TeamService struct {
	teamRepo    ports.ContestTeamRepository
	contestRepo ports.ContestLookupRepository
}

func NewTeamService(teamRepo ports.ContestTeamRepository, contestRepo ports.ContestLookupRepository) *TeamService {
	return &TeamService{
		teamRepo:    teamRepo,
		contestRepo: contestRepo,
	}
}
