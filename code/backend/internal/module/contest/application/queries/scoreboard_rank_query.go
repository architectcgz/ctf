package queries

import (
	"context"

	"ctf-platform/pkg/errcode"
)

func (s *ScoreboardService) GetTeamRank(ctx context.Context, contestID, teamID int64) (*TeamRankResult, error) {
	rank, ok, err := s.stateStore.GetLiveTeamRank(ctx, contestID, teamID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if !ok {
		return &TeamRankResult{TeamID: teamID, Rank: 0, Score: 0}, nil
	}

	return &TeamRankResult{
		TeamID: teamID,
		Rank:   rank.Rank,
		Score:  rank.Score,
	}, nil
}
