package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"

	"ctf-platform/internal/apperror"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *AWDService) resolveUserTeamID(ctx context.Context, userID, contestID int64) (int64, error) {
	registration, err := s.repo.FindRegistration(ctx, contestID, userID)
	if err == nil {
		if err := contestdomain.RegistrationStatusError(registration.Status); err != nil {
			return 0, err
		}
		if registration.TeamID == nil || *registration.TeamID <= 0 {
			return 0, contestcontracts.ErrAWDTeamRequired
		}
		return *registration.TeamID, nil
	}
	if !errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
		return 0, apperror.ErrInternal.WithCause(err)
	}

	team, err := s.repo.FindContestTeamByMember(ctx, contestID, userID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestUserTeamNotFound) {
			return 0, contestcontracts.ErrNotRegistered
		}
		return 0, apperror.ErrInternal.WithCause(err)
	}
	return team.ID, nil
}
