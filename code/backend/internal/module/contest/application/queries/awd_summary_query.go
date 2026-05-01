package queries

import (
	"context"

	"ctf-platform/pkg/errcode"
)

func (s *AWDService) GetRoundSummary(ctx context.Context, contestID, roundID int64) (*AWDRoundSummaryResult, error) {
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

	return &AWDRoundSummaryResult{
		Round: &AWDRoundResult{
			ID:           round.ID,
			ContestID:    round.ContestID,
			RoundNumber:  round.RoundNumber,
			Status:       round.Status,
			StartedAt:    round.StartedAt,
			EndedAt:      round.EndedAt,
			AttackScore:  round.AttackScore,
			DefenseScore: round.DefenseScore,
			CreatedAt:    round.CreatedAt,
			UpdatedAt:    round.UpdatedAt,
		},
		Metrics: metrics,
		Items:   respItems,
	}, nil
}
