package queries

import (
	"context"
	"fmt"

	"ctf-platform/internal/dto"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	"ctf-platform/pkg/errcode"
)

// InstanceService 保留 runtime compat import path，并把实例查询委托给 instance owner。
type InstanceService struct {
	service instancecontracts.InstanceQueryService
}

func NewInstanceService(service instancecontracts.InstanceQueryService) *InstanceService {
	return &InstanceService{service: service}
}

func (s *InstanceService) GetAccessURL(ctx context.Context, instanceID, userID int64) (string, error) {
	if s == nil || s.service == nil {
		return "", errRuntimeCompatInstanceQueryServiceUnavailable()
	}
	return s.service.GetAccessURL(ctx, instanceID, userID)
}

func (s *InstanceService) GetUserInstances(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error) {
	if s == nil || s.service == nil {
		return nil, errRuntimeCompatInstanceQueryServiceUnavailable()
	}
	return s.service.GetUserInstances(ctx, userID)
}

func (s *InstanceService) ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error) {
	if s == nil || s.service == nil {
		return nil, errRuntimeCompatInstanceQueryServiceUnavailable()
	}
	return s.service.ListTeacherInstances(ctx, requesterID, requesterRole, query)
}

func errRuntimeCompatInstanceQueryServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("runtime instance query compat service is not configured"))
}
