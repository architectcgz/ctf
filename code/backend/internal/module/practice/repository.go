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
