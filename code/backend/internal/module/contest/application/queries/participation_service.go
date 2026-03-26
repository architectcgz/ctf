package queries

import contestports "ctf-platform/internal/module/contest/ports"

type ParticipationService struct {
	contestRepo contestports.ContestLookupRepository
	repo        contestports.ContestParticipationRepository
	teamRepo    contestports.ContestTeamFinder
}

func NewParticipationService(contestRepo contestports.ContestLookupRepository, repo contestports.ContestParticipationRepository, teamRepo contestports.ContestTeamFinder) *ParticipationService {
	return &ParticipationService{
		contestRepo: contestRepo,
		repo:        repo,
		teamRepo:    teamRepo,
	}
}
