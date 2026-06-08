package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"
	"time"

	"ctf-platform/internal/apperror"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *ParticipationService) RegisterContest(ctx context.Context, contestID, userID int64) error {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return contestcontracts.ErrContestNotFound
		}
		return apperror.ErrInternal.WithCause(err)
	}
	if contest.Status != contestentity.ContestStatusRegistration {
		return contestcontracts.ErrContestRegistrationClosed
	}

	var teamID *int64
	team, err := s.teamRepo.FindUserTeamInContest(ctx, userID, contestID)
	if err == nil && team != nil && team.ID > 0 {
		teamID = &team.ID
	} else if err != nil && !errors.Is(err, contestports.ErrContestUserTeamNotFound) {
		return apperror.ErrInternal.WithCause(err)
	}

	now := time.Now().UTC()
	registration, err := s.repo.FindRegistration(ctx, contestID, userID)
	if err != nil {
		if !errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
			return apperror.ErrInternal.WithCause(err)
		}
		registration = &contestentity.ContestRegistration{
			ContestID: contestID,
			UserID:    userID,
			TeamID:    teamID,
			Status:    contestentity.ContestRegistrationStatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if createErr := s.repo.CreateRegistration(ctx, registration); createErr != nil {
			return apperror.ErrInternal.WithCause(createErr)
		}
		return nil
	}

	if registration.Status != contestentity.ContestRegistrationStatusApproved {
		registration.Status = contestentity.ContestRegistrationStatusPending
		registration.ReviewedBy = nil
		registration.ReviewedAt = nil
	}
	registration.TeamID = teamID
	registration.UpdatedAt = now
	if err := s.repo.SaveRegistration(ctx, registration); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}
