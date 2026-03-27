package queries

import (
	"context"

	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) ListRounds(ctx context.Context, contestID int64) ([]*dto.AWDRoundResp, error) {
	if _, err := s.ensureAWDContest(ctx, contestID); err != nil {
		return nil, err
	}

	rounds, err := s.repo.ListRoundsByContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := make([]*dto.AWDRoundResp, 0, len(rounds))
	for _, round := range rounds {
		roundCopy := round
		resp = append(resp, contestdomain.AWDRoundRespFromModel(&roundCopy))
	}
	return resp, nil
}
