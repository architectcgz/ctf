package composition

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	containerruntimeinfra "ctf-platform/internal/module/container_runtime/infrastructure"
	instanceinfra "ctf-platform/internal/module/instance/infrastructure"
)

func withInstanceRuntimeLifecycleTx(
	ctx context.Context,
	db *gorm.DB,
	instanceRepo *instanceinfra.Repository,
	allocationRepo *containerruntimeinfra.AllocationRepository,
	fn func(instanceTx *instanceinfra.Repository, allocationTx *containerruntimeinfra.AllocationRepository) error,
) error {
	if db == nil || instanceRepo == nil || allocationRepo == nil || fn == nil {
		return fmt.Errorf("instance runtime lifecycle transaction is not configured")
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(instanceRepo.WithDB(tx), allocationRepo.WithDB(tx))
	})
}
