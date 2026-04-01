package queries

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) ListTrafficEvents(ctx context.Context, contestID, roundID int64, req *dto.ListAWDTrafficEventsReq) (*dto.AWDTrafficEventPageResp, error) {
	if _, err := s.ensureAWDRound(ctx, contestID, roundID); err != nil {
		return nil, err
	}
	records, err := s.repo.ListTrafficEvents(ctx, contestID, roundID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	filtered := filterAWDTrafficEvents(buildAWDTrafficEvents(records), req)
	pageItems, total, page, size := paginateAWDTrafficEvents(filtered, req.Page, req.Size)
	return &dto.AWDTrafficEventPageResp{
		List:     pageItems,
		Total:    total,
		Page:     page,
		PageSize: size,
	}, nil
}
