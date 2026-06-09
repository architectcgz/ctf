package composition

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	instanceinfra "ctf-platform/internal/module/instance/infrastructure"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
)

func withInstanceRuntimeLifecycleTx(
	ctx context.Context,
	db *gorm.DB,
	instanceRepo *instanceinfra.Repository,
	allocationRepo *runtimeinfra.AllocationRepository,
	fn func(instanceTx *instanceinfra.Repository, allocationTx *runtimeinfra.AllocationRepository) error,
) error {
	if db == nil || instanceRepo == nil || allocationRepo == nil || fn == nil {
		return fmt.Errorf("instance runtime lifecycle transaction is not configured")
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(instanceRepo.WithDB(tx), allocationRepo.WithDB(tx))
	})
}
