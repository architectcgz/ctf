package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ctf-platform/internal/model"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeportreservation "ctf-platform/internal/module/runtime/contracts/portreservation"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type Repository struct {
	db           *gorm.DB
	runtimePorts runtimeports.PortReservationOwner
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		db:           db,
		runtimePorts: runtimeportreservation.NewOwner(db),
	}
}

func (r *Repository) WithDB(db *gorm.DB) *Repository {
	return &Repository{
		db:           db,
		runtimePorts: runtimeportreservation.NewOwner(db),
	}
}

func (r *Repository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *Repository) WithinInstanceStartTx(ctx context.Context, fn func(txRepo practiceports.PracticeInstanceStartTxRepository) error) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(r.WithDB(tx))
	})
}

func (r *Repository) WithinInstanceRestartTx(ctx context.Context, fn func(txRepo practiceports.PracticeInstanceRestartTxRepository) error) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(r.WithDB(tx))
	})
}

func (r *Repository) WithinAWDServiceOperationTx(ctx context.Context, fn func(txRepo practiceports.PracticeAWDServiceOperationTxRepository) error) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(r.WithDB(tx))
	})
}

func (r *Repository) FindContestByID(ctx context.Context, contestID int64) (*practiceports.ContestRecord, error) {
	var contest contestRow
	if err := r.dbWithContext(ctx).Where("id = ?", contestID).First(&contest).Error; err != nil {
		return nil, err
	}
	return contest.toRecord(), nil
}

func (r *Repository) ListDesiredRuntimeAWDContests(ctx context.Context) ([]*practiceports.ContestRecord, error) {
	var contests []*contestRow
	if err := r.dbWithContext(ctx).
		Where("mode = ? AND status IN ? AND deleted_at IS NULL",
			practiceports.ContestModeAWD,
			[]string{practiceports.ContestStatusRunning, practiceports.ContestStatusFrozen},
		).
		Order("id ASC").
		Find(&contests).Error; err != nil {
		return nil, err
	}
	return contestRowsToRecords(contests), nil
}

func (r *Repository) FindContestChallenge(ctx context.Context, contestID, challengeID int64) (*practiceports.ContestChallengeRecord, error) {
	var contestChallenge contestChallengeProjection
	if err := r.dbWithContext(ctx).
		Table("contest_challenges").
		Select("contest_id, challenge_id, is_visible").
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Where("deleted_at IS NULL").
		Take(&contestChallenge).Error; err != nil {
		return nil, err
	}
	return contestChallenge.toRecord(), nil
}

func (r *Repository) FindContestAWDService(ctx context.Context, contestID, serviceID int64) (*practiceports.ContestAWDServiceRecord, error) {
	var service contestAWDServiceRow
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND id = ?", contestID, serviceID).
		Where("deleted_at IS NULL").
		First(&service).Error; err != nil {
		return nil, err
	}
	return service.toRecord(), nil
}

func (r *Repository) FindContestAWDServiceRuntimeSubject(ctx context.Context, contestID, serviceID int64) (*practiceports.ContestAWDServiceRuntimeSubject, error) {
	var service contestAWDServiceRow
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND id = ?", contestID, serviceID).
		Where("deleted_at IS NULL").
		First(&service).Error; err != nil {
		return nil, err
	}
	return buildContestAWDServiceRuntimeSubject(service.toRecord())
}

func (r *Repository) ListContestAWDServices(ctx context.Context, contestID int64) ([]*practiceports.ContestAWDServiceRecord, error) {
	var services []*contestAWDServiceRow
	if err := r.dbWithContext(ctx).
		Where("contest_id = ?", contestID).
		Where("deleted_at IS NULL").
		Order("\"order\" ASC, id ASC").
		Find(&services).Error; err != nil {
		return nil, err
	}
	return contestAWDServiceRowsToRecords(services), nil
}

func (r *Repository) ListContestAWDInstances(ctx context.Context, contestID int64) ([]*instancecontracts.Instance, error) {
	var instances []*instancecontracts.Instance
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND team_id IS NOT NULL AND service_id IS NOT NULL", contestID).
		Where("status IN ?", []string{
			instancecontracts.InstanceStatusPending,
			instancecontracts.InstanceStatusCreating,
			instancecontracts.InstanceStatusRunning,
		}).
		Order("created_at DESC").
		Find(&instances).Error; err != nil {
		return nil, err
	}
	return instances, nil
}

func (r *Repository) FindContestTeam(ctx context.Context, contestID, teamID int64) (*practiceports.ContestTeamRecord, error) {
	var team contestTeamRow
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND id = ?", contestID, teamID).
		First(&team).Error; err != nil {
		return nil, err
	}
	return team.toRecord(), nil
}

func (r *Repository) ListContestTeams(ctx context.Context, contestID int64) ([]*practiceports.ContestTeamRecord, error) {
	var teams []*contestTeamRow
	if err := r.dbWithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("created_at ASC, id ASC").
		Find(&teams).Error; err != nil {
		return nil, err
	}
	return contestTeamRowsToRecords(teams), nil
}

func (r *Repository) FindContestRegistration(ctx context.Context, contestID, userID int64) (*practiceports.ContestParticipation, error) {
	var registration contestRegistrationProjection
	if err := r.dbWithContext(ctx).
		Table("contest_registrations").
		Select("team_id, status").
		Where("contest_id = ? AND user_id = ?", contestID, userID).
		Take(&registration).Error; err != nil {
		return nil, err
	}
	return &practiceports.ContestParticipation{
		Status: registration.Status,
		TeamID: registration.TeamID,
	}, nil
}

func (r *Repository) ListContestAWDScopeControls(ctx context.Context, contestID int64) ([]*runtimecontracts.AWDScopeControl, error) {
	var controls []*runtimecontracts.AWDScopeControl
	if err := r.dbWithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("team_id ASC, scope_type ASC, service_id ASC, control_type ASC, id ASC").
		Find(&controls).Error; err != nil {
		return nil, err
	}
	return controls, nil
}

func (r *Repository) ListScopeAWDScopeControls(ctx context.Context, contestID, teamID, serviceID int64) ([]*runtimecontracts.AWDScopeControl, error) {
	var controls []*runtimecontracts.AWDScopeControl
	query := r.dbWithContext(ctx).
		Where("contest_id = ? AND team_id = ?", contestID, teamID)
	if serviceID > 0 {
		query = query.Where(
			"(scope_type = ? AND service_id = 0) OR (scope_type = ? AND service_id = ?)",
			runtimecontracts.AWDScopeControlScopeTeam,
			runtimecontracts.AWDScopeControlScopeTeamService,
			serviceID,
		)
	} else {
		query = query.Where("scope_type = ? AND service_id = 0", runtimecontracts.AWDScopeControlScopeTeam)
	}
	if err := query.
		Order("scope_type ASC, service_id ASC, control_type ASC, id ASC").
		Find(&controls).Error; err != nil {
		return nil, err
	}
	return controls, nil
}

func (r *Repository) UpsertAWDScopeControl(ctx context.Context, control *runtimecontracts.AWDScopeControl) error {
	if control == nil {
		return nil
	}

	now := time.Now().UTC()
	if control.CreatedAt.IsZero() {
		control.CreatedAt = now
	}
	control.UpdatedAt = now

	return r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "contest_id"},
			{Name: "team_id"},
			{Name: "scope_type"},
			{Name: "service_id"},
			{Name: "control_type"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"reason":     control.Reason,
			"updated_by": control.UpdatedBy,
			"updated_at": now,
		}),
	}).Create(control).Error
}

func (r *Repository) DeleteAWDScopeControl(ctx context.Context, contestID, teamID int64, scopeType, controlType string, serviceID int64) error {
	return r.dbWithContext(ctx).
		Where("contest_id = ? AND team_id = ? AND scope_type = ? AND control_type = ? AND service_id = ?",
			contestID, teamID, scopeType, controlType, serviceID).
		Delete(&runtimecontracts.AWDScopeControl{}).Error
}

func (r *Repository) LockInstanceScope(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) error {
	if scope.ServiceID != nil {
		return r.dbWithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", *scope.ServiceID).
			First(&contestAWDServiceRow{}).Error
	}
	switch scope.ShareScope {
	case instancecontracts.ShareScopeShared:
		return r.dbWithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", challengeID).
			First(&model.Challenge{}).Error
	case instancecontracts.ShareScopePerTeam:
		if scope.TeamID != nil {
			return r.dbWithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", *scope.TeamID).
				First(&contestTeamRow{}).Error
		}
	}
	if scope.TeamID != nil && scope.ShareScope == instancecontracts.ShareScopePerTeam {
		return r.dbWithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", *scope.TeamID).
			First(&contestTeamRow{}).Error
	}
	return r.dbWithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", userID).
		First(&model.User{}).Error
}

func (r *Repository) FindScopedExistingInstance(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*instancecontracts.Instance, error) {
	now := time.Now().UTC()
	query := r.scopedInstanceQuery(ctx, userID, challengeID, scope).
		Where("share_scope = ?", scope.ShareScope).
		Where(
			"(status IN ? OR (status = ? AND expires_at > ?))",
			[]string{instancecontracts.InstanceStatusPending, instancecontracts.InstanceStatusCreating},
			instancecontracts.InstanceStatusRunning,
			now,
		)

	var instance instancecontracts.Instance
	if err := query.Order("created_at DESC, id DESC").First(&instance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) FindScopedRestartableInstance(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*instancecontracts.Instance, error) {
	query := r.scopedInstanceQuery(ctx, userID, challengeID, scope).
		Where("share_scope = ?", scope.ShareScope).
		Where("status IN ?", []string{
			instancecontracts.InstanceStatusPending,
			instancecontracts.InstanceStatusCreating,
			instancecontracts.InstanceStatusRunning,
			instancecontracts.InstanceStatusFailed,
		})

	var instance instancecontracts.Instance
	if err := query.Order("created_at DESC, id DESC").First(&instance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) scopedInstanceQuery(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) *gorm.DB {
	query := r.dbWithContext(ctx).Model(&instancecontracts.Instance{})
	if scope.ServiceID != nil {
		query = query.Where("service_id = ?", *scope.ServiceID)
	} else {
		query = query.Where("challenge_id = ?", challengeID)
	}

	switch {
	case scope.ShareScope == instancecontracts.ShareScopeShared && scope.ContestID != nil:
		query = query.Where("contest_id = ? AND team_id IS NULL", *scope.ContestID)
	case scope.ShareScope == instancecontracts.ShareScopeShared:
		query = query.Where("contest_id IS NULL AND team_id IS NULL")
	case scope.TeamID != nil && scope.ContestID != nil:
		query = query.Where("contest_id = ? AND team_id = ?", *scope.ContestID, *scope.TeamID)
	case scope.ContestID != nil:
		query = query.Where("contest_id = ? AND user_id = ? AND team_id IS NULL", *scope.ContestID, userID)
	default:
		query = query.Where("user_id = ? AND contest_id IS NULL AND team_id IS NULL", userID)
	}
	return query
}

func (r *Repository) CountScopedRunningInstances(ctx context.Context, userID int64, scope practiceports.InstanceScope) (int, error) {
	now := time.Now().UTC()
	query := r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("share_scope = ?", scope.ShareScope).
		Where(
			"(status IN ? OR (status = ? AND expires_at > ?))",
			[]string{instancecontracts.InstanceStatusPending, instancecontracts.InstanceStatusCreating},
			instancecontracts.InstanceStatusRunning,
			now,
		)

	switch {
	case scope.ShareScope == instancecontracts.ShareScopeShared && scope.ContestID != nil:
		query = query.Where("contest_id = ? AND team_id IS NULL", *scope.ContestID)
	case scope.ShareScope == instancecontracts.ShareScopeShared:
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
	return r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("id = ?", instanceID).
		Updates(map[string]any{
			"expires_at": expiresAt,
			"updated_at": time.Now().UTC(),
		}).Error
}

func (r *Repository) ResetInstanceRuntimeForRestart(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
	if instanceID <= 0 {
		return nil
	}
	if expiresAt.IsZero() {
		expiresAt = time.Now().UTC()
	}

	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var instance instancecontracts.Instance
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Select("id", "host_port").
			Where("id = ?", instanceID).
			First(&instance).Error; err != nil {
			return err
		}

		repoWithTx := r.WithDB(tx)
		hostPort, err := repoWithTx.runtimePorts.SyncInstanceHostPortForRestart(ctx, instance.ID, instance.HostPort, preserveHostPort)
		if err != nil {
			return err
		}
		instance.HostPort = hostPort

		updates := map[string]any{
			"container_id":    "",
			"network_id":      "",
			"runtime_details": "",
			"access_url":      "",
			"status":          status,
			"expires_at":      expiresAt,
			"destroyed_at":    nil,
			"updated_at":      time.Now().UTC(),
		}
		if preserveHostPort && instance.HostPort > 0 {
			updates["host_port"] = instance.HostPort
		}
		if !preserveHostPort {
			updates["host_port"] = 0
		}
		return tx.Model(&instancecontracts.Instance{}).
			Where("id = ?", instanceID).
			Updates(updates).Error
	})
}

func (r *Repository) IsHostPortReusableForRestart(ctx context.Context, instanceID int64, hostPort int) (bool, error) {
	return r.runtimePorts.IsHostPortReusableForRestart(ctx, instanceID, hostPort)
}

func (r *Repository) CreateInstance(ctx context.Context, instance *instancecontracts.Instance) error {
	return r.dbWithContext(ctx).Create(instance).Error
}

func (r *Repository) CreateAWDServiceOperation(ctx context.Context, operation *runtimecontracts.AWDServiceOperation) error {
	return r.dbWithContext(ctx).Create(operation).Error
}

func (r *Repository) FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error {
	if instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Model(&runtimecontracts.AWDServiceOperation{}).
		Where("instance_id = ? AND status IN ?", instanceID, []string{
			runtimecontracts.AWDServiceOperationStatusRequested,
			runtimecontracts.AWDServiceOperationStatusProvisioning,
			runtimecontracts.AWDServiceOperationStatusRecovering,
		}).
		Updates(map[string]any{
			"status":        status,
			"error_message": errorMessage,
			"finished_at":   finishedAt,
			"updated_at":    time.Now().UTC(),
		}).Error
}

func (r *Repository) FinishAWDServiceOperation(ctx context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error {
	if operationID <= 0 {
		return nil
	}
	updates := map[string]any{
		"status":        status,
		"error_message": errorMessage,
		"finished_at":   finishedAt,
		"updated_at":    time.Now().UTC(),
	}
	return r.dbWithContext(ctx).
		Model(&runtimecontracts.AWDServiceOperation{}).
		Where("id = ?", operationID).
		Updates(updates).Error
}

func (r *Repository) ReserveAvailablePort(ctx context.Context, start, end int) (int, error) {
	return r.runtimePorts.ReserveAvailablePort(ctx, start, end)
}

func (r *Repository) ReserveAvailablePortExcluding(ctx context.Context, start, end, excludedPort int) (int, error) {
	return r.runtimePorts.ReserveAvailablePortExcluding(ctx, start, end, excludedPort)
}

func (r *Repository) BindReservedPort(ctx context.Context, port int, instanceID int64) error {
	return r.runtimePorts.BindReservedPort(ctx, port, instanceID)
}

func (r *Repository) ReleaseReservedPort(ctx context.Context, port int) error {
	return r.runtimePorts.ReleaseReservedPort(ctx, port)
}

func (r *Repository) ReleasePortForInstance(ctx context.Context, port int, instanceID int64) error {
	return r.runtimePorts.ReleasePortForInstance(ctx, port, instanceID)
}

// CreateSubmission 创建提交记录
func (r *Repository) CreateSubmission(ctx context.Context, submission *practiceports.SubmissionRecord) error {
	row := submissionRowFromRecord(submission)
	if row == nil {
		return nil
	}
	if err := r.dbWithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	submission.ID = row.ID
	return nil
}

// FindCorrectSubmission 查找用户是否已正确提交过该题
func (r *Repository) FindCorrectSubmission(ctx context.Context, userID, challengeID int64) (*practiceports.SubmissionRecord, error) {
	var submission submissionRow
	err := r.dbWithContext(ctx).Where("user_id = ? AND challenge_id = ? AND is_correct = ?", userID, challengeID, true).
		First(&submission).Error
	return submission.toRecord(), err
}

func (r *Repository) FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*instancecontracts.Instance, error) {
	var instance instancecontracts.Instance
	err := r.dbWithContext(ctx).
		Where("user_id = ? AND contest_id IS NULL AND team_id IS NULL AND challenge_id = ? AND status IN ?", userID, challengeID,
			[]string{instancecontracts.InstanceStatusPending, instancecontracts.InstanceStatusCreating, instancecontracts.InstanceStatusRunning}).
		First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) ListChallengeSubmissions(ctx context.Context, userID, challengeID int64, limit int) ([]practiceports.SubmissionRecord, error) {
	if limit <= 0 {
		limit = 20
	}

	var submissions []submissionRow
	err := r.dbWithContext(ctx).
		Where("user_id = ? AND challenge_id = ? AND contest_id IS NULL", userID, challengeID).
		Order("submitted_at DESC, id DESC").
		Limit(limit).
		Find(&submissions).Error
	return submissionRowsToRecords(submissions), err
}

func (r *Repository) UpdateSubmission(ctx context.Context, submission *practiceports.SubmissionRecord) error {
	row := submissionRowFromRecord(submission)
	if row == nil {
		return nil
	}
	return r.dbWithContext(ctx).Save(row).Error
}

func (r *Repository) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
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
		Submission: practiceports.SubmissionRecord{
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

func (r *Repository) GetTeacherManualReviewSubmissionByID(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
	rows, _, err := r.listTeacherManualReviewSubmissions(ctx, &practicecontracts.TeacherManualReviewSubmissionQuery{
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

func (r *Repository) ListTeacherManualReviewSubmissions(ctx context.Context, query *practicecontracts.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
	return r.listTeacherManualReviewSubmissions(ctx, query, nil)
}

func (r *Repository) listTeacherManualReviewSubmissions(
	ctx context.Context,
	query *practicecontracts.TeacherManualReviewSubmissionQuery,
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
func (r *Repository) CountRecentSubmissions(ctx context.Context, userID, challengeID int64, since time.Time) (int64, error) {
	var count int64
	err := r.dbWithContext(ctx).Model(&submissionRow{}).
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
