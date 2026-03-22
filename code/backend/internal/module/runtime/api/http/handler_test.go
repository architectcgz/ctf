package http

import (
	"context"
	"testing"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	runtimeapp "ctf-platform/internal/module/runtime/application"
)

type stubRuntimeService struct{}

func (stubRuntimeService) DestroyInstanceWithContext(context.Context, int64, int64) error {
	return nil
}

func (stubRuntimeService) ExtendInstanceWithContext(context.Context, int64, int64) (*dto.InstanceResp, error) {
	return nil, nil
}

func (stubRuntimeService) GetAccessURLWithContext(context.Context, int64, int64) (string, error) {
	return "", nil
}

func (stubRuntimeService) GetUserInstancesWithContext(context.Context, int64) ([]*dto.InstanceInfo, error) {
	return nil, nil
}

func (stubRuntimeService) ListTeacherInstances(context.Context, int64, string, *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error) {
	return nil, nil
}

func (stubRuntimeService) DestroyTeacherInstance(context.Context, int64, int64, string) error {
	return nil
}

func (stubRuntimeService) IssueProxyTicket(context.Context, authctx.CurrentUser, int64) (string, error) {
	return "", nil
}

func (stubRuntimeService) ResolveProxyTicket(context.Context, string) (*runtimeapp.ProxyTicketClaims, error) {
	return nil, nil
}

func (stubRuntimeService) ProxyTicketMaxAge() int {
	return 0
}

func (stubRuntimeService) ProxyBodyPreviewSize() int {
	return 0
}

func TestHandlerContractsCompile(t *testing.T) {
	var _ runtimeService = stubRuntimeService{}
	_ = NewHandler(stubRuntimeService{}, nil, CookieConfig{})
}
