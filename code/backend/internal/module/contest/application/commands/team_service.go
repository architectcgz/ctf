package commands

import "ctf-platform/internal/module/contest/ports"

type TeamService struct {
	teamRepo    ports.ContestTeamRepository
	contestRepo ports.Repository
}

func NewTeamService(teamRepo ports.ContestTeamRepository, contestRepo ports.Repository) *TeamService {
	return &TeamService{
		teamRepo:    teamRepo,
		contestRepo: contestRepo,
	}
}
