package infrastructure

import (
	"time"

	"ctf-platform/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *Repository) FindWriteupByChallengeID(challengeID int64) (*model.ChallengeWriteup, error) {
	var writeup model.ChallengeWriteup
	err := r.db.Where("challenge_id = ?", challengeID).First(&writeup).Error
	if err != nil {
		return nil, err
	}
	return &writeup, nil
}

func (r *Repository) UpsertWriteup(writeup *model.ChallengeWriteup) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "challenge_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "content", "visibility", "release_at", "created_by", "updated_at"}),
	}).Create(writeup).Error
}

func (r *Repository) DeleteWriteupByChallengeID(challengeID int64) error {
	return r.db.Where("challenge_id = ?", challengeID).Delete(&model.ChallengeWriteup{}).Error
}

func (r *Repository) FindReleasedWriteupByChallengeID(challengeID int64, now time.Time) (*model.ChallengeWriteup, error) {
	var writeup model.ChallengeWriteup
	err := r.db.
		Where("challenge_id = ?", challengeID).
		Where("visibility = ? OR (visibility = ? AND release_at IS NOT NULL AND release_at <= ?)",
			model.WriteupVisibilityPublic,
			model.WriteupVisibilityScheduled,
			now,
		).
		First(&writeup).Error
	if err != nil {
		return nil, err
	}
	return &writeup, nil
}

func (r *Repository) FindChallengeTopologyByChallengeID(challengeID int64) (*model.ChallengeTopology, error) {
	var topology model.ChallengeTopology
	err := r.db.Where("challenge_id = ?", challengeID).First(&topology).Error
	if err != nil {
		return nil, err
	}
	return &topology, nil
}

func (r *Repository) UpsertChallengeTopology(topology *model.ChallengeTopology) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "challenge_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"template_id", "entry_node_key", "spec", "updated_at", "deleted_at"}),
	}).Create(topology).Error
}

func (r *Repository) DeleteChallengeTopologyByChallengeID(challengeID int64) error {
	return r.db.Where("challenge_id = ?", challengeID).Delete(&model.ChallengeTopology{}).Error
}

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

func (r *TemplateRepository) Create(template *model.EnvironmentTemplate) error {
	return r.db.Create(template).Error
}

func (r *TemplateRepository) Update(template *model.EnvironmentTemplate) error {
	return r.db.Save(template).Error
}

func (r *TemplateRepository) Delete(id int64) error {
	return r.db.Delete(&model.EnvironmentTemplate{}, id).Error
}

func (r *TemplateRepository) FindByID(id int64) (*model.EnvironmentTemplate, error) {
	var template model.EnvironmentTemplate
	err := r.db.Where("id = ?", id).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *TemplateRepository) List(keyword string) ([]*model.EnvironmentTemplate, error) {
	var templates []*model.EnvironmentTemplate
	db := r.db.Model(&model.EnvironmentTemplate{})
	if keyword != "" {
		pattern := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR description LIKE ?", pattern, pattern)
	}
	err := db.Order("updated_at DESC").Find(&templates).Error
	return templates, err
}

func (r *TemplateRepository) IncrementUsage(id int64) error {
	return r.db.Model(&model.EnvironmentTemplate{}).
		Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr("usage_count + 1")).Error
}
