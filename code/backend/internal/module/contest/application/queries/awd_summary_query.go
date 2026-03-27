package queries

import (
	"context"

	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) GetRoundSummary(ctx context.Context, contestID, roundID int64) (*dto.AWDRoundSummaryResp, error) {
	round, err := s.ensureAWDRound(ctx, contestID, roundID)
	if err != nil {
		return nil, err
	}

	teams, err := s.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}

	services, err := s.repo.ListServicesByRound(ctx, roundID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	attackLogs, err := s.repo.ListAttackLogsByRound(ctx, roundID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	metrics, respItems := buildAWDRoundSummary(teams, services, attackLogs)

	return &dto.AWDRoundSummaryResp{
		Round:   contestdomain.AWDRoundRespFromModel(round),
		Metrics: metrics,
		Items:   respItems,
	}, nil
}
