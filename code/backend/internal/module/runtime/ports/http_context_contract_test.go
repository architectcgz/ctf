package ports_test

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type ctxOnlyInstanceRepository struct{}

func (ctxOnlyInstanceRepository) FindByIDWithContext(context.Context, int64) (*model.Instance, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) FindUserByID(context.Context, int64) (*model.User, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) FindAccessibleByIDForUser(context.Context, int64, int64) (*model.Instance, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) ListVisibleByUser(context.Context, int64) ([]runtimeports.UserVisibleInstanceRow, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) ListTeacherInstances(context.Context, runtimeports.TeacherInstanceFilter) ([]runtimeports.TeacherInstanceRow, error) {
	return nil, nil
}

func (ctxOnlyInstanceRepository) AtomicExtendByIDWithContext(context.Context, int64, int, time.Duration) error {
	return nil
}

func (ctxOnlyInstanceRepository) UpdateStatusAndReleasePortWithContext(context.Context, int64, string) error {
	return nil
}

type ctxOnlyProxyTicketInstanceReader struct{}

func (ctxOnlyProxyTicketInstanceReader) FindByIDWithContext(context.Context, int64) (*model.Instance, error) {
	return nil, nil
}

var _ runtimeports.InstanceRepository = (*ctxOnlyInstanceRepository)(nil)
var _ runtimeports.ProxyTicketInstanceReader = (*ctxOnlyProxyTicketInstanceReader)(nil)
