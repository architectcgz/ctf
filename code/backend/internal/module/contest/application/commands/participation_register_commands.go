package commands

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *ParticipationService) RegisterContest(ctx context.Context, contestID, userID int64) error {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if contest.Status != model.ContestStatusRegistration {
		return errcode.ErrContestRegistrationClosed
	}

	var teamID *int64
	team, err := s.teamRepo.FindUserTeamInContest(ctx, userID, contestID)
	if err == nil && team != nil && team.ID > 0 {
		teamID = &team.ID
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.ErrInternal.WithCause(err)
	}

	now := time.Now()
	registration, err := s.repo.FindRegistration(ctx, contestID, userID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrInternal.WithCause(err)
		}
		registration = &model.ContestRegistration{
			ContestID: contestID,
			UserID:    userID,
			TeamID:    teamID,
			Status:    model.ContestRegistrationStatusPending,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if createErr := s.repo.CreateRegistration(ctx, registration); createErr != nil {
			return errcode.ErrInternal.WithCause(createErr)
		}
		return nil
	}

	if registration.Status != model.ContestRegistrationStatusApproved {
		registration.Status = model.ContestRegistrationStatusPending
		registration.ReviewedBy = nil
		registration.ReviewedAt = nil
	}
	registration.TeamID = teamID
	registration.UpdatedAt = now
	if err := s.repo.SaveRegistration(ctx, registration); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}
