package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"

	"ctf-platform/internal/apperror"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *TeamService) JoinTeam(ctx context.Context, contestID, userID, teamID int64) (*TeamResp, error) {
	if err := s.ensureTeamJoinableContest(ctx, contestID); err != nil {
		return nil, err
	}
	if err := s.ensureApprovedRegistration(ctx, contestID, userID); err != nil {
		return nil, err
	}

	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestTeamNotFound) {
			return nil, contestcontracts.ErrTeamNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if team.ContestID != contestID {
		return nil, contestcontracts.ErrTeamNotFound
	}

	existingTeam, err := s.teamRepo.FindUserTeamInContest(ctx, userID, team.ContestID)
	if err != nil {
		if !errors.Is(err, contestports.ErrContestUserTeamNotFound) {
			return nil, apperror.ErrInternal.WithCause(err)
		}
	} else if existingTeam.ID > 0 {
		return nil, contestcontracts.ErrAlreadyInTeam
	}

	err = s.teamRepo.AddMemberWithLock(ctx, contestID, team.ID, userID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrTeamFull) {
			return nil, contestcontracts.ErrTeamFull
		}
		if errors.Is(err, contestdomain.ErrAlreadyJoinedContest) {
			return nil, contestcontracts.ErrAlreadyInTeam
		}
		if s.teamRepo.IsUniqueViolation(err, "uk_team_members_contest_user") {
			return nil, contestcontracts.ErrAlreadyInTeam
		}
		if errors.Is(err, contestports.ErrContestParticipationRegistrationNotFound) {
			return nil, contestcontracts.ErrNotRegistered
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}

	count, _ := s.teamRepo.GetMemberCount(ctx, team.ID)
	return teamRespFromModel(team, int(count)), nil
}
