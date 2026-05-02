package commands

import "ctf-platform/internal/module/contest/ports"

type teamCommandRepository interface {
	ports.ContestTeamFinder
	ports.ContestTeamWriteRepository
	ports.ContestTeamLookupRepository
	ports.ContestTeamMembershipRepository
	ports.ContestTeamRegistrationLookupRepository
}

type TeamService struct {
	teamRepo    teamCommandRepository
	contestRepo ports.ContestLookupRepository
}

func NewTeamService(teamRepo teamCommandRepository, contestRepo ports.ContestLookupRepository) *TeamService {
	return &TeamService{
		teamRepo:    teamRepo,
		contestRepo: contestRepo,
	}
}
