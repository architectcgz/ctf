package ports_test

import (
	"context"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyAWDServiceTemplateCommandRepository struct{}

func (ctxOnlyAWDServiceTemplateCommandRepository) CreateAWDServiceTemplate(context.Context, *model.AWDServiceTemplate) error {
	return nil
}

func (ctxOnlyAWDServiceTemplateCommandRepository) FindAWDServiceTemplateByID(context.Context, int64) (*model.AWDServiceTemplate, error) {
	return nil, nil
}

func (ctxOnlyAWDServiceTemplateCommandRepository) UpdateAWDServiceTemplate(context.Context, *model.AWDServiceTemplate) error {
	return nil
}

func (ctxOnlyAWDServiceTemplateCommandRepository) DeleteAWDServiceTemplate(context.Context, int64) error {
	return nil
}

var _ challengeports.AWDServiceTemplateCommandRepository = (*ctxOnlyAWDServiceTemplateCommandRepository)(nil)
