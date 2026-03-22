package application

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/errcode"
)

func (s *QueryService) GetTimeline(ctx context.Context, userID int64, limit, offset int) (*dto.TimelineResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	events, err := s.repo.GetUserTimeline(ctx, userID, limit, offset)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.TimelineResp{Events: events}, nil
}
