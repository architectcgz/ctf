package teacher

import (
	"gorm.io/gorm"

	readmodelinfra "ctf-platform/internal/module/teaching_readmodel/infrastructure"
)

type Repository = readmodelinfra.Repository

func NewRepository(db *gorm.DB) *Repository {
	return readmodelinfra.NewRepository(db)
}
