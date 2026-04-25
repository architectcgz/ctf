package commands

import (
	"context"
	"errors"

	"gorm.io/gorm"

	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *SubmissionService) resolveTeamID(ctx context.Context, userID, contestID int64) (*int64, error) {
	registration, err := s.repo.FindRegistration(ctx, contestID, userID)
	if err == nil {
		if err := contestdomain.RegistrationStatusError(registration.Status); err != nil {
			return nil, err
		}
		return registration.TeamID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	team, err := s.teamRepo.FindUserTeamInContest(ctx, userID, contestID)
	if err == nil && team.ID > 0 {
		return &team.ID, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return nil, errcode.ErrNotRegistered
}
