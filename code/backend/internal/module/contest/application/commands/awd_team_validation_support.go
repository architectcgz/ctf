package commands

import (
	"context"
	"errors"

	"gorm.io/gorm"

	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) resolveUserTeamID(ctx context.Context, userID, contestID int64) (int64, error) {
	registration, err := s.repo.FindRegistration(ctx, contestID, userID)
	if err == nil {
		if err := contestdomain.RegistrationStatusError(registration.Status); err != nil {
			return 0, err
		}
		if registration.TeamID == nil || *registration.TeamID <= 0 {
			return 0, errcode.ErrAWDTeamRequired
		}
		return *registration.TeamID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errcode.ErrInternal.WithCause(err)
	}

	team, err := s.repo.FindContestTeamByMember(ctx, contestID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errcode.ErrNotRegistered
		}
		return 0, errcode.ErrInternal.WithCause(err)
	}
	return team.ID, nil
}
