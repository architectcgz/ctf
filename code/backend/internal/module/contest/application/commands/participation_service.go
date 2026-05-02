package commands

import contestports "ctf-platform/internal/module/contest/ports"

type participationCommandRepository interface {
	contestports.ContestParticipationRegistrationLookupRepository
	contestports.ContestParticipationRegistrationWriteRepository
	contestports.ContestParticipationUserLookupRepository
	contestports.ContestParticipationAnnouncementWriteRepository
}

type ParticipationService struct {
	contestRepo contestports.ContestLookupRepository
	repo        participationCommandRepository
	teamRepo    contestports.ContestTeamFinder
	broadcaster contestports.RealtimeBroadcaster
}

func NewParticipationService(contestRepo contestports.ContestLookupRepository, repo participationCommandRepository, teamRepo contestports.ContestTeamFinder) *ParticipationService {
	return &ParticipationService{
		contestRepo: contestRepo,
		repo:        repo,
		teamRepo:    teamRepo,
	}
}

func (s *ParticipationService) SetRealtimeBroadcaster(broadcaster contestports.RealtimeBroadcaster) {
	s.broadcaster = broadcaster
}
