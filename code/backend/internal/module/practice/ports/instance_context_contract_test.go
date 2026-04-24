package ports_test

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type ctxOnlyInstanceRepository struct{}

func (ctxOnlyInstanceRepository) FindByID(context.Context, int64) (*model.Instance, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) UpdateRuntimeWithContext(context.Context, *model.Instance) error {
	return nil
}

func (ctxOnlyInstanceRepository) RefreshInstanceExpiryWithContext(context.Context, int64, time.Time) error {
	return nil
}

func (ctxOnlyInstanceRepository) UpdateStatusAndReleasePort(context.Context, int64, string) error {
	return nil
}

func (ctxOnlyInstanceRepository) FindByUserAndChallengeWithContext(context.Context, int64, int64) (*model.Instance, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) ListPendingInstancesWithContext(context.Context, int) ([]*model.Instance, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) TryTransitionStatusWithContext(context.Context, int64, string, string) (bool, error) {
	return false, nil
}

func (ctxOnlyInstanceRepository) CountInstancesByStatusWithContext(context.Context, []string) (int64, error) {
	return 0, nil
}

var _ practiceports.InstanceRepository = (*ctxOnlyInstanceRepository)(nil)
