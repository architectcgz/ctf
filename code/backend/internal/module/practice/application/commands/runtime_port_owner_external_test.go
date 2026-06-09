package commands_test

import (
	"gorm.io/gorm"

	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

func newPracticeRepositoryWithRuntimePortOwner(db *gorm.DB) *practiceinfra.Repository {
	return practiceinfra.NewRepositoryWithRuntimePortOwner(db, func(db *gorm.DB) runtimeports.PortReservationOwner {
		return runtimeinfra.NewAllocationRepository(db)
	})
}
