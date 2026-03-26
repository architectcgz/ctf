package queries

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *TeamService) ListTeams(ctx context.Context, contestID int64) ([]*dto.TeamResp, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teams, err := s.teamRepo.ListByContest(contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teamIDs := make([]int64, len(teams))
	for i, team := range teams {
		teamIDs[i] = team.ID
	}

	countMap, err := s.teamRepo.GetMemberCountBatch(teamIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*dto.TeamResp, 0, len(teams))
	for _, team := range teams {
		result = append(result, contestdomain.TeamRespFromModel(team, countMap[team.ID]))
	}
	return result, nil
}

func (s *TeamService) GetMyTeam(ctx context.Context, contestID, userID int64) (map[string]any, error) {
	team, err := s.teamRepo.FindUserTeamInContest(userID, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teamResp, members, err := s.GetTeamInfo(team.ID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"id":              teamResp.ID,
		"name":            teamResp.Name,
		"invite_code":     teamResp.InviteCode,
		"captain_user_id": teamResp.CaptainID,
		"members":         members,
	}, nil
}
