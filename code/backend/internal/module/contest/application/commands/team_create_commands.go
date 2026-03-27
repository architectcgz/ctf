package commands

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *TeamService) CreateTeam(ctx context.Context, contestID, captainID int64, req *dto.CreateTeamReq) (*dto.TeamResp, error) {
	if err := s.ensureTeamJoinableContest(ctx, contestID); err != nil {
		return nil, err
	}
	if err := s.ensureApprovedRegistration(contestID, captainID); err != nil {
		return nil, err
	}

	existingTeam, err := s.teamRepo.FindUserTeamInContest(captainID, contestID)
	if err == nil && existingTeam.ID > 0 {
		return nil, errcode.ErrAlreadyInTeam
	}

	maxMembers := req.MaxMembers
	if maxMembers == 0 {
		maxMembers = 4
	}

	const maxRetries = 3
	var team *model.Team
	for i := 0; i < maxRetries; i++ {
		inviteCode, err := generateInviteCode()
		if err != nil {
			return nil, errcode.ErrInviteCodeGenerationFailed.WithCause(err)
		}

		team = &model.Team{
			ContestID:  contestID,
			Name:       req.Name,
			CaptainID:  captainID,
			InviteCode: inviteCode,
			MaxMembers: maxMembers,
		}

		err = s.teamRepo.CreateWithMember(team, captainID)
		if err == nil {
			break
		}

		if s.teamRepo.IsUniqueViolation(err, "uk_teams_invite_code") {
			continue
		}
		if s.teamRepo.IsUniqueViolation(err, "uk_teams_contest_name") {
			return nil, errcode.ErrTeamNameExists
		}
		if s.teamRepo.IsUniqueViolation(err, "uk_team_members_contest_user") {
			return nil, errcode.ErrAlreadyInTeam
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotRegistered
		}
		if !isUniqueConflict(err) {
			return nil, errcode.ErrInternal.WithCause(err)
		}
	}

	if team == nil || team.ID == 0 {
		return nil, errcode.ErrInviteCodeGenerationFailed
	}

	return contestdomain.TeamRespFromModel(team, 1), nil
}
