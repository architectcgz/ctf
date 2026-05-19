package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"

	"ctf-platform/internal/apperror"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
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
			return nil, contestcontracts.ErrInviteCodeGenerationFailed.WithCause(err)
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
		return nil, apperror.ErrInternal.WithCause(err)
	}

	return nil, contestcontracts.ErrInviteCodeGenerationFailed
}

func mapCreateTeamError(err error, s *TeamService) (retry bool, mapped error) {
	if s.teamRepo.IsUniqueViolation(err, "uk_teams_invite_code") {
		return true, nil
	}
	if s.teamRepo.IsUniqueViolation(err, "uk_teams_contest_name") {
		return false, contestcontracts.ErrTeamNameExists
	}
	if s.teamRepo.IsUniqueViolation(err, "uk_team_members_contest_user") {
		return false, contestcontracts.ErrAlreadyInTeam
	}
	if errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
		return false, contestcontracts.ErrNotRegistered
	}
	if isUniqueConflict(err) {
		return true, nil
	}
	return false, nil
}
