package commands

import (
	"gorm.io/gorm"

	containerruntimeinfra "ctf-platform/internal/module/container_runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
)

func newPracticeRepositoryWithRuntimePortOwner(db *gorm.DB) *practiceinfra.Repository {
	return practiceinfra.NewRepositoryWithRuntimePortOwner(db, func(db *gorm.DB) runtimeports.PortReservationOwner {
		return containerruntimeinfra.NewAllocationRepository(db)
	})
}
