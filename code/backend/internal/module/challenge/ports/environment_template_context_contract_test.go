package ports_test

import (
	"context"

	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyEnvironmentTemplateRepository struct{}

func (ctxOnlyEnvironmentTemplateRepository) Create(context.Context, *challengeentity.EnvironmentTemplate) error {
	return nil
}

func (ctxOnlyEnvironmentTemplateRepository) Update(context.Context, *challengeentity.EnvironmentTemplate) error {
	return nil
}

func (ctxOnlyEnvironmentTemplateRepository) Delete(context.Context, int64) error {
	return nil
}

func (ctxOnlyEnvironmentTemplateRepository) FindByID(context.Context, int64) (*challengeentity.EnvironmentTemplate, error) {
	return nil, nil
}

func (ctxOnlyEnvironmentTemplateRepository) List(context.Context, string) ([]*challengeentity.EnvironmentTemplate, error) {
	return nil, nil
}

func (ctxOnlyEnvironmentTemplateRepository) IncrementUsage(context.Context, int64) error {
	return nil
}

var _ challengeports.EnvironmentTemplateCommandRepository = (*ctxOnlyEnvironmentTemplateRepository)(nil)
var _ challengeports.EnvironmentTemplateQueryRepository = (*ctxOnlyEnvironmentTemplateRepository)(nil)
var _ challengeports.EnvironmentTemplateUsageRepository = (*ctxOnlyEnvironmentTemplateRepository)(nil)
