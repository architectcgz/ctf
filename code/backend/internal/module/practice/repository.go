package practice

import (
	"ctf-platform/internal/model"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// CreateSubmission 创建提交记录
func (r *Repository) CreateSubmission(submission *model.Submission) error {
	return r.db.Create(submission).Error
}

// FindCorrectSubmission 查找用户是否已正确提交过该题
func (r *Repository) FindCorrectSubmission(userID, challengeID int64) (*model.Submission, error) {
	var submission model.Submission
	err := r.db.Where("user_id = ? AND challenge_id = ? AND is_correct = ?", userID, challengeID, true).
		First(&submission).Error
	return &submission, err
}

// CountRecentSubmissions 统计时间窗口内的提交次数
func (r *Repository) CountRecentSubmissions(userID, challengeID int64, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&model.Submission{}).
		Where("user_id = ? AND challenge_id = ? AND submitted_at >= ?", userID, challengeID, since).
		Count(&count).Error
	return count, err
}

// IsUniqueViolation 检测是否为唯一约束冲突错误
func (r *Repository) IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// 23505 是 PostgreSQL 唯一约束冲突错误码
		return pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "idx_submissions_user_challenge_correct")
	}
	return false
}

// GetUserProgress 获取用户解题进度统计
func (r *Repository) GetUserProgress(userID int64) (totalScore int, totalSolved int, err error) {
	var result struct {
		TotalScore  int `gorm:"column:total_score"`
		TotalSolved int `gorm:"column:total_solved"`
	}
	err = r.db.Table("submissions s").
		Select("COALESCE(SUM(c.points), 0) as total_score, COUNT(DISTINCT s.challenge_id) as total_solved").
		Joins("JOIN challenges c ON s.challenge_id = c.id").
		Where("s.user_id = ? AND s.is_correct = ? AND c.status = ?", userID, true, "published").
		Scan(&result).Error
	return result.TotalScore, result.TotalSolved, err
}

// GetCategoryStats 获取分类统计
func (r *Repository) GetCategoryStats(userID int64) ([]struct {
	Category string
	Solved   int
	Total    int
}, error) {
	var stats []struct {
		Category string
		Solved   int
		Total    int
	}
	err := r.db.Raw(`
		SELECT
			c.category,
			COUNT(DISTINCT CASE WHEN s.is_correct THEN c.id END) as solved,
			COUNT(DISTINCT c.id) as total
		FROM challenges c
		LEFT JOIN submissions s ON c.id = s.challenge_id AND s.user_id = ? AND s.is_correct = true
		WHERE c.status = ?
		GROUP BY c.category
		ORDER BY c.category
	`, userID, "published").Scan(&stats).Error
	return stats, err
}

// GetDifficultyStats 获取难度统计
func (r *Repository) GetDifficultyStats(userID int64) ([]struct {
	Difficulty string
	Solved     int
	Total      int
}, error) {
	var stats []struct {
		Difficulty string
		Solved     int
		Total      int
	}
	err := r.db.Raw(`
		SELECT
			c.difficulty,
			COUNT(DISTINCT CASE WHEN s.is_correct THEN c.id END) as solved,
			COUNT(DISTINCT c.id) as total
		FROM challenges c
		LEFT JOIN submissions s ON c.id = s.challenge_id AND s.user_id = ? AND s.is_correct = true
		WHERE c.status = ?
		GROUP BY c.difficulty
		ORDER BY
			CASE c.difficulty
				WHEN 'beginner' THEN 1
				WHEN 'easy' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'hard' THEN 4
				WHEN 'insane' THEN 5
			END
	`, userID, "published").Scan(&stats).Error
	return stats, err
}

// GetUserRank 获取用户排名
func (r *Repository) GetUserRank(userID int64) (int, error) {
	var rank int
	err := r.db.Raw(`
		WITH ranked_users AS (
			SELECT
				s.user_id,
				RANK() OVER (ORDER BY SUM(c.points) DESC) as rank
			FROM submissions s
			JOIN challenges c ON s.challenge_id = c.id
			WHERE s.is_correct = true AND c.status = ?
			GROUP BY s.user_id
		)
		SELECT COALESCE(rank, 0) FROM ranked_users WHERE user_id = ?
	`, "published", userID).Scan(&rank).Error
	if err != nil {
		return 0, err
	}
	if rank == 0 {
		rank = 1
	}
	return rank, nil
}

// GetUserTimeline 获取用户时间线
func (r *Repository) GetUserTimeline(userID int64, limit int) ([]struct {
	Type        string
	ChallengeID int64
	Title       string
	Timestamp   time.Time
	IsCorrect   *bool
	Points      *int
}, error) {
	var events []struct {
		Type        string
		ChallengeID int64
		Title       string
		Timestamp   time.Time
		IsCorrect   *bool
		Points      *int
	}

	if limit <= 0 {
		limit = 100
	}

	err := r.db.Raw(`
		SELECT * FROM (
			SELECT 'instance_start' as type, i.challenge_id, i.created_at as timestamp,
				NULL::boolean as is_correct, NULL::integer as points
			FROM instances i
			WHERE i.user_id = ?
			UNION ALL
			SELECT 'flag_submit' as type, s.challenge_id, s.submitted_at as timestamp,
				s.is_correct, CASE WHEN s.is_correct THEN c.points ELSE NULL END as points
			FROM submissions s
			LEFT JOIN challenges c ON s.challenge_id = c.id
			WHERE s.user_id = ?
			UNION ALL
			SELECT 'instance_destroy' as type, i.challenge_id, i.updated_at as timestamp,
				NULL::boolean as is_correct, NULL::integer as points
			FROM instances i
			WHERE i.user_id = ? AND i.status IN ('stopped', 'expired')
		) events
		LEFT JOIN challenges c ON events.challenge_id = c.id
		ORDER BY events.timestamp DESC
		LIMIT ?
	`, userID, userID, userID, limit).Scan(&events).Error

	return events, err
}
