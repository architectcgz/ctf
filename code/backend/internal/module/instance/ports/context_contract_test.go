package ports_test

import (
	"context"
	"time"

	instanceentity "ctf-platform/internal/module/instance/entity"
	instanceports "ctf-platform/internal/module/instance/ports"
)

type ctxOnlyInstanceRepository struct{}

func (ctxOnlyInstanceRepository) FindByID(context.Context, int64) (*instanceentity.Instance, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) FindUserByID(context.Context, int64) (*instanceports.InstanceUser, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) FindAccessibleByIDForUser(context.Context, int64, int64) (*instanceentity.Instance, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) ListVisibleByUser(context.Context, int64) ([]instanceports.UserVisibleInstanceRow, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) ListTeacherInstances(context.Context, instanceports.TeacherInstanceFilter) (*instanceports.TeacherInstancePage, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) AtomicExtendByID(context.Context, int64, int, time.Duration) error {
	return nil
}

func (ctxOnlyInstanceRepository) MarkStopping(context.Context, int64) (bool, error) {
	return true, nil
}

func (ctxOnlyInstanceRepository) FinalizeStoppedRuntime(context.Context, int64) error {
	return nil
}

func (ctxOnlyInstanceRepository) UpdateStatusAndReleasePort(context.Context, int64, string) error {
	return nil
}

type ctxOnlyProxyTicketInstanceReader struct{}

func (ctxOnlyProxyTicketInstanceReader) FindByID(context.Context, int64) (*instanceentity.Instance, error) {
	return nil, nil
}

func (ctxOnlyProxyTicketInstanceReader) FindAWDTargetProxyScope(context.Context, int64, int64, int64, int64) (*instanceports.AWDTargetProxyScope, error) {
	return nil, nil
}

func (ctxOnlyProxyTicketInstanceReader) FindAWDDefenseSSHScope(context.Context, int64, int64, int64) (*instanceports.AWDDefenseSSHScope, error) {
	return nil, nil
}

var _ instanceports.InstanceLookupRepository = (*ctxOnlyInstanceRepository)(nil)
var _ instanceports.InstanceUserLookupRepository = (*ctxOnlyInstanceRepository)(nil)
var _ instanceports.InstanceAccessRepository = (*ctxOnlyInstanceRepository)(nil)
var _ instanceports.UserVisibleInstanceRepository = (*ctxOnlyInstanceRepository)(nil)
var _ instanceports.TeacherInstanceQueryRepository = (*ctxOnlyInstanceRepository)(nil)
var _ instanceports.InstanceExtendRepository = (*ctxOnlyInstanceRepository)(nil)
var _ instanceports.InstanceStatusRepository = (*ctxOnlyInstanceRepository)(nil)
var _ instanceports.ProxyTicketInstanceReader = (*ctxOnlyProxyTicketInstanceReader)(nil)
