package ports_test

import (
	"context"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyAWDServiceTemplateCommandRepository struct{}

func (ctxOnlyAWDServiceTemplateCommandRepository) CreateAWDServiceTemplateWithContext(context.Context, *model.AWDServiceTemplate) error {
	return nil
}

func (ctxOnlyAWDServiceTemplateCommandRepository) FindAWDServiceTemplateByIDWithContext(context.Context, int64) (*model.AWDServiceTemplate, error) {
	return nil, nil
}

func (ctxOnlyAWDServiceTemplateCommandRepository) UpdateAWDServiceTemplateWithContext(context.Context, *model.AWDServiceTemplate) error {
	return nil
}

func (ctxOnlyAWDServiceTemplateCommandRepository) DeleteAWDServiceTemplateWithContext(context.Context, int64) error {
	return nil
}

var _ challengeports.AWDServiceTemplateCommandRepository = (*ctxOnlyAWDServiceTemplateCommandRepository)(nil)
