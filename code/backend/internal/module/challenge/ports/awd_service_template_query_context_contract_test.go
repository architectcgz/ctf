package ports_test

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyAWDServiceTemplateQueryRepository struct{}

func (ctxOnlyAWDServiceTemplateQueryRepository) FindAWDServiceTemplateByIDWithContext(context.Context, int64) (*model.AWDServiceTemplate, error) {
	return nil, nil
}

func (ctxOnlyAWDServiceTemplateQueryRepository) ListAWDServiceTemplatesWithContext(context.Context, *dto.AWDServiceTemplateQuery) ([]*model.AWDServiceTemplate, int64, error) {
	return nil, 0, nil
}

var _ challengeports.AWDServiceTemplateQueryRepository = (*ctxOnlyAWDServiceTemplateQueryRepository)(nil)
