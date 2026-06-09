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
	runtimeRepo *runtimeinfra.Repository,
	fn func(instanceTx *instanceinfra.Repository, runtimeTx *runtimeinfra.Repository) error,
) error {
	if db == nil || instanceRepo == nil || runtimeRepo == nil || fn == nil {
		return fmt.Errorf("instance runtime lifecycle transaction is not configured")
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(instanceRepo.WithDB(tx), runtimeRepo.WithDB(tx))
	})
}
