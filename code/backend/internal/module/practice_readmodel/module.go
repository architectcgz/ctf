package practice_readmodel

import (
	"context"

	"ctf-platform/internal/dto"
	readmodelapp "ctf-platform/internal/module/practice_readmodel/application"
)

type Module struct {
	service *readmodelapp.QueryService
}

func NewModule(service *readmodelapp.QueryService) *Module {
	return &Module{service: service}
}

func (m *Module) GetProgress(ctx context.Context, userID int64) (*dto.ProgressResp, error) {
	return m.service.GetProgress(ctx, userID)
}

func (m *Module) GetTimeline(ctx context.Context, userID int64, limit, offset int) (*dto.TimelineResp, error) {
	return m.service.GetTimeline(ctx, userID, limit, offset)
}
