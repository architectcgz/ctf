package commands

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/pkg/errcode"
)

func (s *TeamService) LeaveTeam(_ context.Context, contestID, userID, teamID int64) error {
	team, err := s.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrTeamNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if team.ContestID != contestID {
		return errcode.ErrTeamNotFound
	}
	if team.CaptainID == userID {
		return errcode.ErrCaptainCannotLeave
	}

	members, err := s.teamRepo.GetMembers(teamID)
	if err != nil {
		return err
	}
	if !teamHasMember(members, userID) {
		return errcode.ErrNotInTeam
	}

	return s.teamRepo.RemoveMember(teamID, userID)
}
