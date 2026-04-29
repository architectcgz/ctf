package infrastructure

import (
	"context"
	"errors"
	"fmt"

	"ctf-platform/internal/model"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *Repository) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	if err := r.dbWithContext(ctx).Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

// Upsert 插入或更新能力画像
func (r *Repository) Upsert(ctx context.Context, profile *model.SkillProfile) error {
	return r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "dimension"}},
		DoUpdates: clause.AssignmentColumns([]string{"score", "updated_at"}),
	}).Create(profile).Error
}

// FindByUserID 查询用户所有维度画像
func (r *Repository) FindByUserID(ctx context.Context, userID int64) ([]*model.SkillProfile, error) {
	var profiles []*model.SkillProfile
	err := r.dbWithContext(ctx).Where("user_id = ?", userID).Find(&profiles).Error
	return profiles, err
}

func (r *Repository) ListSolvedChallengeIDs(ctx context.Context, userID int64) ([]int64, error) {
	var ids []int64
	err := r.dbWithContext(ctx).Raw(`
		SELECT DISTINCT s.challenge_id AS challenge_id
		FROM submissions s
		WHERE s.user_id = ?
			AND s.is_correct = TRUE
			AND s.contest_id IS NULL
		ORDER BY challenge_id ASC
	`, userID).Scan(&ids).Error
	return ids, err
}

// BatchUpsert 批量插入或更新
func (r *Repository) BatchUpsert(ctx context.Context, profiles []*model.SkillProfile) error {
	if len(profiles) == 0 {
		return nil
	}
	return r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "dimension"}},
		DoUpdates: clause.AssignmentColumns([]string{"score", "updated_at"}),
	}).Create(profiles).Error
}

func (r *Repository) ListStudentIDs(ctx context.Context) ([]int64, error) {
	var ids []int64
	err := r.dbWithContext(ctx).Model(&model.User{}).
		Where("role = ? AND deleted_at IS NULL", model.RoleStudent).
		Pluck("id", &ids).Error
	return ids, err
}

// GetDimensionScores 查询用户各维度得分统计
func (r *Repository) GetDimensionScores(ctx context.Context, userID int64) ([]assessmentdomain.DimensionScore, error) {
	var scores []assessmentdomain.DimensionScore
	err := r.dbWithContext(ctx).Raw(`
		SELECT
			c.category AS dimension,
			COALESCE(SUM(c.points), 0) AS total_score,
			COALESCE(SUM(
				CASE WHEN EXISTS (
					SELECT 1
					FROM submissions s
					WHERE s.challenge_id = c.id
						AND s.user_id = ?
						AND s.is_correct = TRUE
						AND s.contest_id IS NULL
				) THEN c.points ELSE 0 END
			), 0) AS user_score
		FROM challenges c
		WHERE c.status = 'published'
		GROUP BY c.category
		ORDER BY c.category
	`, userID).Scan(&scores).Error
	return scores, err
}

// GetDimensionScore 查询用户单个维度得分统计（增量更新用）
func (r *Repository) GetDimensionScore(ctx context.Context, userID int64, dimension string) (*assessmentdomain.DimensionScore, error) {
	var score assessmentdomain.DimensionScore
	err := r.dbWithContext(ctx).Raw(`
		SELECT
			c.category AS dimension,
			COALESCE(SUM(c.points), 0) AS total_score,
			COALESCE(SUM(
				CASE WHEN EXISTS (
					SELECT 1
					FROM submissions s
					WHERE s.challenge_id = c.id
						AND s.user_id = ?
						AND s.is_correct = TRUE
						AND s.contest_id IS NULL
				) THEN c.points ELSE 0 END
			), 0) AS user_score
		FROM challenges c
		WHERE c.status = 'published' AND c.category = ?
		GROUP BY c.category
	`, userID, dimension).Scan(&score).Error
	if err != nil {
		return nil, err
	}
	return &score, nil
}
