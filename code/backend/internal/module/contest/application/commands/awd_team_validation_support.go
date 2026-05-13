package commands

import (
	"context"
	"errors"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
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
	if !errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
		return 0, errcode.ErrInternal.WithCause(err)
	}

	team, err := s.repo.FindContestTeamByMember(ctx, contestID, userID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestUserTeamNotFound) {
			return 0, errcode.ErrNotRegistered
		}
		return 0, errcode.ErrInternal.WithCause(err)
	}
	return team.ID, nil
}
