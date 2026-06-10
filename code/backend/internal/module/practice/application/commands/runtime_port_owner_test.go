package commands

import (
	"gorm.io/gorm"

	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
)

func newPracticeRepositoryWithRuntimePortOwner(db *gorm.DB) *practiceinfra.Repository {
	return practiceinfra.NewRepositoryWithRuntimePortOwner(db, func(db *gorm.DB) runtimeports.PortReservationOwner {
		return runtimeinfra.NewAllocationRepository(db)
	})
}
