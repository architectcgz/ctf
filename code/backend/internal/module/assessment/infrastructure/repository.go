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
	if ctx == nil {
		ctx = context.Background()
	}
	return r.db.WithContext(ctx)
}

func (r *Repository) FindUserByID(userID int64) (*model.User, error) {
	return r.FindUserByIDWithContext(context.Background(), userID)
}

func (r *Repository) FindUserByIDWithContext(ctx context.Context, userID int64) (*model.User, error) {
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
func (r *Repository) Upsert(profile *model.SkillProfile) error {
	return r.UpsertWithContext(context.Background(), profile)
}

func (r *Repository) UpsertWithContext(ctx context.Context, profile *model.SkillProfile) error {
	return r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "dimension"}},
		DoUpdates: clause.AssignmentColumns([]string{"score", "updated_at"}),
	}).Create(profile).Error
}

// FindByUserID 查询用户所有维度画像
func (r *Repository) FindByUserID(userID int64) ([]*model.SkillProfile, error) {
	return r.FindByUserIDWithContext(context.Background(), userID)
}

func (r *Repository) FindByUserIDWithContext(ctx context.Context, userID int64) ([]*model.SkillProfile, error) {
	var profiles []*model.SkillProfile
	err := r.dbWithContext(ctx).Where("user_id = ?", userID).Find(&profiles).Error
	return profiles, err
}

func (r *Repository) ListSolvedChallengeIDsWithContext(ctx context.Context, userID int64) ([]int64, error) {
	var ids []int64
	err := r.dbWithContext(ctx).Model(&model.Submission{}).
		Where("user_id = ? AND is_correct = ?", userID, true).
		Distinct("challenge_id").
		Pluck("challenge_id", &ids).Error
	return ids, err
}

// BatchUpsert 批量插入或更新
func (r *Repository) BatchUpsert(profiles []*model.SkillProfile) error {
	return r.BatchUpsertWithContext(context.Background(), profiles)
}

func (r *Repository) BatchUpsertWithContext(ctx context.Context, profiles []*model.SkillProfile) error {
	if len(profiles) == 0 {
		return nil
	}
	return r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "dimension"}},
		DoUpdates: clause.AssignmentColumns([]string{"score", "updated_at"}),
	}).Create(profiles).Error
}

func (r *Repository) ListStudentIDs() ([]int64, error) {
	return r.ListStudentIDsWithContext(context.Background())
}

func (r *Repository) ListStudentIDsWithContext(ctx context.Context) ([]int64, error) {
	var ids []int64
	err := r.dbWithContext(ctx).Model(&model.User{}).
		Where("role = ? AND deleted_at IS NULL", model.RoleStudent).
		Pluck("id", &ids).Error
	return ids, err
}

// GetDimensionScores 查询用户各维度得分统计
func (r *Repository) GetDimensionScores(userID int64) ([]assessmentdomain.DimensionScore, error) {
	return r.GetDimensionScoresWithContext(context.Background(), userID)
}

func (r *Repository) GetDimensionScoresWithContext(ctx context.Context, userID int64) ([]assessmentdomain.DimensionScore, error) {
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
func (r *Repository) GetDimensionScore(userID int64, dimension string) (*assessmentdomain.DimensionScore, error) {
	return r.GetDimensionScoreWithContext(context.Background(), userID, dimension)
}

func (r *Repository) GetDimensionScoreWithContext(ctx context.Context, userID int64, dimension string) (*assessmentdomain.DimensionScore, error) {
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
