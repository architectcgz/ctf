package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"

	"ctf-platform/internal/apperror"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *SubmissionService) resolveTeamID(ctx context.Context, userID, contestID int64) (*int64, error) {
	registration, err := s.repo.FindRegistration(ctx, contestID, userID)
	if err == nil {
		if err := contestdomain.RegistrationStatusError(registration.Status); err != nil {
			return nil, err
		}
		return registration.TeamID, nil
	} else if !errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	team, err := s.teamRepo.FindUserTeamInContest(ctx, userID, contestID)
	if err == nil && team.ID > 0 {
		return &team.ID, nil
	}
	if err != nil && !errors.Is(err, contestports.ErrContestUserTeamNotFound) {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	return nil, contestcontracts.ErrNotRegistered
}
