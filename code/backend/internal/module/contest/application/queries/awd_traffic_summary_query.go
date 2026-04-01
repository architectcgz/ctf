package queries

import (
	"context"

	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) GetTrafficSummary(ctx context.Context, contestID, roundID int64) (*dto.AWDTrafficSummaryResp, error) {
	round, err := s.ensureAWDRound(ctx, contestID, roundID)
	if err != nil {
		return nil, err
	}
	records, err := s.repo.ListTrafficEvents(ctx, contestID, roundID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return buildAWDTrafficSummary(contestdomain.AWDRoundRespFromModel(round), buildAWDTrafficEvents(records)), nil
}
