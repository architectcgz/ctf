package contracts

import (
	"gorm.io/gorm"

	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

func NewPortReservationOwner(db *gorm.DB) runtimeports.PortReservationOwner {
	return runtimeinfra.NewRepository(db)
}
