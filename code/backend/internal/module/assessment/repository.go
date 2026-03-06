package assessment

import (
	"ctf-platform/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Upsert 插入或更新能力画像
func (r *Repository) Upsert(profile *model.SkillProfile) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "dimension"}},
		DoUpdates: clause.AssignmentColumns([]string{"score", "updated_at"}),
	}).Create(profile).Error
}

// FindByUserID 查询用户所有维度画像
func (r *Repository) FindByUserID(userID int64) ([]*model.SkillProfile, error) {
	var profiles []*model.SkillProfile
	err := r.db.Where("user_id = ?", userID).Find(&profiles).Error
	return profiles, err
}

// BatchUpsert 批量插入或更新
func (r *Repository) BatchUpsert(profiles []*model.SkillProfile) error {
	if len(profiles) == 0 {
		return nil
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "dimension"}},
		DoUpdates: clause.AssignmentColumns([]string{"score", "updated_at"}),
	}).Create(profiles).Error
}
