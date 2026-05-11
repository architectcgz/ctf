package commands

import (
	"context"
	"fmt"

	"ctf-platform/internal/dto"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	"ctf-platform/pkg/errcode"
)

// InstanceService 保留 runtime compat import path，并把实例命令委托给 instance owner。
type InstanceService struct {
	service instancecontracts.InstanceCommandService
}

func NewInstanceService(service instancecontracts.InstanceCommandService) *InstanceService {
	return &InstanceService{service: service}
}

func (s *InstanceService) DestroyInstance(ctx context.Context, instanceID, userID int64) error {
	if s == nil || s.service == nil {
		return errRuntimeCompatInstanceServiceUnavailable()
	}
	return s.service.DestroyInstance(ctx, instanceID, userID)
}

func (s *InstanceService) ExtendInstance(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error) {
	if s == nil || s.service == nil {
		return nil, errRuntimeCompatInstanceServiceUnavailable()
	}
	return s.service.ExtendInstance(ctx, instanceID, userID)
}

func (s *InstanceService) DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error {
	if s == nil || s.service == nil {
		return errRuntimeCompatInstanceServiceUnavailable()
	}
	return s.service.DestroyTeacherInstance(ctx, instanceID, requesterID, requesterRole)
}

func errRuntimeCompatInstanceServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("runtime instance compat service is not configured"))
}
