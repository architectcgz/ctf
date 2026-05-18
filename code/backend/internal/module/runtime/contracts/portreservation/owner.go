package portreservation

import (
	"gorm.io/gorm"

	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

func NewOwner(db *gorm.DB) runtimeports.PortReservationOwner {
	return runtimeinfra.NewRepository(db)
}
