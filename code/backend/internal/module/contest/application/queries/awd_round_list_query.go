package queries

import (
	"context"

	"ctf-platform/pkg/errcode"
)

func (s *AWDService) ListRounds(ctx context.Context, contestID int64) ([]AWDRoundResult, error) {
	if _, err := s.ensureAWDContest(ctx, contestID); err != nil {
		return nil, err
	}

	rounds, err := s.repo.ListRoundsByContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := make([]AWDRoundResult, 0, len(rounds))
	for _, round := range rounds {
		resp = append(resp, AWDRoundResult{
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
		})
	}
	return resp, nil
}
