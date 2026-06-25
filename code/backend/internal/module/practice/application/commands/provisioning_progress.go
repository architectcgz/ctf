package commands

import (
	"context"
	"fmt"

	"ctf-platform/internal/apperror"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

type provisioningProgressRepository interface {
	RecordProvisioningProgress(ctx context.Context, progress instancecontracts.ProvisioningProgress) (bool, error)
}

func (s *serviceCore) recordProvisioningProgress(ctx context.Context, progress instancecontracts.ProvisioningProgress) error {
	if s == nil || s.instanceRepo == nil {
		return nil
	}
	repo, ok := s.instanceRepo.(provisioningProgressRepository)
	if !ok {
		return nil
	}
	if progress.InstanceID <= 0 {
		return nil
	}
	if progress.Stage == "" {
		return apperror.ErrInternal.WithCause(fmt.Errorf("provisioning stage is empty"))
	}
	if _, err := repo.RecordProvisioningProgress(ctx, progress); err != nil {
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}
