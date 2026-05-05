package queries

import (
	"context"

	"ctf-platform/pkg/errcode"
)

func (s *AWDService) GetTrafficSummary(ctx context.Context, contestID, roundID int64) (*AWDTrafficSummaryResult, error) {
	round, err := s.ensureAWDRound(ctx, contestID, roundID)
	if err != nil {
		return nil, err
	}
	records, err := s.repo.ListTrafficEvents(ctx, contestID, roundID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return buildAWDTrafficSummary(&AWDRoundResult{
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
	}, buildAWDTrafficEvents(records)), nil
}
