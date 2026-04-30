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

func (ctxOnlyInstanceRepository) UpdateRuntime(context.Context, *model.Instance) error {
	return nil
}

func (ctxOnlyInstanceRepository) RefreshInstanceExpiry(context.Context, int64, time.Time) error {
	return nil
}

func (ctxOnlyInstanceRepository) UpdateStatusAndReleasePort(context.Context, int64, string) error {
	return nil
}

func (ctxOnlyInstanceRepository) FindByUserAndChallenge(context.Context, int64, int64) (*model.Instance, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) ListPendingInstances(context.Context, int) ([]*model.Instance, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) TryTransitionStatus(context.Context, int64, string, string) (bool, error) {
	return false, nil
}

func (ctxOnlyInstanceRepository) CountInstancesByStatus(context.Context, []string) (int64, error) {
	return 0, nil
}

var _ practiceports.InstanceRepository = (*ctxOnlyInstanceRepository)(nil)

type ctxOnlyPracticeCommandTxRepository struct{}

func (ctxOnlyPracticeCommandTxRepository) LockInstanceScope(context.Context, int64, int64, practiceports.InstanceScope) error {
	return nil
}

func (ctxOnlyPracticeCommandTxRepository) FindScopedExistingInstance(context.Context, int64, int64, practiceports.InstanceScope) (*model.Instance, error) {
	return nil, nil
}

func (ctxOnlyPracticeCommandTxRepository) FindScopedRestartableInstance(context.Context, int64, int64, practiceports.InstanceScope) (*model.Instance, error) {
	return nil, nil
}

func (ctxOnlyPracticeCommandTxRepository) CountScopedRunningInstances(context.Context, int64, practiceports.InstanceScope) (int, error) {
	return 0, nil
}

func (ctxOnlyPracticeCommandTxRepository) RefreshInstanceExpiry(context.Context, int64, time.Time) error {
	return nil
}

func (ctxOnlyPracticeCommandTxRepository) ResetInstanceRuntimeForRestart(context.Context, int64, string) error {
	return nil
}

func (ctxOnlyPracticeCommandTxRepository) CreateInstance(context.Context, *model.Instance) error {
	return nil
}

func (ctxOnlyPracticeCommandTxRepository) ReserveAvailablePort(context.Context, int, int) (int, error) {
	return 0, nil
}

func (ctxOnlyPracticeCommandTxRepository) BindReservedPort(context.Context, int, int64) error {
	return nil
}

var _ practiceports.PracticeCommandTxRepository = (*ctxOnlyPracticeCommandTxRepository)(nil)
