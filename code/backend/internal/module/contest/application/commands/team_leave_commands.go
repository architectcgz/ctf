package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"

	"ctf-platform/internal/apperror"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *TeamService) LeaveTeam(ctx context.Context, contestID, userID, teamID int64) error {
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
	if team.CaptainID == userID {
		return contestcontracts.ErrCaptainCannotLeave
	}

	members, err := s.teamRepo.GetMembers(ctx, teamID)
	if err != nil {
		return err
	}
	if !teamHasMember(members, userID) {
		return contestcontracts.ErrNotInTeam
	}

	return s.teamRepo.RemoveMember(ctx, teamID, userID)
}
