package ports_test

import (
	"context"
	"time"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type ctxOnlyInstanceRepository struct{}

func (ctxOnlyInstanceRepository) FindByID(context.Context, int64) (*instanceentity.Instance, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) FindUserByID(context.Context, int64) (*identitycontracts.User, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) FindAccessibleByIDForUser(context.Context, int64, int64) (*instanceentity.Instance, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) ListVisibleByUser(context.Context, int64) ([]runtimeports.UserVisibleInstanceRow, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) ListTeacherInstances(context.Context, runtimeports.TeacherInstanceFilter) ([]runtimeports.TeacherInstanceRow, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) AtomicExtendByID(context.Context, int64, int, time.Duration) error {
	return nil
}

func (ctxOnlyInstanceRepository) UpdateStatusAndReleasePort(context.Context, int64, string) error {
	return nil
}

type ctxOnlyProxyTicketInstanceReader struct{}

func (ctxOnlyProxyTicketInstanceReader) FindByID(context.Context, int64) (*instanceentity.Instance, error) {
	return nil, nil
}

func (ctxOnlyProxyTicketInstanceReader) FindAWDTargetProxyScope(context.Context, int64, int64, int64, int64) (*runtimeports.AWDTargetProxyScope, error) {
	return nil, nil
}

func (ctxOnlyProxyTicketInstanceReader) FindAWDDefenseSSHScope(context.Context, int64, int64, int64) (*runtimeports.AWDDefenseSSHScope, error) {
	return nil, nil
}

var _ runtimeports.InstanceLookupRepository = (*ctxOnlyInstanceRepository)(nil)
var _ runtimeports.InstanceUserLookupRepository = (*ctxOnlyInstanceRepository)(nil)
var _ runtimeports.InstanceAccessRepository = (*ctxOnlyInstanceRepository)(nil)
var _ runtimeports.UserVisibleInstanceRepository = (*ctxOnlyInstanceRepository)(nil)
var _ runtimeports.TeacherInstanceQueryRepository = (*ctxOnlyInstanceRepository)(nil)
var _ runtimeports.InstanceExtendRepository = (*ctxOnlyInstanceRepository)(nil)
var _ runtimeports.InstanceStatusRepository = (*ctxOnlyInstanceRepository)(nil)
var _ runtimeports.ProxyTicketInstanceReader = (*ctxOnlyProxyTicketInstanceReader)(nil)
