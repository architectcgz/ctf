package queries

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/module/challenge/domain"
	challengeports "ctf-platform/internal/module/challenge/ports"
	"ctf-platform/pkg/errcode"
)

type AWDServiceTemplateQueryService struct {
	repo challengeports.AWDServiceTemplateQueryRepository
}

func NewAWDServiceTemplateQueryService(repo challengeports.AWDServiceTemplateQueryRepository) *AWDServiceTemplateQueryService {
	return &AWDServiceTemplateQueryService{repo: repo}
}

func (s *AWDServiceTemplateQueryService) GetTemplateWithContext(ctx context.Context, id int64) (*dto.AWDServiceTemplateResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	item, err := s.repo.FindAWDServiceTemplateByIDWithContext(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return domain.AWDServiceTemplateRespFromModel(item), nil
}

func (s *AWDServiceTemplateQueryService) ListTemplatesWithContext(ctx context.Context, req *dto.AWDServiceTemplateQuery) (*dto.AWDServiceTemplatePageResp, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	items, total, err := s.repo.ListAWDServiceTemplatesWithContext(ctx, req)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	page := 1
	size := 20
	if req != nil {
		if req.Page > 0 {
			page = req.Page
		}
		if req.Size > 0 {
			size = req.Size
		}
	}
	resp := &dto.AWDServiceTemplatePageResp{
		Items: make([]*dto.AWDServiceTemplateResp, 0, len(items)),
		Total: total,
		Page:  page,
		Size:  size,
	}
	for _, item := range items {
		resp.Items = append(resp.Items, domain.AWDServiceTemplateRespFromModel(item))
	}
	return resp, nil
}
