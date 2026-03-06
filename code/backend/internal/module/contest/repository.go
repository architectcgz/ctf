package contest

import (
	"ctf-platform/internal/model"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByID(id int64) (*model.Contest, error) {
	var contest model.Contest
	err := r.db.Where("id = ?", id).First(&contest).Error
	return &contest, err
}
