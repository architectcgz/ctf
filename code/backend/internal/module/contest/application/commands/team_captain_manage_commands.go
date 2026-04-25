package commands

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/pkg/errcode"
)

func (s *TeamService) DismissTeam(ctx context.Context, contestID, captainID, teamID int64) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrTeamNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if team.ContestID != contestID {
		return errcode.ErrTeamNotFound
	}
	if team.CaptainID != captainID {
		return errcode.ErrNotCaptain
	}
	return s.teamRepo.DeleteWithMembers(ctx, teamID)
}

func (s *TeamService) KickMember(ctx context.Context, contestID, captainID, teamID, memberUserID int64) error {
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrTeamNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}
	if team.ContestID != contestID {
		return errcode.ErrTeamNotFound
	}
	if team.CaptainID != captainID {
		return errcode.ErrNotCaptain
	}
	if memberUserID == captainID {
		return errcode.ErrCaptainCannotLeave
	}

	members, err := s.teamRepo.GetMembers(ctx, teamID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if !teamHasMember(members, memberUserID) {
		return errcode.ErrNotInTeam
	}
	if err := s.teamRepo.RemoveMember(ctx, teamID, memberUserID); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}
