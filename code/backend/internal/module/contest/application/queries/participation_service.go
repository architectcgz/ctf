package queries

import contestports "ctf-platform/internal/module/contest/ports"

type participationQueryRepository interface {
	contestports.ContestParticipationRegistrationLookupRepository
	contestports.ContestParticipationRegistrationListRepository
	contestports.ContestParticipationAnnouncementReadRepository
	contestports.ContestParticipationProgressRepository
}

type ParticipationService struct {
	contestRepo contestports.ContestLookupRepository
	repo        participationQueryRepository
	teamRepo    contestports.ContestTeamFinder
}

func NewParticipationService(contestRepo contestports.ContestLookupRepository, repo participationQueryRepository, teamRepo contestports.ContestTeamFinder) *ParticipationService {
	return &ParticipationService{
		contestRepo: contestRepo,
		repo:        repo,
		teamRepo:    teamRepo,
	}
}
