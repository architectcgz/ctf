package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ctf-platform/internal/model"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

type Repository struct {
	db *gorm.DB
}

type userVisibleInstanceRow struct {
	ID              int64            `gorm:"column:id"`
	ContestMode     string           `gorm:"column:contest_mode"`
	ChallengeID     int64            `gorm:"column:challenge_id"`
	ChallengeTitle  string           `gorm:"column:challenge_title"`
	Category        string           `gorm:"column:category"`
	Difficulty      string           `gorm:"column:difficulty"`
	FlagType        string           `gorm:"column:flag_type"`
	ServiceName     string           `gorm:"column:service_name"`
	ServiceSnapshot string           `gorm:"column:service_snapshot"`
	Status          string           `gorm:"column:status"`
	ShareScope      model.ShareScope `gorm:"column:share_scope"`
	AccessURL       string           `gorm:"column:access_url"`
	ExpiresAt       time.Time        `gorm:"column:expires_at"`
	ExtendCount     int              `gorm:"column:extend_count"`
	MaxExtends      int              `gorm:"column:max_extends"`
	CreatedAt       time.Time        `gorm:"column:created_at"`
}

type teacherInstanceRow struct {
	ID              int64     `gorm:"column:id"`
	StudentID       int64     `gorm:"column:student_id"`
	StudentName     string    `gorm:"column:student_name"`
	StudentUsername string    `gorm:"column:student_username"`
	StudentNo       *string   `gorm:"column:student_no"`
	ClassName       string    `gorm:"column:class_name"`
	ContestMode     string    `gorm:"column:contest_mode"`
	ChallengeID     int64     `gorm:"column:challenge_id"`
	ChallengeTitle  string    `gorm:"column:challenge_title"`
	ServiceName     string    `gorm:"column:service_name"`
	ServiceSnapshot string    `gorm:"column:service_snapshot"`
	Status          string    `gorm:"column:status"`
	AccessURL       string    `gorm:"column:access_url"`
	ExpiresAt       time.Time `gorm:"column:expires_at"`
	ExtendCount     int       `gorm:"column:extend_count"`
	MaxExtends      int       `gorm:"column:max_extends"`
	CreatedAt       time.Time `gorm:"column:created_at"`
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithDB(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*model.Instance, error) {
	var instance model.Instance
	err := r.dbWithContext(ctx).Where("id = ?", id).First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

func (r *Repository) FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	var challenge model.Challenge
	if err := r.dbWithContext(ctx).Where("id = ?", challengeID).First(&challenge).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (r *Repository) FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*model.Instance, error) {
	var instance model.Instance
	err := r.dbWithContext(ctx).Where("user_id = ? AND contest_id IS NULL AND team_id IS NULL AND challenge_id = ? AND status IN ?", userID, challengeID,
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

func (r *Repository) FindByContestUserID(ctx context.Context, contestID, userID int64) ([]*model.Instance, error) {
	var instances []*model.Instance
	err := r.dbWithContext(ctx).Where("contest_id = ? AND user_id = ? AND team_id IS NULL AND status IN ?", contestID, userID,
		[]string{model.InstanceStatusPending, model.InstanceStatusCreating, model.InstanceStatusRunning}).
		Order("created_at DESC").
		Find(&instances).Error
	return instances, err
}

func (r *Repository) FindByContestUserAndChallenge(ctx context.Context, contestID, userID, challengeID int64) (*model.Instance, error) {
	var instance model.Instance
	err := r.dbWithContext(ctx).Where("contest_id = ? AND user_id = ? AND team_id IS NULL AND challenge_id = ? AND status IN ?",
		contestID, userID, challengeID, []string{model.InstanceStatusPending, model.InstanceStatusCreating, model.InstanceStatusRunning}).
		First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) FindByContestTeamID(ctx context.Context, contestID, teamID int64) ([]*model.Instance, error) {
	var instances []*model.Instance
	err := r.dbWithContext(ctx).Where("contest_id = ? AND team_id = ? AND status IN ?", contestID, teamID,
		[]string{model.InstanceStatusPending, model.InstanceStatusCreating, model.InstanceStatusRunning}).
		Order("created_at DESC").
		Find(&instances).Error
	return instances, err
}

func (r *Repository) FindByContestTeamAndChallenge(ctx context.Context, contestID, teamID, challengeID int64) (*model.Instance, error) {
	var instance model.Instance
	err := r.dbWithContext(ctx).Where("contest_id = ? AND team_id = ? AND challenge_id = ? AND status IN ?",
		contestID, teamID, challengeID, []string{model.InstanceStatusPending, model.InstanceStatusCreating, model.InstanceStatusRunning}).
		First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error {
	return r.dbWithContext(ctx).Model(&model.Instance{}).
		Where("id = ?", instanceID).
		Updates(map[string]any{
			"expires_at": expiresAt,
			"updated_at": time.Now(),
		}).Error
}

func (r *Repository) UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error {
	if id <= 0 {
		return nil
	}

	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var instance model.Instance
		if err := tx.Select("id", "host_port").Where("id = ?", id).First(&instance).Error; err != nil {
			return err
		}

		updates := map[string]any{
			"status":     status,
			"updated_at": time.Now(),
		}
		if status == model.InstanceStatusStopped || status == model.InstanceStatusExpired {
			updates["destroyed_at"] = time.Now()
		}
		if err := tx.Model(&model.Instance{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
			return err
		}

		deleteQuery := tx.Where("instance_id = ?", id)
		if instance.HostPort > 0 {
			deleteQuery = deleteQuery.Or("port = ?", instance.HostPort)
		}
		if err := deleteQuery.Delete(&model.PortAllocation{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) UpdateRuntime(ctx context.Context, instance *model.Instance) error {
	return r.dbWithContext(ctx).Model(&model.Instance{}).
		Where("id = ?", instance.ID).
		Updates(map[string]any{
			"contest_id":      instance.ContestID,
			"team_id":         instance.TeamID,
			"host_port":       instance.HostPort,
			"container_id":    instance.ContainerID,
			"network_id":      instance.NetworkID,
			"runtime_details": instance.RuntimeDetails,
			"access_url":      instance.AccessURL,
			"status":          instance.Status,
			"updated_at":      time.Now(),
		}).Error
}

func (r *Repository) FindAWDDefenseWorkspace(ctx context.Context, contestID, teamID, serviceID int64) (*model.AWDDefenseWorkspace, error) {
	var workspace model.AWDDefenseWorkspace
	err := r.dbWithContext(ctx).
		Where("contest_id = ? AND team_id = ? AND service_id = ?", contestID, teamID, serviceID).
		First(&workspace).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &workspace, nil
}

func (r *Repository) UpsertAWDDefenseWorkspace(ctx context.Context, workspace *model.AWDDefenseWorkspace) error {
	if workspace == nil {
		return nil
	}

	if workspace.WorkspaceRevision <= 0 {
		workspace.WorkspaceRevision = 1
	}
	if strings.TrimSpace(workspace.Status) == "" {
		workspace.Status = model.AWDDefenseWorkspaceStatusPending
	}

	now := time.Now()
	if err := r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "contest_id"},
			{Name: "team_id"},
			{Name: "service_id"},
		},
		DoUpdates: clause.Assignments(map[string]any{
			"instance_id":        workspace.InstanceID,
			"workspace_revision": workspace.WorkspaceRevision,
			"status":             workspace.Status,
			"container_id":       workspace.ContainerID,
			"seed_signature":     workspace.SeedSignature,
			"updated_at":         now,
		}),
	}).Create(workspace).Error; err != nil {
		return err
	}

	stored, err := r.FindAWDDefenseWorkspace(ctx, workspace.ContestID, workspace.TeamID, workspace.ServiceID)
	if err != nil {
		return err
	}
	if stored != nil {
		*workspace = *stored
	}
	return nil
}

func (r *Repository) BumpAWDDefenseWorkspaceRevision(ctx context.Context, contestID, teamID, serviceID, instanceID int64, seedSignature string) error {
	now := time.Now()
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var workspace model.AWDDefenseWorkspace
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("contest_id = ? AND team_id = ? AND service_id = ?", contestID, teamID, serviceID).
			First(&workspace).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			return tx.Create(&model.AWDDefenseWorkspace{
				ContestID:         contestID,
				TeamID:            teamID,
				ServiceID:         serviceID,
				InstanceID:        instanceID,
				WorkspaceRevision: 1,
				Status:            model.AWDDefenseWorkspaceStatusProvisioning,
				SeedSignature:     seedSignature,
				CreatedAt:         now,
				UpdatedAt:         now,
			}).Error
		}

		return tx.Model(&model.AWDDefenseWorkspace{}).
			Where("id = ?", workspace.ID).
			Updates(map[string]any{
				"instance_id":        instanceID,
				"workspace_revision": workspace.WorkspaceRevision + 1,
				"status":             model.AWDDefenseWorkspaceStatusProvisioning,
				"container_id":       "",
				"seed_signature":     seedSignature,
				"updated_at":         now,
			}).Error
	})
}

func (r *Repository) FindAccessibleByIDForUser(ctx context.Context, instanceID, userID int64) (*model.Instance, error) {
	var instance model.Instance
	err := r.db.WithContext(ctx).
		Table("instances AS inst").
		Select("inst.*").
		Joins("LEFT JOIN team_members AS tm ON tm.team_id = inst.team_id AND tm.contest_id = inst.contest_id AND tm.user_id = ?", userID).
		Joins("LEFT JOIN contest_registrations AS reg ON reg.contest_id = inst.contest_id AND reg.user_id = ? AND reg.status = ?", userID, model.ContestRegistrationStatusApproved).
		Where("inst.id = ?", instanceID).
		Where(strings.Join([]string{
			"(inst.share_scope = 'shared' AND inst.contest_id IS NULL)",
			"(inst.share_scope = 'shared' AND inst.contest_id IS NOT NULL AND reg.user_id IS NOT NULL)",
			"(inst.share_scope <> 'shared' AND inst.team_id IS NULL AND inst.user_id = ?)",
			"(inst.team_id IS NOT NULL AND tm.user_id IS NOT NULL)",
		}, " OR "), userID).
		First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) FindVisibleByUser(ctx context.Context, userID int64) ([]*model.Instance, error) {
	var instances []*model.Instance
	err := r.db.WithContext(ctx).
		Table("instances AS inst").
		Select("DISTINCT inst.*").
		Joins("LEFT JOIN team_members AS tm ON tm.team_id = inst.team_id AND tm.contest_id = inst.contest_id AND tm.user_id = ?", userID).
		Joins("LEFT JOIN contest_registrations AS reg ON reg.contest_id = inst.contest_id AND reg.user_id = ? AND reg.status = ?", userID, model.ContestRegistrationStatusApproved).
		Where("inst.status IN ?", []string{model.InstanceStatusPending, model.InstanceStatusCreating, model.InstanceStatusRunning, model.InstanceStatusFailed, model.InstanceStatusExpired}).
		Where(strings.Join([]string{
			"(inst.share_scope = 'shared' AND inst.contest_id IS NULL)",
			"(inst.share_scope = 'shared' AND inst.contest_id IS NOT NULL AND reg.user_id IS NOT NULL)",
			"(inst.share_scope <> 'shared' AND inst.team_id IS NULL AND inst.user_id = ?)",
			"(inst.team_id IS NOT NULL AND tm.user_id IS NOT NULL)",
		}, " OR "), userID).
		Order("inst.created_at DESC").
		Scan(&instances).Error
	return instances, err
}

func (r *Repository) ListVisibleByUser(ctx context.Context, userID int64) ([]runtimeports.UserVisibleInstanceRow, error) {
	rows := make([]userVisibleInstanceRow, 0)
	err := r.db.WithContext(ctx).
		Table("instances AS inst").
		Select(strings.Join([]string{
			"inst.id",
			"COALESCE(co.mode, '') AS contest_mode",
			"CASE WHEN co.mode = 'awd' THEN cas.awd_challenge_id ELSE inst.challenge_id END AS challenge_id",
			"c.title AS challenge_title",
			"c.category",
			"c.difficulty",
			"c.flag_type",
			"cas.display_name AS service_name",
			"cas.service_snapshot AS service_snapshot",
			"inst.status",
			"inst.share_scope",
			"inst.access_url",
			"inst.expires_at",
			"inst.extend_count",
			"inst.max_extends",
			"inst.created_at",
		}, ", ")).
		Joins("LEFT JOIN contests co ON co.id = inst.contest_id").
		Joins("LEFT JOIN contest_awd_services AS cas ON cas.id = inst.service_id AND cas.deleted_at IS NULL").
		Joins("LEFT JOIN challenges c ON c.id = inst.challenge_id").
		Joins("LEFT JOIN team_members AS tm ON tm.team_id = inst.team_id AND tm.contest_id = inst.contest_id AND tm.user_id = ?", userID).
		Joins("LEFT JOIN contest_registrations AS reg ON reg.contest_id = inst.contest_id AND reg.user_id = ? AND reg.status = ?", userID, model.ContestRegistrationStatusApproved).
		Where("inst.status IN ?", []string{model.InstanceStatusPending, model.InstanceStatusCreating, model.InstanceStatusRunning, model.InstanceStatusFailed, model.InstanceStatusExpired}).
		Where("(co.mode IS NULL OR co.mode <> ? OR cas.id IS NOT NULL)", model.ContestModeAWD).
		Where(strings.Join([]string{
			"(inst.share_scope = 'shared' AND inst.contest_id IS NULL)",
			"(inst.share_scope = 'shared' AND inst.contest_id IS NOT NULL AND reg.user_id IS NOT NULL)",
			"(inst.share_scope <> 'shared' AND inst.team_id IS NULL AND inst.user_id = ?)",
			"(inst.team_id IS NOT NULL AND tm.user_id IS NOT NULL)",
		}, " OR "), userID).
		Order("inst.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	items := make([]runtimeports.UserVisibleInstanceRow, len(rows))
	for idx, row := range rows {
		metadata := buildRuntimeInstanceMetadata(row.ContestMode, row.ServiceSnapshot, row.ServiceName, row.ChallengeTitle, row.Category, row.Difficulty, row.FlagType)
		items[idx] = runtimeports.UserVisibleInstanceRow{
			ID:             row.ID,
			ContestMode:    row.ContestMode,
			ChallengeID:    row.ChallengeID,
			ChallengeTitle: metadata.Title,
			Category:       metadata.Category,
			Difficulty:     metadata.Difficulty,
			FlagType:       metadata.FlagType,
			Status:         row.Status,
			ShareScope:     row.ShareScope,
			AccessURL:      row.AccessURL,
			ExpiresAt:      row.ExpiresAt,
			ExtendCount:    row.ExtendCount,
			MaxExtends:     row.MaxExtends,
			CreatedAt:      row.CreatedAt,
		}
	}
	return items, nil
}

func (r *Repository) FindExpired(ctx context.Context) ([]*model.Instance, error) {
	var instances []*model.Instance
	err := r.dbWithContext(ctx).Where("status = ? AND expires_at < ?",
		model.InstanceStatusRunning, time.Now()).
		Find(&instances).Error
	return instances, err
}

func (r *Repository) ListRecoverableActiveInstances(ctx context.Context) ([]*model.Instance, error) {
	var instances []*model.Instance
	err := r.dbWithContext(ctx).
		Where("status IN ?", []string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
		Where("expires_at > ?", time.Now()).
		Order("updated_at ASC, id ASC").
		Find(&instances).Error
	return instances, err
}

func (r *Repository) RequeueLostRuntime(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, nil
	}

	result := r.dbWithContext(ctx).Model(&model.Instance{}).
		Where("id = ? AND status IN ? AND expires_at > ?",
			id,
			[]string{model.InstanceStatusCreating, model.InstanceStatusRunning},
			time.Now(),
		).
		Updates(map[string]any{
			"status":          model.InstanceStatusPending,
			"container_id":    "",
			"network_id":      "",
			"runtime_details": "",
			"access_url":      "",
			"updated_at":      time.Now(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) CreateAWDServiceOperation(ctx context.Context, operation *model.AWDServiceOperation) error {
	return r.dbWithContext(ctx).Create(operation).Error
}

func (r *Repository) FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error {
	if instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Model(&model.AWDServiceOperation{}).
		Where("instance_id = ? AND status IN ?", instanceID, []string{
			model.AWDServiceOperationStatusRequested,
			model.AWDServiceOperationStatusProvisioning,
			model.AWDServiceOperationStatusRecovering,
		}).
		Updates(map[string]any{
			"status":        status,
			"error_message": errorMessage,
			"finished_at":   finishedAt,
			"updated_at":    time.Now(),
		}).Error
}

func (r *Repository) FinishAWDServiceOperation(ctx context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error {
	if operationID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Model(&model.AWDServiceOperation{}).
		Where("id = ?", operationID).
		Updates(map[string]any{
			"status":        status,
			"error_message": errorMessage,
			"finished_at":   finishedAt,
			"updated_at":    time.Now(),
		}).Error
}

func (r *Repository) ListTeacherInstances(ctx context.Context, filter runtimeports.TeacherInstanceFilter) ([]runtimeports.TeacherInstanceRow, error) {
	rows := make([]teacherInstanceRow, 0)

	query := r.db.WithContext(ctx).
		Table("instances AS i").
		Select(strings.Join([]string{
			"i.id",
			"u.id AS student_id",
			"u.username AS student_name",
			"u.username AS student_username",
			"NULLIF(u.student_no, '') AS student_no",
			"u.class_name",
			"COALESCE(co.mode, '') AS contest_mode",
			"CASE WHEN co.mode = 'awd' THEN cas.awd_challenge_id ELSE i.challenge_id END AS challenge_id",
			"c.title AS challenge_title",
			"cas.display_name AS service_name",
			"cas.service_snapshot AS service_snapshot",
			"i.status",
			"i.access_url",
			"i.expires_at",
			"i.extend_count",
			"i.max_extends",
			"i.created_at",
		}, ", ")).
		Joins("JOIN users u ON u.id = i.user_id").
		Joins("LEFT JOIN contests co ON co.id = i.contest_id").
		Joins("LEFT JOIN contest_awd_services AS cas ON cas.id = i.service_id AND cas.deleted_at IS NULL").
		Joins("LEFT JOIN challenges c ON c.id = i.challenge_id").
		Where("i.status <> ?", model.InstanceStatusStopped).
		Where("(co.mode IS NULL OR co.mode <> ? OR cas.id IS NOT NULL)", model.ContestModeAWD).
		Where("u.role = ? AND u.deleted_at IS NULL", model.RoleStudent)

	if filter.ClassName != "" {
		query = query.Where("u.class_name = ?", filter.ClassName)
	}
	if filter.StudentNo != "" {
		query = query.Where("u.student_no = ?", filter.StudentNo)
	}
	if filter.Keyword != "" {
		pattern := "%" + strings.ToLower(filter.Keyword) + "%"
		query = query.Where(
			"(LOWER(u.username) LIKE ? OR LOWER(COALESCE(NULLIF(u.student_no, ''), '')) LIKE ?)",
			pattern,
			pattern,
		)
	}

	if err := query.Order("i.created_at DESC").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list teacher instances: %w", err)
	}

	items := make([]runtimeports.TeacherInstanceRow, len(rows))
	for idx, row := range rows {
		metadata := buildRuntimeInstanceMetadata(row.ContestMode, row.ServiceSnapshot, row.ServiceName, row.ChallengeTitle, "", "", "")
		items[idx] = runtimeports.TeacherInstanceRow{
			ID:              row.ID,
			StudentID:       row.StudentID,
			StudentName:     row.StudentName,
			StudentUsername: row.StudentUsername,
			StudentNo:       row.StudentNo,
			ClassName:       row.ClassName,
			ChallengeID:     row.ChallengeID,
			ChallengeTitle:  metadata.Title,
			Status:          row.Status,
			AccessURL:       row.AccessURL,
			ExpiresAt:       row.ExpiresAt,
			ExtendCount:     row.ExtendCount,
			MaxExtends:      row.MaxExtends,
			CreatedAt:       row.CreatedAt,
		}
	}
	return items, nil
}

type runtimeInstanceMetadata struct {
	Title      string
	Category   string
	Difficulty string
	FlagType   string
}

func buildRuntimeInstanceMetadata(contestMode, serviceSnapshot, serviceName, challengeTitle, category, difficulty, flagType string) runtimeInstanceMetadata {
	metadata := runtimeInstanceMetadata{
		Title:      challengeTitle,
		Category:   category,
		Difficulty: difficulty,
		FlagType:   flagType,
	}
	if contestMode != model.ContestModeAWD {
		return metadata
	}

	snapshot, err := model.DecodeContestAWDServiceSnapshot(serviceSnapshot)
	if err != nil {
		return metadata
	}
	if title := strings.TrimSpace(snapshot.Name); title != "" {
		metadata.Title = title
	} else if title := strings.TrimSpace(serviceName); title != "" {
		metadata.Title = title
	}
	if value := strings.TrimSpace(snapshot.Category); value != "" {
		metadata.Category = value
	}
	if value := strings.TrimSpace(snapshot.Difficulty); value != "" {
		metadata.Difficulty = value
	}
	if snapshot.FlagConfig != nil {
		if value, ok := snapshot.FlagConfig["flag_type"].(string); ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				metadata.FlagType = trimmed
			}
		}
	}
	return metadata
}

func (r *Repository) UpdateExtend(ctx context.Context, id int64, expiresAt time.Time, extendCount int) error {
	return r.dbWithContext(ctx).Model(&model.Instance{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"expires_at":   expiresAt,
			"extend_count": extendCount,
		}).Error
}

func (r *Repository) AtomicExtend(ctx context.Context, id int64, userID int64, maxExtends int, duration time.Duration) error {
	result := r.dbWithContext(ctx).Model(&model.Instance{}).
		Where("id = ? AND user_id = ? AND status = ? AND extend_count < ?",
			id, userID, model.InstanceStatusRunning, maxExtends).
		Updates(map[string]interface{}{
			"expires_at":   gorm.Expr("expires_at + ?", duration),
			"extend_count": gorm.Expr("extend_count + 1"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errcode.ErrExtendLimitExceeded
	}
	return nil
}

func (r *Repository) AtomicExtendByID(ctx context.Context, id int64, maxExtends int, duration time.Duration) error {
	result := r.db.WithContext(ctx).Model(&model.Instance{}).
		Where("id = ? AND status = ? AND extend_count < ?",
			id, model.InstanceStatusRunning, maxExtends).
		Updates(map[string]interface{}{
			"expires_at":   gorm.Expr("expires_at + ?", duration),
			"extend_count": gorm.Expr("extend_count + 1"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errcode.ErrExtendLimitExceeded
	}
	return nil
}

func (r *Repository) CountRunning(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Instance{}).
		Where("status = ?", model.InstanceStatusRunning).
		Count(&count).Error
	return count, err
}

func (r *Repository) ListPendingInstances(ctx context.Context, limit int) ([]*model.Instance, error) {
	if limit <= 0 {
		return []*model.Instance{}, nil
	}

	instances := make([]*model.Instance, 0, limit)
	err := r.db.WithContext(ctx).
		Where("status = ?", model.InstanceStatusPending).
		Order("created_at ASC, id ASC").
		Limit(limit).
		Find(&instances).Error
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (r *Repository) TryTransitionStatus(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error) {
	result := r.db.WithContext(ctx).Model(&model.Instance{}).
		Where("id = ? AND status = ?", id, fromStatus).
		Updates(map[string]any{
			"status":     toStatus,
			"updated_at": time.Now(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) CountInstancesByStatus(ctx context.Context, statuses []string) (int64, error) {
	if len(statuses) == 0 {
		return 0, nil
	}

	var count int64
	err := r.db.WithContext(ctx).Model(&model.Instance{}).
		Where("status IN ?", statuses).
		Count(&count).Error
	return count, err
}

func (r *Repository) ReserveAvailablePort(ctx context.Context, start, end int) (int, error) {
	for port := start; port < end; port++ {
		if err := r.dbWithContext(ctx).Create(&model.PortAllocation{Port: port}).Error; err != nil {
			if isPortAllocationConflict(err) {
				continue
			}
			return 0, err
		}
		return port, nil
	}
	return 0, fmt.Errorf("no available port in range %d-%d", start, end)
}

func (r *Repository) BindReservedPort(ctx context.Context, port int, instanceID int64) error {
	return r.dbWithContext(ctx).Model(&model.PortAllocation{}).
		Where("port = ?", port).
		Updates(map[string]any{
			"instance_id": instanceID,
			"updated_at":  time.Now(),
		}).Error
}

func (r *Repository) ReleasePort(ctx context.Context, port int) error {
	if port <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).Where("port = ?", port).Delete(&model.PortAllocation{}).Error
}

func (r *Repository) ListActiveContainerIDs(ctx context.Context) ([]string, error) {
	var items []struct {
		ContainerID    string
		RuntimeDetails string
	}
	if err := r.dbWithContext(ctx).Model(&model.Instance{}).
		Where("status IN ?", []string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
		Select("container_id, runtime_details").
		Scan(&items).Error; err != nil {
		return nil, err
	}
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		ids := []string{item.ContainerID}
		details, err := model.DecodeInstanceRuntimeDetails(item.RuntimeDetails)
		if err == nil {
			for _, container := range details.Containers {
				ids = append(ids, container.ContainerID)
			}
		}
		for _, containerID := range ids {
			if containerID == "" {
				continue
			}
			if _, exists := seen[containerID]; exists {
				continue
			}
			seen[containerID] = struct{}{}
			result = append(result, containerID)
		}
	}
	return result, nil
}

func (r *Repository) ListAllocatedPorts(ctx context.Context) ([]int, error) {
	var ports []int
	if err := r.dbWithContext(ctx).Model(&model.PortAllocation{}).Pluck("port", &ports).Error; err == nil {
		return ports, nil
	} else if !strings.Contains(strings.ToLower(err.Error()), "no such table") && !strings.Contains(strings.ToLower(err.Error()), "does not exist") {
		return nil, err
	}

	var accessURLs []string
	if err := r.dbWithContext(ctx).Model(&model.Instance{}).
		Where("status IN ?", []string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
		Where("access_url <> ''").
		Pluck("access_url", &accessURLs).Error; err != nil {
		return nil, err
	}

	ports = make([]int, 0, len(accessURLs))
	for _, rawURL := range accessURLs {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			continue
		}
		portValue := parsed.Port()
		if portValue == "" {
			continue
		}
		port, err := strconv.Atoi(portValue)
		if err != nil {
			continue
		}
		ports = append(ports, port)
	}
	return ports, nil
}

func isPortAllocationConflict(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	lowered := strings.ToLower(err.Error())
	return strings.Contains(lowered, "unique constraint failed") ||
		strings.Contains(lowered, "duplicate key value") ||
		strings.Contains(lowered, "duplicate entry")
}
