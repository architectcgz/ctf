package ports_test

import (
	"context"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyEnvironmentTemplateRepository struct{}

func (ctxOnlyEnvironmentTemplateRepository) Create(context.Context, *model.EnvironmentTemplate) error {
	return nil
}

func (ctxOnlyEnvironmentTemplateRepository) Update(context.Context, *model.EnvironmentTemplate) error {
	return nil
}

func (ctxOnlyEnvironmentTemplateRepository) DeleteWithContext(context.Context, int64) error {
	return nil
}

func (ctxOnlyEnvironmentTemplateRepository) FindByID(context.Context, int64) (*model.EnvironmentTemplate, error) {
	return nil, nil
}

func (ctxOnlyEnvironmentTemplateRepository) List(context.Context, string) ([]*model.EnvironmentTemplate, error) {
	return nil, nil
}

func (ctxOnlyEnvironmentTemplateRepository) IncrementUsage(context.Context, int64) error {
	return nil
}

var _ challengeports.EnvironmentTemplateRepository = (*ctxOnlyEnvironmentTemplateRepository)(nil)
