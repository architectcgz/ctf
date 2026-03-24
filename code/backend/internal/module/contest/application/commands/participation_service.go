package commands

import contestports "ctf-platform/internal/module/contest/ports"

type ParticipationService struct {
	contestRepo contestports.Repository
	repo        contestports.ContestParticipationRepository
	teamRepo    contestports.ContestTeamFinder
}

func NewParticipationService(contestRepo contestports.Repository, repo contestports.ContestParticipationRepository, teamRepo contestports.ContestTeamFinder) *ParticipationService {
	return &ParticipationService{
		contestRepo: contestRepo,
		repo:        repo,
		teamRepo:    teamRepo,
	}
}
