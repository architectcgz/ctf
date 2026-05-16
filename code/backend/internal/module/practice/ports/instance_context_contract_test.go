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

func (ctxOnlyInstanceRepository) FinishActiveAWDServiceOperationForInstance(context.Context, int64, string, string, time.Time) error {
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

var _ practiceports.PracticeInstanceLookupRepository = (*ctxOnlyInstanceRepository)(nil)
var _ practiceports.PracticeInstanceRuntimeWriteRepository = (*ctxOnlyInstanceRepository)(nil)
var _ practiceports.PracticeInstanceAWDOperationRepository = (*ctxOnlyInstanceRepository)(nil)
var _ practiceports.PracticeInstanceStatusRepository = (*ctxOnlyInstanceRepository)(nil)
var _ practiceports.PracticePendingInstanceRepository = (*ctxOnlyInstanceRepository)(nil)
var _ practiceports.PracticeInstanceStatsRepository = (*ctxOnlyInstanceRepository)(nil)

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

func (ctxOnlyPracticeCommandTxRepository) ResetInstanceRuntimeForRestart(context.Context, int64, string, time.Time, bool) error {
	return nil
}

func (ctxOnlyPracticeCommandTxRepository) CreateInstance(context.Context, *model.Instance) error {
	return nil
}

func (ctxOnlyPracticeCommandTxRepository) CreateAWDServiceOperation(context.Context, *model.AWDServiceOperation) error {
	return nil
}

func (ctxOnlyPracticeCommandTxRepository) FinishAWDServiceOperation(context.Context, int64, string, string, time.Time) error {
	return nil
}

func (ctxOnlyPracticeCommandTxRepository) ReserveAvailablePort(context.Context, int, int) (int, error) {
	return 0, nil
}

func (ctxOnlyPracticeCommandTxRepository) ReserveAvailablePortExcluding(context.Context, int, int, int) (int, error) {
	return 0, nil
}

func (ctxOnlyPracticeCommandTxRepository) BindReservedPort(context.Context, int, int64) error {
	return nil
}

func (ctxOnlyPracticeCommandTxRepository) ReleaseReservedPort(context.Context, int) error {
	return nil
}

func (ctxOnlyPracticeCommandTxRepository) ReleasePortForInstance(context.Context, int, int64) error {
	return nil
}

var _ practiceports.PracticeInstanceScopeLockRepository = (*ctxOnlyPracticeCommandTxRepository)(nil)
var _ practiceports.PracticeScopedExistingInstanceRepository = (*ctxOnlyPracticeCommandTxRepository)(nil)
var _ practiceports.PracticeScopedRestartableInstanceRepository = (*ctxOnlyPracticeCommandTxRepository)(nil)
var _ practiceports.PracticeScopedRunningCountRepository = (*ctxOnlyPracticeCommandTxRepository)(nil)
var _ practiceports.PracticeInstanceExpiryRepository = (*ctxOnlyPracticeCommandTxRepository)(nil)
var _ practiceports.PracticeInstanceRestartRepository = (*ctxOnlyPracticeCommandTxRepository)(nil)
var _ practiceports.PracticeInstanceCreateRepository = (*ctxOnlyPracticeCommandTxRepository)(nil)
var _ practiceports.PracticeAWDServiceOperationCreateRepository = (*ctxOnlyPracticeCommandTxRepository)(nil)
var _ practiceports.PracticeAWDServiceOperationFinishRepository = (*ctxOnlyPracticeCommandTxRepository)(nil)
var _ practiceports.PracticePortReservationRepository = (*ctxOnlyPracticeCommandTxRepository)(nil)
var _ practiceports.PracticeInstanceStartTxRepository = (*ctxOnlyPracticeCommandTxRepository)(nil)
var _ practiceports.PracticeInstanceRestartTxRepository = (*ctxOnlyPracticeCommandTxRepository)(nil)
var _ practiceports.PracticeAWDServiceOperationTxRepository = (*ctxOnlyPracticeCommandTxRepository)(nil)
