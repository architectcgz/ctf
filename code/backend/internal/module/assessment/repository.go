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

// DimensionScore 维度得分统计
type DimensionScore struct {
	Dimension  string
	TotalScore int
	UserScore  int
}

// GetDimensionScores 查询用户各维度得分统计
func (r *Repository) GetDimensionScores(userID int64) ([]DimensionScore, error) {
	var scores []DimensionScore
	err := r.db.Raw(`
		SELECT
			c.category as dimension,
			SUM(c.points) as total_score,
			COALESCE(SUM(CASE WHEN s.is_correct = 1 THEN c.points ELSE 0 END), 0) as user_score
		FROM challenges c
		LEFT JOIN submissions s ON c.id = s.challenge_id AND s.user_id = ?
		WHERE c.status = 'published'
		GROUP BY c.category
	`, userID).Scan(&scores).Error
	return scores, err
}

// GetDimensionScore 查询用户单个维度得分统计（增量更新用）
func (r *Repository) GetDimensionScore(userID int64, dimension string) (*DimensionScore, error) {
	var score DimensionScore
	err := r.db.Raw(`
		SELECT
			c.category as dimension,
			SUM(c.points) as total_score,
			COALESCE(SUM(CASE WHEN s.is_correct = 1 THEN c.points ELSE 0 END), 0) as user_score
		FROM challenges c
		LEFT JOIN submissions s ON c.id = s.challenge_id AND s.user_id = ?
		WHERE c.status = 'published' AND c.category = ?
		GROUP BY c.category
	`, userID, dimension).Scan(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}
