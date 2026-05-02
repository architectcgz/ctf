package queries

import (
	contestports "ctf-platform/internal/module/contest/ports"
)

type teamQueryRepository interface {
	contestports.ContestTeamFinder
	contestports.ContestTeamLookupRepository
	contestports.ContestTeamMembershipRepository
	contestports.ContestTeamListRepository
	contestports.ContestTeamUserLookupRepository
}

type TeamService struct {
	teamRepo    teamQueryRepository
	contestRepo contestports.ContestLookupRepository
}

func NewTeamService(teamRepo teamQueryRepository, contestRepo contestports.ContestLookupRepository) *TeamService {
	return &TeamService{
		teamRepo:    teamRepo,
		contestRepo: contestRepo,
	}
}
