package queries

import (
	"context"

	"ctf-platform/internal/dto"
)

type Service interface {
	GetProgress(ctx context.Context, userID int64) (*dto.ProgressResp, error)
	GetTimeline(ctx context.Context, userID int64, limit, offset int) (*dto.TimelineResp, error)
}
