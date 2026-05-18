package commands

import (
	"context"
	"errors"

	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func resolveCreateTeamMaxMembers(req CreateTeamInput) int {
	maxMembers := req.MaxMembers
	if maxMembers == 0 {
		maxMembers = 4
	}
	return maxMembers
}

func (s *TeamService) createTeamWithInviteRetries(ctx context.Context, contestID, captainID int64, teamName string, maxMembers int) (*contestentity.Team, error) {
	const maxRetries = 3
	var team *contestentity.Team
	for i := 0; i < maxRetries; i++ {
		inviteCode, err := generateInviteCode()
		if err != nil {
			return nil, errcode.ErrInviteCodeGenerationFailed.WithCause(err)
		}

		team = &contestentity.Team{
			ContestID:  contestID,
			Name:       teamName,
			CaptainID:  captainID,
			InviteCode: inviteCode,
			MaxMembers: maxMembers,
		}
		err = s.teamRepo.CreateWithMember(ctx, team, captainID)
		if err == nil {
			return team, nil
		}
		next, mapped := mapCreateTeamError(err, s)
		if mapped != nil {
			return nil, mapped
		}
		if next {
			continue
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return nil, errcode.ErrInviteCodeGenerationFailed
}

func mapCreateTeamError(err error, s *TeamService) (retry bool, mapped error) {
	if s.teamRepo.IsUniqueViolation(err, "uk_teams_invite_code") {
		return true, nil
	}
	if s.teamRepo.IsUniqueViolation(err, "uk_teams_contest_name") {
		return false, errcode.ErrTeamNameExists
	}
	if s.teamRepo.IsUniqueViolation(err, "uk_team_members_contest_user") {
		return false, errcode.ErrAlreadyInTeam
	}
	if errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
		return false, errcode.ErrNotRegistered
	}
	if isUniqueConflict(err) {
		return true, nil
	}
	return false, nil
}
