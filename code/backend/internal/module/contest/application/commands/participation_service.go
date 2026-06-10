package commands

import (
	contestports "ctf-platform/internal/module/contest/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type participationCommandRepository interface {
	contestports.ContestParticipationRegistrationLookupRepository
	contestports.ContestParticipationRegistrationWriteRepository
	contestports.ContestParticipationUserLookupRepository
	contestports.ContestParticipationAnnouncementWriteRepository
}

type ParticipationService struct {
	contestRepo          contestports.ContestLookupRepository
	repo                 participationCommandRepository
	announcementTxRunner contestports.ContestParticipationAnnouncementTxRunner
	teamRepo             contestports.ContestTeamFinder
	eventBus             platformevents.Bus
}

func NewParticipationService(contestRepo contestports.ContestLookupRepository, repo participationCommandRepository, teamRepo contestports.ContestTeamFinder) *ParticipationService {
	service := &ParticipationService{
		contestRepo: contestRepo,
		repo:        repo,
		teamRepo:    teamRepo,
	}
	if runner, ok := repo.(contestports.ContestParticipationAnnouncementTxRunner); ok {
		service.announcementTxRunner = runner
	}
	return service
}

func (s *ParticipationService) SetEventBus(bus platformevents.Bus) *ParticipationService {
	if s == nil {
		return nil
	}
	s.eventBus = bus
	return s
}
