package ports_test

import (
	"context"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyEnvironmentTemplateRepository struct{}

func (ctxOnlyEnvironmentTemplateRepository) CreateWithContext(context.Context, *model.EnvironmentTemplate) error {
	return nil
}

func (ctxOnlyEnvironmentTemplateRepository) UpdateWithContext(context.Context, *model.EnvironmentTemplate) error {
	return nil
}

func (ctxOnlyEnvironmentTemplateRepository) DeleteWithContext(context.Context, int64) error {
	return nil
}

func (ctxOnlyEnvironmentTemplateRepository) FindByIDWithContext(context.Context, int64) (*model.EnvironmentTemplate, error) {
	return nil, nil
}

func (ctxOnlyEnvironmentTemplateRepository) ListWithContext(context.Context, string) ([]*model.EnvironmentTemplate, error) {
	return nil, nil
}

func (ctxOnlyEnvironmentTemplateRepository) IncrementUsageWithContext(context.Context, int64) error {
	return nil
}

var _ challengeports.EnvironmentTemplateRepository = (*ctxOnlyEnvironmentTemplateRepository)(nil)
