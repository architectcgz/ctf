package queries

import (
	"context"
	"errors"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *TeamService) ListTeams(ctx context.Context, contestID int64) ([]*TeamResult, error) {
	if _, err := s.contestRepo.FindByID(ctx, contestID); err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teams, err := s.teamRepo.ListByContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teamIDs := make([]int64, len(teams))
	for i, team := range teams {
		teamIDs[i] = team.ID
	}

	countMap, err := s.teamRepo.GetMemberCountBatch(ctx, teamIDs)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	result := make([]*TeamResult, 0, len(teams))
	for _, team := range teams {
		result = append(result, teamResultFromModel(team, countMap[team.ID]))
	}
	return result, nil
}

func (s *TeamService) GetMyTeam(ctx context.Context, contestID, userID int64) (*MyTeamResult, error) {
	team, err := s.teamRepo.FindUserTeamInContest(ctx, userID, contestID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestUserTeamNotFound) {
			return nil, nil
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teamResp, members, err := s.GetTeamInfo(ctx, team.ID)
	if err != nil {
		return nil, err
	}

	return &MyTeamResult{
		ID:         teamResp.ID,
		Name:       teamResp.Name,
		InviteCode: teamResp.InviteCode,
		CaptainID:  teamResp.CaptainID,
		Members:    members,
	}, nil
}
