package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	assessmentapp "ctf-platform/internal/module/assessment/application"
)

type ReportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) Create(ctx context.Context, report *model.Report) error {
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *ReportRepository) FindByID(ctx context.Context, reportID int64) (*model.Report, error) {
	var report model.Report
	if err := r.db.WithContext(ctx).First(&report, reportID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcodeReportNotFound()
		}
		return nil, err
	}
	return &report, nil
}

func (r *ReportRepository) MarkReady(ctx context.Context, reportID int64, filePath string, expiresAt time.Time) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.Report{}).
		Where("id = ?", reportID).
		Updates(map[string]any{
			"status":       model.ReportStatusReady,
			"file_path":    filePath,
			"expires_at":   expiresAt,
			"error_msg":    nil,
			"completed_at": &now,
		}).Error
}

func (r *ReportRepository) MarkFailed(ctx context.Context, reportID int64, message string) error {
	return r.db.WithContext(ctx).Model(&model.Report{}).
		Where("id = ?", reportID).
		Updates(map[string]any{
			"status":    model.ReportStatusFailed,
			"error_msg": &message,
		}).Error
}

func (r *ReportRepository) FindUserByID(ctx context.Context, userID int64) (*assessmentapp.ReportUser, error) {
	var user assessmentapp.ReportUser
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Select("id, username, class_name, role").
		Where("id = ? AND deleted_at IS NULL", userID).
		Scan(&user).Error
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return &user, nil
}

func (r *ReportRepository) GetPersonalStats(ctx context.Context, userID int64) (*assessmentapp.PersonalReportStats, error) {
	var stats assessmentapp.PersonalReportStats
	err := r.db.WithContext(ctx).Raw(`
		WITH solved AS (
			SELECT DISTINCT challenge_id
			FROM submissions
			WHERE user_id = ? AND is_correct = TRUE AND contest_id IS NULL
		),
		user_scores AS (
			SELECT
				u.id AS user_id,
				COALESCE(SUM(c.points), 0) AS total_score
			FROM users u
			LEFT JOIN (
				SELECT DISTINCT user_id, challenge_id
				FROM submissions
				WHERE is_correct = TRUE AND contest_id IS NULL
			) solved_challenges ON solved_challenges.user_id = u.id
			LEFT JOIN challenges c ON c.id = solved_challenges.challenge_id AND c.status = 'published'
			WHERE u.deleted_at IS NULL
			GROUP BY u.id
		),
		ranked AS (
			SELECT user_id, RANK() OVER (ORDER BY total_score DESC, user_id ASC) AS rank
			FROM user_scores
		)
		SELECT
			COALESCE((
				SELECT SUM(c.points)
				FROM solved s
				JOIN challenges c ON c.id = s.challenge_id AND c.status = 'published'
			), 0) AS total_score,
			COALESCE((SELECT COUNT(*) FROM solved), 0) AS total_solved,
			COALESCE((
				SELECT COUNT(*)
				FROM submissions
				WHERE user_id = ? AND contest_id IS NULL
			), 0) AS total_attempts,
			COALESCE((SELECT rank FROM ranked WHERE user_id = ?), 1) AS rank
	`, userID, userID, userID).Scan(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *ReportRepository) ListPersonalDimensionStats(ctx context.Context, userID int64) ([]assessmentapp.ReportDimensionStat, error) {
	stats := make([]assessmentapp.ReportDimensionStat, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT
			c.category AS dimension,
			COUNT(DISTINCT CASE
				WHEN s.user_id = ? AND s.is_correct = TRUE AND s.contest_id IS NULL
				THEN c.id
			END) AS solved,
			COUNT(DISTINCT c.id) AS total
		FROM challenges c
		LEFT JOIN submissions s ON s.challenge_id = c.id
		WHERE c.status = 'published'
		GROUP BY c.category
		ORDER BY c.category
	`, userID).Scan(&stats).Error
	return stats, err
}

func (r *ReportRepository) CountClassStudents(ctx context.Context, className string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("class_name = ? AND role = ? AND deleted_at IS NULL", className, model.RoleStudent).
		Count(&count).Error
	return int(count), err
}

func (r *ReportRepository) GetClassAverageScore(ctx context.Context, className string) (float64, error) {
	var avgScore float64
	err := r.db.WithContext(ctx).Raw(`
		WITH user_scores AS (
			SELECT
				u.id AS user_id,
				COALESCE(SUM(c.points), 0) AS total_score
			FROM users u
			LEFT JOIN (
				SELECT DISTINCT user_id, challenge_id
				FROM submissions
				WHERE is_correct = TRUE AND contest_id IS NULL
			) solved ON solved.user_id = u.id
			LEFT JOIN challenges c ON c.id = solved.challenge_id AND c.status = 'published'
			WHERE u.class_name = ? AND u.role = ? AND u.deleted_at IS NULL
			GROUP BY u.id
		)
		SELECT COALESCE(AVG(total_score), 0) AS avg_score
		FROM user_scores
	`, className, model.RoleStudent).Scan(&avgScore).Error
	return avgScore, err
}

func (r *ReportRepository) ListClassDimensionAverages(ctx context.Context, className string) ([]assessmentapp.ClassDimensionAverage, error) {
	rows := make([]assessmentapp.ClassDimensionAverage, 0)
	err := r.db.WithContext(ctx).Raw(`
		SELECT sp.dimension, COALESCE(AVG(sp.score), 0) AS avg_score
		FROM users u
		JOIN skill_profiles sp ON sp.user_id = u.id
		WHERE u.class_name = ? AND u.role = ? AND u.deleted_at IS NULL
		GROUP BY sp.dimension
		ORDER BY sp.dimension
	`, className, model.RoleStudent).Scan(&rows).Error
	return rows, err
}

func (r *ReportRepository) ListClassTopStudents(ctx context.Context, className string, limit int) ([]assessmentapp.ClassTopStudent, error) {
	rows := make([]assessmentapp.ClassTopStudent, 0)
	err := r.db.WithContext(ctx).Raw(`
		WITH user_scores AS (
			SELECT
				u.id AS user_id,
				u.username,
				COALESCE(SUM(c.points), 0) AS total_score
			FROM users u
			LEFT JOIN (
				SELECT DISTINCT user_id, challenge_id
				FROM submissions
				WHERE is_correct = TRUE AND contest_id IS NULL
			) solved ON solved.user_id = u.id
			LEFT JOIN challenges c ON c.id = solved.challenge_id AND c.status = 'published'
			WHERE u.class_name = ? AND u.role = ? AND u.deleted_at IS NULL
			GROUP BY u.id, u.username
		)
		SELECT
			user_id,
			username,
			total_score,
			RANK() OVER (ORDER BY total_score DESC, user_id ASC) AS rank
		FROM user_scores
		ORDER BY total_score DESC, user_id ASC
		LIMIT ?
	`, className, model.RoleStudent, limit).Scan(&rows).Error
	return rows, err
}

func errcodeReportNotFound() error {
	return gorm.ErrRecordNotFound
}
