package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"

	"ctf-platform/internal/apperror"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *TeamService) DismissTeam(ctx context.Context, contestID, captainID, teamID int64) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestTeamNotFound) {
			return contestcontracts.ErrTeamNotFound
		}
		return apperror.ErrInternal.WithCause(err)
	}
	if team.ContestID != contestID {
		return contestcontracts.ErrTeamNotFound
	}
	if team.CaptainID != captainID {
		return contestcontracts.ErrNotCaptain
	}
	return s.teamRepo.DeleteWithMembers(ctx, teamID)
}

func (s *TeamService) KickMember(ctx context.Context, contestID, captainID, teamID, memberUserID int64) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestTeamNotFound) {
			return contestcontracts.ErrTeamNotFound
		}
		return apperror.ErrInternal.WithCause(err)
	}
	if team.ContestID != contestID {
		return contestcontracts.ErrTeamNotFound
	}
	if team.CaptainID != captainID {
		return contestcontracts.ErrNotCaptain
	}
	if memberUserID == captainID {
		return contestcontracts.ErrCaptainCannotLeave
	}

	members, err := s.teamRepo.GetMembers(ctx, teamID)
	if err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	if !teamHasMember(members, memberUserID) {
		return contestcontracts.ErrNotInTeam
	}
	if err := s.teamRepo.RemoveMember(ctx, teamID, memberUserID); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}
