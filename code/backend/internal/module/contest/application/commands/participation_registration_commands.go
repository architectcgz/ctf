package commands

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
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
	team, err := s.teamRepo.FindUserTeamInContest(userID, contestID)
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

func (s *ParticipationService) ReviewRegistration(ctx context.Context, contestID, registrationID, reviewerID int64, req *dto.ReviewContestRegistrationReq) (*dto.ContestRegistrationResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	registration, err := s.repo.FindRegistrationByID(ctx, contestID, registrationID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrContestRegistrationNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if registration.Status != model.ContestRegistrationStatusPending {
		return nil, errcode.ErrInvalidStatusTransition
	}

	now := time.Now()
	registration.Status = req.Status
	registration.ReviewedBy = &reviewerID
	registration.ReviewedAt = &now
	registration.UpdatedAt = now
	if req.Status == model.ContestRegistrationStatusRejected {
		registration.TeamID = nil
	}
	if err := s.repo.SaveRegistration(ctx, registration); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	user, err := s.repo.FindUserByID(ctx, registration.UserID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.ContestRegistrationResp{
		ID:         registration.ID,
		ContestID:  registration.ContestID,
		UserID:     registration.UserID,
		Username:   user.Username,
		TeamID:     registration.TeamID,
		Status:     registration.Status,
		ReviewedBy: registration.ReviewedBy,
		ReviewedAt: registration.ReviewedAt,
		CreatedAt:  registration.CreatedAt,
		UpdatedAt:  registration.UpdatedAt,
	}, nil
}
