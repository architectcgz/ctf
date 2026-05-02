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

func (s *ParticipationService) ReviewRegistration(ctx context.Context, contestID, registrationID, reviewerID int64, req ReviewRegistrationInput) (*dto.ContestRegistrationResp, error) {
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

	resp := contestResponseMapperInst.ToContestRegistrationRespBasePtr(registration)
	resp.Username = user.Username
	return resp, nil
}
