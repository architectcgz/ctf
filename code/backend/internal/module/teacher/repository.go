package teacher

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type progressRow struct {
	Key    string
	Total  int
	Solved int
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

func (r *Repository) CountStudentsByClass(ctx context.Context, className string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("class_name = ? AND role = ? AND deleted_at IS NULL", className, model.RoleStudent).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count students by class: %w", err)
	}
	return count, nil
}

func (r *Repository) ListClasses(ctx context.Context) ([]dto.TeacherClassItem, error) {
	items := make([]dto.TeacherClassItem, 0)
	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Select("class_name AS name, COUNT(*) AS student_count").
		Where("role = ? AND class_name <> '' AND deleted_at IS NULL", model.RoleStudent).
		Group("class_name").
		Order("class_name ASC").
		Scan(&items).Error; err != nil {
		return nil, fmt.Errorf("list classes: %w", err)
	}
	return items, nil
}

func (r *Repository) ListStudentsByClass(ctx context.Context, className, keyword, studentNo string) ([]dto.TeacherStudentItem, error) {
	items := make([]dto.TeacherStudentItem, 0)
	query := r.db.WithContext(ctx).Model(&model.User{}).
		Select("id, username, NULLIF(name, '') AS name, NULLIF(student_no, '') AS student_no").
		Where("role = ? AND class_name = ? AND deleted_at IS NULL", model.RoleStudent, className)
	if keyword != "" {
		likeKeyword := "%" + strings.ToLower(keyword) + "%"
		query = query.Where("(LOWER(username) LIKE ? OR LOWER(name) LIKE ?)", likeKeyword, likeKeyword)
	}
	if studentNo != "" {
		query = query.Where("student_no = ?", studentNo)
	}
	if err := query.Order("COALESCE(NULLIF(student_no, ''), username) ASC, username ASC").
		Scan(&items).Error; err != nil {
		return nil, fmt.Errorf("list students by class: %w", err)
	}
	return items, nil
}

func (r *Repository) CountPublishedChallenges(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Challenge{}).
		Where("status = ?", model.ChallengeStatusPublished).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count published challenges: %w", err)
	}
	return count, nil
}

func (r *Repository) CountSolvedChallenges(ctx context.Context, userID int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Table("submissions AS s").
		Joins("JOIN challenges c ON c.id = s.challenge_id").
		Where("s.user_id = ? AND s.is_correct = ? AND c.status = ?", userID, true, model.ChallengeStatusPublished).
		Distinct("s.challenge_id").
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count solved challenges: %w", err)
	}
	return count, nil
}

func (r *Repository) GetCategoryProgress(ctx context.Context, userID int64) ([]progressRow, error) {
	rows := make([]progressRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			c.category AS key,
			COUNT(DISTINCT c.id) AS total,
			COUNT(DISTINCT CASE WHEN s.is_correct THEN c.id END) AS solved
		FROM challenges c
		LEFT JOIN submissions s
			ON s.challenge_id = c.id
			AND s.user_id = ?
			AND s.is_correct = TRUE
		WHERE c.status = ?
		GROUP BY c.category
		ORDER BY c.category
	`, userID, model.ChallengeStatusPublished).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get category progress: %w", err)
	}
	return rows, nil
}

func (r *Repository) GetDifficultyProgress(ctx context.Context, userID int64) ([]progressRow, error) {
	rows := make([]progressRow, 0)
	if err := r.db.WithContext(ctx).Raw(`
		SELECT
			c.difficulty AS key,
			COUNT(DISTINCT c.id) AS total,
			COUNT(DISTINCT CASE WHEN s.is_correct THEN c.id END) AS solved
		FROM challenges c
		LEFT JOIN submissions s
			ON s.challenge_id = c.id
			AND s.user_id = ?
			AND s.is_correct = TRUE
		WHERE c.status = ?
		GROUP BY c.difficulty
		ORDER BY
			CASE c.difficulty
				WHEN 'beginner' THEN 1
				WHEN 'easy' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'hard' THEN 4
				WHEN 'hell' THEN 5
				ELSE 99
			END
	`, userID, model.ChallengeStatusPublished).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("get difficulty progress: %w", err)
	}
	return rows, nil
}
