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

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
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

func (r *Repository) WithinTransaction(ctx context.Context, fn func(txRepo practiceports.PracticeCommandTxRepository) error) error {
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

func (r *Repository) FindContestAWDServiceWithContext(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error) {
	var service model.ContestAWDService
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND id = ?", contestID, serviceID).
		Where("deleted_at IS NULL").
		First(&service).Error; err != nil {
		return nil, err
	}
	return &service, nil
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

func (r *Repository) LockInstanceScope(userID, challengeID int64, scope practiceports.InstanceScope) error {
	return r.LockInstanceScopeWithContext(context.Background(), userID, challengeID, scope)
}

func (r *Repository) LockInstanceScopeWithContext(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) error {
	if scope.ServiceID != nil {
		return r.dbWithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", *scope.ServiceID).
			First(&model.ContestAWDService{}).Error
	}
	switch scope.ShareScope {
	case model.InstanceSharingShared:
		return r.dbWithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", challengeID).
			First(&model.Challenge{}).Error
	case model.InstanceSharingPerTeam:
		if scope.TeamID != nil {
			return r.dbWithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", *scope.TeamID).
				First(&model.Team{}).Error
		}
	}
	if scope.TeamID != nil && scope.ShareScope == model.InstanceSharingPerTeam {
		return r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", *scope.TeamID).
			First(&model.Team{}).Error
	}
	return r.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", userID).
		First(&model.User{}).Error
}

func (r *Repository) FindScopedExistingInstance(userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error) {
	return r.FindScopedExistingInstanceWithContext(context.Background(), userID, challengeID, scope)
}

func (r *Repository) FindScopedExistingInstanceWithContext(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error) {
	now := time.Now()
	query := r.dbWithContext(ctx).Model(&model.Instance{}).
		Where("share_scope = ?", scope.ShareScope).
		Where(
			"(status IN ? OR (status = ? AND expires_at > ?))",
			[]string{model.InstanceStatusPending, model.InstanceStatusCreating},
			model.InstanceStatusRunning,
			now,
		)
	if scope.ServiceID != nil {
		query = query.Where("service_id = ?", *scope.ServiceID)
	} else {
		query = query.Where("challenge_id = ?", challengeID)
	}

	switch {
	case scope.ShareScope == model.InstanceSharingShared && scope.ContestID != nil:
		query = query.Where("contest_id = ? AND team_id IS NULL", *scope.ContestID)
	case scope.ShareScope == model.InstanceSharingShared:
		query = query.Where("contest_id IS NULL AND team_id IS NULL")
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

func (r *Repository) CountScopedRunningInstances(userID int64, scope practiceports.InstanceScope) (int, error) {
	return r.CountScopedRunningInstancesWithContext(context.Background(), userID, scope)
}

func (r *Repository) CountScopedRunningInstancesWithContext(ctx context.Context, userID int64, scope practiceports.InstanceScope) (int, error) {
	now := time.Now()
	query := r.dbWithContext(ctx).Model(&model.Instance{}).
		Where("share_scope = ?", scope.ShareScope).
		Where(
			"(status IN ? OR (status = ? AND expires_at > ?))",
			[]string{model.InstanceStatusPending, model.InstanceStatusCreating},
			model.InstanceStatusRunning,
			now,
		)

	switch {
	case scope.ShareScope == model.InstanceSharingShared && scope.ContestID != nil:
		query = query.Where("contest_id = ? AND team_id IS NULL", *scope.ContestID)
	case scope.ShareScope == model.InstanceSharingShared:
		query = query.Where("contest_id IS NULL AND team_id IS NULL")
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

func (r *Repository) RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error {
	return r.dbWithContext(ctx).Model(&model.Instance{}).
		Where("id = ?", instanceID).
		Updates(map[string]any{
			"expires_at": expiresAt,
			"updated_at": time.Now(),
		}).Error
}

func (r *Repository) CreateInstance(instance *model.Instance) error {
	return r.CreateInstanceWithContext(context.Background(), instance)
}

func (r *Repository) CreateInstanceWithContext(ctx context.Context, instance *model.Instance) error {
	return r.dbWithContext(ctx).Create(instance).Error
}

func (r *Repository) ReserveAvailablePort(start, end int) (int, error) {
	return r.ReserveAvailablePortWithContext(context.Background(), start, end)
}

func (r *Repository) ReserveAvailablePortWithContext(ctx context.Context, start, end int) (int, error) {
	for port := start; port < end; port++ {
		if err := r.dbWithContext(ctx).Create(&model.PortAllocation{Port: port}).Error; err != nil {
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
	return r.BindReservedPortWithContext(context.Background(), port, instanceID)
}

func (r *Repository) BindReservedPortWithContext(ctx context.Context, port int, instanceID int64) error {
	return r.dbWithContext(ctx).Model(&model.PortAllocation{}).
		Where("port = ?", port).
		Updates(map[string]any{
			"instance_id": instanceID,
			"updated_at":  time.Now(),
		}).Error
}

// CreateSubmission 创建提交记录
func (r *Repository) CreateSubmission(submission *model.Submission) error {
	return r.CreateSubmissionWithContext(context.Background(), submission)
}

func (r *Repository) CreateSubmissionWithContext(ctx context.Context, submission *model.Submission) error {
	return r.dbWithContext(ctx).Create(submission).Error
}

// FindCorrectSubmission 查找用户是否已正确提交过该题
func (r *Repository) FindCorrectSubmission(userID, challengeID int64) (*model.Submission, error) {
	return r.FindCorrectSubmissionWithContext(context.Background(), userID, challengeID)
}

func (r *Repository) FindCorrectSubmissionWithContext(ctx context.Context, userID, challengeID int64) (*model.Submission, error) {
	var submission model.Submission
	err := r.dbWithContext(ctx).Where("user_id = ? AND challenge_id = ? AND is_correct = ?", userID, challengeID, true).
		First(&submission).Error
	return &submission, err
}

func (r *Repository) FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*model.Instance, error) {
	var instance model.Instance
	err := r.dbWithContext(ctx).
		Where("user_id = ? AND contest_id IS NULL AND team_id IS NULL AND challenge_id = ? AND status IN ?", userID, challengeID,
			[]string{model.InstanceStatusPending, model.InstanceStatusCreating, model.InstanceStatusRunning}).
		First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) ListChallengeSubmissions(userID, challengeID int64, limit int) ([]model.Submission, error) {
	return r.ListChallengeSubmissionsWithContext(context.Background(), userID, challengeID, limit)
}

func (r *Repository) ListChallengeSubmissionsWithContext(ctx context.Context, userID, challengeID int64, limit int) ([]model.Submission, error) {
	if limit <= 0 {
		limit = 20
	}

	var submissions []model.Submission
	err := r.dbWithContext(ctx).
		Where("user_id = ? AND challenge_id = ? AND contest_id IS NULL", userID, challengeID).
		Order("submitted_at DESC, id DESC").
		Limit(limit).
		Find(&submissions).Error
	return submissions, err
}

func (r *Repository) UpdateSubmission(submission *model.Submission) error {
	return r.UpdateSubmissionWithContext(context.Background(), submission)
}

func (r *Repository) UpdateSubmissionWithContext(ctx context.Context, submission *model.Submission) error {
	return r.dbWithContext(ctx).Save(submission).Error
}

func (r *Repository) FindUserByID(userID int64) (*model.User, error) {
	return r.FindUserByIDWithContext(context.Background(), userID)
}

func (r *Repository) FindUserByIDWithContext(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	if err := r.dbWithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

type teacherManualReviewSubmissionRow struct {
	ID              int64
	UserID          int64
	ChallengeID     int64
	ContestID       *int64
	Flag            string
	IsCorrect       bool
	ReviewStatus    string
	ReviewedBy      *int64
	ReviewedAt      *time.Time
	ReviewComment   string
	Score           int
	SubmittedAt     time.Time
	UpdatedAt       time.Time
	StudentUsername string
	StudentName     string
	ClassName       string
	ChallengeTitle  string
	ReviewerName    string
}

func (r teacherManualReviewSubmissionRow) toRecord() practiceports.TeacherManualReviewSubmissionRecord {
	return practiceports.TeacherManualReviewSubmissionRecord{
		Submission: model.Submission{
			ID:            r.ID,
			UserID:        r.UserID,
			ChallengeID:   r.ChallengeID,
			ContestID:     r.ContestID,
			Flag:          r.Flag,
			IsCorrect:     r.IsCorrect,
			ReviewStatus:  r.ReviewStatus,
			ReviewedBy:    r.ReviewedBy,
			ReviewedAt:    r.ReviewedAt,
			ReviewComment: r.ReviewComment,
			Score:         r.Score,
			SubmittedAt:   r.SubmittedAt,
			UpdatedAt:     r.UpdatedAt,
		},
		StudentUsername: r.StudentUsername,
		StudentName:     r.StudentName,
		ClassName:       r.ClassName,
		ChallengeTitle:  r.ChallengeTitle,
		ReviewerName:    r.ReviewerName,
	}
}

func (r *Repository) GetTeacherManualReviewSubmissionByID(id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
	return r.GetTeacherManualReviewSubmissionByIDWithContext(context.Background(), id)
}

func (r *Repository) GetTeacherManualReviewSubmissionByIDWithContext(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
	rows, _, err := r.listTeacherManualReviewSubmissions(ctx, &dto.TeacherManualReviewSubmissionQuery{
		Page: 1,
		Size: 1,
	}, func(db *gorm.DB) *gorm.DB {
		return db.Where("s.id = ?", id)
	})
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	record := rows[0]
	return &record, nil
}

func (r *Repository) ListTeacherManualReviewSubmissions(query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
	return r.ListTeacherManualReviewSubmissionsWithContext(context.Background(), query)
}

func (r *Repository) ListTeacherManualReviewSubmissionsWithContext(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
	return r.listTeacherManualReviewSubmissions(ctx, query, nil)
}

func (r *Repository) listTeacherManualReviewSubmissions(
	ctx context.Context,
	query *dto.TeacherManualReviewSubmissionQuery,
	extra func(db *gorm.DB) *gorm.DB,
) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
	base := r.dbWithContext(ctx).Table("submissions AS s").
		Select(strings.TrimSpace(`
			s.id,
			s.user_id,
			s.challenge_id,
			s.contest_id,
			s.flag,
			s.is_correct,
			s.review_status,
			s.reviewed_by,
			s.reviewed_at,
			s.review_comment,
			s.score,
			s.submitted_at,
			s.updated_at,
			u.username AS student_username,
			COALESCE(u.name, '') AS student_name,
			COALESCE(u.class_name, '') AS class_name,
			c.title AS challenge_title,
			COALESCE(reviewer.name, reviewer.username, '') AS reviewer_name
		`)).
		Joins("JOIN users u ON u.id = s.user_id").
		Joins("JOIN challenges c ON c.id = s.challenge_id").
		Joins("LEFT JOIN users reviewer ON reviewer.id = s.reviewed_by").
		Where("c.flag_type = ?", model.FlagTypeManualReview)

	if query != nil {
		if query.StudentID != nil {
			base = base.Where("s.user_id = ?", *query.StudentID)
		}
		if query.ChallengeID != nil {
			base = base.Where("s.challenge_id = ?", *query.ChallengeID)
		}
		if query.ClassName != "" {
			base = base.Where("u.class_name = ?", query.ClassName)
		}
		if query.ReviewStatus != "" {
			base = base.Where("s.review_status = ?", query.ReviewStatus)
		}
	}
	if extra != nil {
		base = extra(base)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := 1
	size := 20
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.Size > 0 {
			size = query.Size
		}
	}
	offset := (page - 1) * size

	var rows []teacherManualReviewSubmissionRow
	if err := base.Order("s.updated_at DESC, s.id DESC").Offset(offset).Limit(size).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	items := make([]practiceports.TeacherManualReviewSubmissionRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toRecord())
	}
	return items, total, nil
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
