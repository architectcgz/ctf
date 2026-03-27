package commands

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *TeamService) JoinTeam(ctx context.Context, contestID, userID, teamID int64) (*dto.TeamResp, error) {
	if err := s.ensureTeamJoinableContest(ctx, contestID); err != nil {
		return nil, err
	}
	if err := s.ensureApprovedRegistration(contestID, userID); err != nil {
		return nil, err
	}

	team, err := s.teamRepo.FindByID(teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrTeamNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if team.ContestID != contestID {
		return nil, errcode.ErrTeamNotFound
	}

	existingTeam, err := s.teamRepo.FindUserTeamInContest(userID, team.ContestID)
	if err == nil && existingTeam.ID > 0 {
		return nil, errcode.ErrAlreadyInTeam
	}

	err = s.teamRepo.AddMemberWithLock(contestID, team.ID, userID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrTeamFull) {
			return nil, errcode.ErrTeamFull
		}
		if errors.Is(err, contestdomain.ErrAlreadyJoinedContest) {
			return nil, errcode.ErrAlreadyInTeam
		}
		if s.teamRepo.IsUniqueViolation(err, "uk_team_members_contest_user") {
			return nil, errcode.ErrAlreadyInTeam
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotRegistered
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	count, _ := s.teamRepo.GetMemberCount(team.ID)
	return contestdomain.TeamRespFromModel(team, int(count)), nil
}
