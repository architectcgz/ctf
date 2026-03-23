package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ctf-platform/internal/model"
	practiceapp "ctf-platform/internal/module/practice/application"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithDB(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) dbWithContext(ctx context.Context) *gorm.DB {
	if ctx == nil {
		ctx = context.Background()
	}
	return r.db.WithContext(ctx)
}

func (r *Repository) WithinTransaction(ctx context.Context, fn func(txRepo practiceapp.PracticeRepository) error) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(r.WithDB(tx))
	})
}

func (r *Repository) FindContestByIDWithContext(ctx context.Context, contestID int64) (*model.Contest, error) {
	var contest model.Contest
	if err := r.dbWithContext(ctx).Where("id = ?", contestID).First(&contest).Error; err != nil {
		return nil, err
	}
	return &contest, nil
}

func (r *Repository) FindContestChallengeWithContext(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error) {
	var contestChallenge model.ContestChallenge
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		First(&contestChallenge).Error; err != nil {
		return nil, err
	}
	return &contestChallenge, nil
}

func (r *Repository) FindContestRegistrationWithContext(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
	var registration model.ContestRegistration
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		First(&registration).Error; err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *Repository) LockInstanceScope(userID int64, scope practiceapp.InstanceScope) error {
	if scope.TeamID != nil {
		return r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", *scope.TeamID).
			First(&model.Team{}).Error
	}
	return r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", userID).
		First(&model.User{}).Error
}

func (r *Repository) FindScopedExistingInstance(userID, challengeID int64, scope practiceapp.InstanceScope) (*model.Instance, error) {
	query := r.db.Model(&model.Instance{}).
		Where("challenge_id = ? AND status IN ?", challengeID, []string{model.InstanceStatusCreating, model.InstanceStatusRunning})

	switch {
	case scope.TeamID != nil && scope.ContestID != nil:
		query = query.Where("contest_id = ? AND team_id = ?", *scope.ContestID, *scope.TeamID)
	case scope.ContestID != nil:
		query = query.Where("contest_id = ? AND user_id = ? AND team_id IS NULL", *scope.ContestID, userID)
	default:
		query = query.Where("user_id = ? AND contest_id IS NULL AND team_id IS NULL", userID)
	}

	var instance model.Instance
	if err := query.First(&instance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) CountScopedRunningInstances(userID int64, scope practiceapp.InstanceScope) (int, error) {
	query := r.db.Model(&model.Instance{}).
		Where("status IN ?", []string{model.InstanceStatusCreating, model.InstanceStatusRunning})

	switch {
	case scope.TeamID != nil && scope.ContestID != nil:
		query = query.Where("contest_id = ? AND team_id = ?", *scope.ContestID, *scope.TeamID)
	case scope.ContestID != nil:
		query = query.Where("contest_id = ? AND user_id = ? AND team_id IS NULL", *scope.ContestID, userID)
	default:
		query = query.Where("user_id = ? AND contest_id IS NULL AND team_id IS NULL", userID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *Repository) CreateInstance(instance *model.Instance) error {
	return r.db.Create(instance).Error
}

func (r *Repository) ReserveAvailablePort(start, end int) (int, error) {
	for port := start; port < end; port++ {
		if err := r.db.Create(&model.PortAllocation{Port: port}).Error; err != nil {
			if isPracticePortAllocationConflict(err) {
				continue
			}
			return 0, err
		}
		return port, nil
	}
	return 0, fmt.Errorf("no available port in range %d-%d", start, end)
}

func (r *Repository) BindReservedPort(port int, instanceID int64) error {
	return r.db.Model(&model.PortAllocation{}).
		Where("port = ?", port).
		Updates(map[string]any{
			"instance_id": instanceID,
			"updated_at":  time.Now(),
		}).Error
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

func isPracticePortAllocationConflict(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	lowered := strings.ToLower(err.Error())
	return strings.Contains(lowered, "unique constraint failed") ||
		strings.Contains(lowered, "duplicate key value") ||
		strings.Contains(lowered, "duplicate entry")
}
