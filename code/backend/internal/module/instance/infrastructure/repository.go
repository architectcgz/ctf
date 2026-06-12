package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceports "ctf-platform/internal/module/instance/ports"
)

type Repository struct {
	db *gorm.DB
}

type RuntimeAllocationRelease struct {
	InstanceID int64
	HostPort   int
}

type userVisibleInstanceRow struct {
	ID              int64                        `gorm:"column:id"`
	ContestMode     string                       `gorm:"column:contest_mode"`
	ChallengeID     int64                        `gorm:"column:challenge_id"`
	ChallengeTitle  string                       `gorm:"column:challenge_title"`
	Category        string                       `gorm:"column:category"`
	Difficulty      string                       `gorm:"column:difficulty"`
	FlagType        string                       `gorm:"column:flag_type"`
	ServiceName     string                       `gorm:"column:service_name"`
	ServiceSnapshot string                       `gorm:"column:service_snapshot"`
	Status          string                       `gorm:"column:status"`
	ShareScope      instancecontracts.ShareScope `gorm:"column:share_scope"`
	AccessURL       string                       `gorm:"column:access_url"`
	ExpiresAt       time.Time                    `gorm:"column:expires_at"`
	ExtendCount     int                          `gorm:"column:extend_count"`
	MaxExtends      int                          `gorm:"column:max_extends"`
	CreatedAt       time.Time                    `gorm:"column:created_at"`
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

type teacherInstanceSummaryRow struct {
	TotalCount        int64 `gorm:"column:total_count"`
	RunningCount      int64 `gorm:"column:running_count"`
	ExpiringSoonCount int64 `gorm:"column:expiring_soon_count"`
	WarningCount      int64 `gorm:"column:warning_count"`
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

func (r *Repository) FindByID(ctx context.Context, id int64) (*instancecontracts.Instance, error) {
	var instance instancecontracts.Instance
	err := r.dbWithContext(ctx).Where("id = ?", id).First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*instancecontracts.Instance, error) {
	var instance instancecontracts.Instance
	err := r.dbWithContext(ctx).
		Where("user_id = ? AND contest_id IS NULL AND team_id IS NULL AND challenge_id = ? AND status IN ?",
			userID,
			challengeID,
			[]string{
				instancecontracts.InstanceStatusPending,
				instancecontracts.InstanceStatusCreating,
				instancecontracts.InstanceStatusRunning,
			},
		).
		First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) FindUserByID(ctx context.Context, userID int64) (*instanceports.InstanceUser, error) {
	var user instanceports.InstanceUser
	if err := r.dbWithContext(ctx).
		Table("users").
		Select("id, role, class_name").
		Where("id = ?", userID).
		Where("deleted_at IS NULL").
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

func (r *Repository) FindAccessibleByIDForUser(ctx context.Context, instanceID, userID int64) (*instancecontracts.Instance, error) {
	var instance instancecontracts.Instance
	query := r.dbWithContext(ctx).
		Table("instances AS inst").
		Select("inst.*").
		Joins("LEFT JOIN team_members AS tm ON tm.team_id = inst.team_id AND tm.contest_id = inst.contest_id AND tm.user_id = ?", userID).
		Joins("LEFT JOIN contest_registrations AS reg ON reg.contest_id = inst.contest_id AND reg.user_id = ? AND reg.status = ?", userID, persistedContestRegistrationStatusApproved)
	query = joinAWDActiveScopeControls(query, "inst.contest_id", "inst.team_id", "inst.service_id", "inst_team_retired_ctl", "inst_service_disabled_ctl")
	err := applyAWDActiveScopeFilter(query, "inst.service_id", "inst_team_retired_ctl", "inst_service_disabled_ctl").
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

func (r *Repository) ListVisibleByUser(ctx context.Context, userID int64) ([]instanceports.UserVisibleInstanceRow, error) {
	rows := make([]userVisibleInstanceRow, 0)
	query := r.dbWithContext(ctx).
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
		Joins("LEFT JOIN contest_registrations AS reg ON reg.contest_id = inst.contest_id AND reg.user_id = ? AND reg.status = ?", userID, persistedContestRegistrationStatusApproved)
	query = joinAWDActiveScopeControls(query, "inst.contest_id", "inst.team_id", "inst.service_id", "list_team_retired_ctl", "list_service_disabled_ctl")
	err := applyAWDActiveScopeFilter(query, "inst.service_id", "list_team_retired_ctl", "list_service_disabled_ctl").
		Where("inst.status IN ?", []string{
			instancecontracts.InstanceStatusPending,
			instancecontracts.InstanceStatusCreating,
			instancecontracts.InstanceStatusRunning,
			instancecontracts.InstanceStatusStopping,
			instancecontracts.InstanceStatusFailed,
			instancecontracts.InstanceStatusExpired,
		}).
		Where("(co.mode IS NULL OR co.mode <> ? OR cas.id IS NOT NULL)", persistedContestModeAWD).
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

	items := make([]instanceports.UserVisibleInstanceRow, len(rows))
	for idx, row := range rows {
		metadata := buildInstanceMetadata(row.ContestMode, row.ServiceSnapshot, row.ServiceName, row.ChallengeTitle, row.Category, row.Difficulty, row.FlagType)
		items[idx] = instanceports.UserVisibleInstanceRow{
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

func (r *Repository) ListTeacherInstances(ctx context.Context, filter instanceports.TeacherInstanceFilter) (*instanceports.TeacherInstancePage, error) {
	rows := make([]teacherInstanceRow, 0)
	now := time.Now().UTC()

	query := r.dbWithContext(ctx).
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
		Where("i.status <> ?", instancecontracts.InstanceStatusStopped).
		Where("(co.mode IS NULL OR co.mode <> ? OR cas.id IS NOT NULL)", persistedContestModeAWD).
		Where("u.role = ? AND u.deleted_at IS NULL", instanceports.InstanceUserRoleStudent)

	query = applyTeacherInstanceQueryFilters(query, filter, now)

	var summary teacherInstanceSummaryRow
	summarySelect := strings.Join([]string{
		"COUNT(*) AS total_count",
		fmt.Sprintf(
			"COALESCE(SUM(CASE WHEN i.status = '%s' AND i.expires_at > ? THEN 1 ELSE 0 END), 0) AS running_count",
			instancecontracts.InstanceStatusRunning,
		),
		fmt.Sprintf(
			"COALESCE(SUM(CASE WHEN i.status = '%s' AND i.expires_at > ? AND i.expires_at <= ? THEN 1 ELSE 0 END), 0) AS expiring_soon_count",
			instancecontracts.InstanceStatusRunning,
		),
		fmt.Sprintf(
			"COALESCE(SUM(CASE WHEN i.status <> '%s' OR i.expires_at <= ? THEN 1 ELSE 0 END), 0) AS warning_count",
			instancecontracts.InstanceStatusRunning,
		),
	}, ", ")
	warningThreshold := now.Add(10 * time.Minute)
	if err := query.Session(&gorm.Session{}).
		Select(summarySelect, now, now, warningThreshold, warningThreshold).
		Scan(&summary).Error; err != nil {
		return nil, fmt.Errorf("summarize teacher instances: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	if err := query.Session(&gorm.Session{}).
		Order("i.created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list teacher instances: %w", err)
	}

	items := make([]instanceports.TeacherInstanceRow, len(rows))
	for idx, row := range rows {
		metadata := buildInstanceMetadata(row.ContestMode, row.ServiceSnapshot, row.ServiceName, row.ChallengeTitle, "", "", "")
		items[idx] = instanceports.TeacherInstanceRow{
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

	return &instanceports.TeacherInstancePage{
		List:  items,
		Total: summary.TotalCount,
		Summary: instanceports.TeacherInstanceListSummary{
			TotalCount:        summary.TotalCount,
			RunningCount:      summary.RunningCount,
			ExpiringSoonCount: summary.ExpiringSoonCount,
			WarningCount:      summary.WarningCount,
		},
	}, nil
}

func (r *Repository) AtomicExtendByID(ctx context.Context, id int64, maxExtends int, duration time.Duration) error {
	result := r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("id = ? AND status = ? AND extend_count < ?",
			id, instancecontracts.InstanceStatusRunning, maxExtends).
		Updates(map[string]any{
			"expires_at":   gorm.Expr("expires_at + ?", duration),
			"extend_count": gorm.Expr("extend_count + 1"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return instancecontracts.ErrExtendLimitExceeded
	}
	return nil
}

func (r *Repository) RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error {
	return r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("id = ?", instanceID).
		Updates(map[string]any{
			"expires_at": expiresAt,
			"updated_at": time.Now().UTC(),
		}).Error
}

func (r *Repository) UpdateRuntime(ctx context.Context, instance *instancecontracts.Instance) error {
	_, err := r.PersistProvisionedRuntime(ctx, instance)
	return err
}

func (r *Repository) BindRuntimeNode(ctx context.Context, id int64, nodeID *int64) (bool, error) {
	if id <= 0 {
		return false, nil
	}
	result := r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("id = ? AND status = ?", id, instancecontracts.InstanceStatusCreating).
		Updates(map[string]any{
			"node_id":    nodeID,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id int64, status string) (*RuntimeAllocationRelease, error) {
	release, _, err := r.updateStatusWithCurrentStatuses(ctx, id, nil, status)
	return release, err
}

func (r *Repository) FailProvisioning(ctx context.Context, id int64) (*RuntimeAllocationRelease, bool, error) {
	return r.updateStatusWithCurrentStatuses(
		ctx,
		id,
		[]string{instancecontracts.InstanceStatusCreating},
		instancecontracts.InstanceStatusFailed,
	)
}

func (r *Repository) PersistProvisionedRuntime(ctx context.Context, instance *instancecontracts.Instance) (bool, error) {
	if instance == nil || instance.ID <= 0 {
		return false, nil
	}
	result := r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("id = ? AND status = ?", instance.ID, instancecontracts.InstanceStatusCreating).
		Updates(map[string]any{
			"contest_id":      instance.ContestID,
			"team_id":         instance.TeamID,
			"node_id":         instance.NodeID,
			"host_port":       instance.HostPort,
			"container_id":    instance.ContainerID,
			"network_id":      instance.NetworkID,
			"runtime_details": instance.RuntimeDetails,
			"access_url":      instance.AccessURL,
			"status":          instance.Status,
			"updated_at":      time.Now().UTC(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) updateStatusWithCurrentStatuses(ctx context.Context, id int64, currentStatuses []string, status string) (*RuntimeAllocationRelease, bool, error) {
	if id <= 0 {
		return nil, false, nil
	}

	var instance instancecontracts.Instance
	query := r.dbWithContext(ctx).Select("id", "host_port").Where("id = ?", id)
	if len(currentStatuses) > 0 {
		query = query.Where("status IN ?", currentStatuses)
	}
	if err := query.First(&instance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) && len(currentStatuses) > 0 {
			return nil, false, nil
		}
		return nil, false, err
	}

	now := time.Now().UTC()
	updates := map[string]any{
		"status":     status,
		"updated_at": now,
	}
	if status == instancecontracts.InstanceStatusStopped || status == instancecontracts.InstanceStatusExpired {
		updates["destroyed_at"] = now
		updates["host_port"] = 0
		updates["container_id"] = ""
		updates["network_id"] = ""
		updates["runtime_details"] = ""
		updates["access_url"] = ""
	}
	updateQuery := r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("id = ?", id)
	if len(currentStatuses) > 0 {
		updateQuery = updateQuery.Where("status IN ?", currentStatuses)
	}
	result := updateQuery.Updates(updates)
	if result.Error != nil {
		return nil, false, result.Error
	}
	if len(currentStatuses) > 0 && result.RowsAffected == 0 {
		return nil, false, nil
	}

	return &RuntimeAllocationRelease{
		InstanceID: instance.ID,
		HostPort:   instance.HostPort,
	}, true, nil
}

func (r *Repository) MarkStopping(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, nil
	}

	result := r.dbWithContext(ctx).
		Model(&instancecontracts.Instance{}).
		Where("id = ? AND status IN ?", id, []string{
			instancecontracts.InstanceStatusPending,
			instancecontracts.InstanceStatusCreating,
			instancecontracts.InstanceStatusRunning,
			instancecontracts.InstanceStatusFailed,
		}).
		Updates(map[string]any{
			"status":     instancecontracts.InstanceStatusStopping,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) FindExpired(ctx context.Context) ([]*instancecontracts.Instance, error) {
	var instances []*instancecontracts.Instance
	err := r.dbWithContext(ctx).
		Where("status = ? AND expires_at < ?", instancecontracts.InstanceStatusRunning, time.Now().UTC()).
		Find(&instances).Error
	return instances, err
}

func (r *Repository) ListRecoverableActiveInstances(ctx context.Context) ([]*instancecontracts.Instance, error) {
	var instances []*instancecontracts.Instance
	err := r.dbWithContext(ctx).
		Where("status IN ?", []string{
			instancecontracts.InstanceStatusCreating,
			instancecontracts.InstanceStatusRunning,
		}).
		Where("expires_at > ?", time.Now().UTC()).
		Order("updated_at ASC, id ASC").
		Find(&instances).Error
	return instances, err
}

func (r *Repository) ListStoppingInstances(ctx context.Context, updatedBefore time.Time, limit int) ([]*instancecontracts.Instance, error) {
	var instances []*instancecontracts.Instance
	query := r.dbWithContext(ctx).
		Where("status = ?", instancecontracts.InstanceStatusStopping)
	if !updatedBefore.IsZero() {
		query = query.Where("updated_at <= ?", updatedBefore)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Order("updated_at ASC, id ASC").Find(&instances).Error
	return instances, err
}

func (r *Repository) RefreshActiveAWDInstanceExpiryByContest(ctx context.Context, contestID int64, activeAt, expiresAt time.Time) error {
	if contestID <= 0 || expiresAt.IsZero() {
		return nil
	}
	return r.dbWithContext(ctx).
		Model(&instancecontracts.Instance{}).
		Where("contest_id = ? AND service_id IS NOT NULL AND status IN ?", contestID, []string{
			instancecontracts.InstanceStatusPending,
			instancecontracts.InstanceStatusCreating,
			instancecontracts.InstanceStatusRunning,
		}).
		Where("expires_at > ?", activeAt.UTC()).
		Updates(map[string]any{
			"expires_at": expiresAt.UTC(),
			"updated_at": time.Now().UTC(),
		}).Error
}

func (r *Repository) RequeueLostRuntime(ctx context.Context, id int64) (bool, error) {
	if id <= 0 {
		return false, nil
	}

	result := r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("id = ? AND status IN ? AND expires_at > ?",
			id,
			[]string{
				instancecontracts.InstanceStatusCreating,
				instancecontracts.InstanceStatusRunning,
			},
			time.Now().UTC(),
		).
		Updates(map[string]any{
			"status":          instancecontracts.InstanceStatusPending,
			"node_id":         nil,
			"container_id":    "",
			"network_id":      "",
			"runtime_details": "",
			"access_url":      "",
			"updated_at":      time.Now().UTC(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) RequeueLostRuntimesByNode(ctx context.Context, nodeID int64) ([]*instancecontracts.Instance, error) {
	if nodeID <= 0 {
		return []*instancecontracts.Instance{}, nil
	}

	now := time.Now().UTC()
	instances := make([]*instancecontracts.Instance, 0)
	if err := r.dbWithContext(ctx).
		Where("node_id = ? AND status IN ? AND expires_at > ?",
			nodeID,
			[]string{
				instancecontracts.InstanceStatusCreating,
				instancecontracts.InstanceStatusRunning,
			},
			now,
		).
		Order("updated_at ASC, id ASC").
		Find(&instances).Error; err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return []*instancecontracts.Instance{}, nil
	}

	ids := make([]int64, 0, len(instances))
	for _, instance := range instances {
		if instance == nil || instance.ID <= 0 {
			continue
		}
		ids = append(ids, instance.ID)
	}
	if len(ids) == 0 {
		return []*instancecontracts.Instance{}, nil
	}

	requeued := make([]*instancecontracts.Instance, 0, len(instances))
	requeueQuery := r.dbWithContext(ctx).Model(&requeued).
		Clauses(clause.Returning{}).
		Where("id IN ?", ids).
		Where("node_id = ? AND status IN ? AND expires_at > ?",
			nodeID,
			[]string{
				instancecontracts.InstanceStatusCreating,
				instancecontracts.InstanceStatusRunning,
			},
			now,
		)
	if err := requeueQuery.
		Updates(map[string]any{
			"status":          instancecontracts.InstanceStatusPending,
			"node_id":         nil,
			"container_id":    "",
			"network_id":      "",
			"runtime_details": "",
			"access_url":      "",
			"updated_at":      now,
		}).Error; err != nil {
		return nil, err
	}
	return requeued, nil
}

func (r *Repository) FinalizeStoppedRuntime(ctx context.Context, id int64) (*RuntimeAllocationRelease, error) {
	return r.finalizeInstanceRuntime(ctx, id, instancecontracts.InstanceStatusStopped)
}

func (r *Repository) ExpireInstanceRuntime(ctx context.Context, id int64) (*RuntimeAllocationRelease, error) {
	return r.finalizeInstanceRuntime(ctx, id, instancecontracts.InstanceStatusExpired)
}

func (r *Repository) finalizeInstanceRuntime(ctx context.Context, id int64, status string) (*RuntimeAllocationRelease, error) {
	if id <= 0 {
		return nil, nil
	}

	var instance instancecontracts.Instance
	if err := r.dbWithContext(ctx).Select("id", "host_port").Where("id = ?", id).First(&instance).Error; err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":          status,
			"host_port":       0,
			"container_id":    "",
			"network_id":      "",
			"runtime_details": "",
			"access_url":      "",
			"destroyed_at":    now,
			"updated_at":      now,
		}).Error; err != nil {
		return nil, err
	}

	return &RuntimeAllocationRelease{
		InstanceID: instance.ID,
		HostPort:   instance.HostPort,
	}, nil
}

func (r *Repository) ListPendingInstances(ctx context.Context, limit int) ([]*instancecontracts.Instance, error) {
	if limit <= 0 {
		return []*instancecontracts.Instance{}, nil
	}

	instances := make([]*instancecontracts.Instance, 0, limit)
	err := r.dbWithContext(ctx).
		Where("status = ?", instancecontracts.InstanceStatusPending).
		Order("created_at ASC, id ASC").
		Limit(limit).
		Find(&instances).Error
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (r *Repository) TryTransitionStatus(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error) {
	result := r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("id = ? AND status = ?", id, fromStatus).
		Updates(map[string]any{
			"status":     toStatus,
			"updated_at": time.Now().UTC(),
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
	err := r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("status IN ?", statuses).
		Count(&count).Error
	return count, err
}

func (r *Repository) CountRunningInstances(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("status = ?", instancecontracts.InstanceStatusRunning).
		Count(&count).Error
	return count, err
}

func applyTeacherInstanceQueryFilters(query *gorm.DB, filter instanceports.TeacherInstanceFilter, now time.Time) *gorm.DB {
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

	switch filter.Status {
	case instancecontracts.InstanceStatusRunning:
		query = query.Where("i.status = ? AND i.expires_at > ?", instancecontracts.InstanceStatusRunning, now)
	case instancecontracts.InstanceStatusCreating:
		query = query.Where("i.status = ?", instancecontracts.InstanceStatusCreating)
	case instancecontracts.InstanceStatusExpired:
		query = query.Where(
			"(i.status = ? AND i.expires_at <= ?) OR i.status = ?",
			instancecontracts.InstanceStatusRunning,
			now,
			instancecontracts.InstanceStatusExpired,
		)
	case instancecontracts.InstanceStatusFailed:
		query = query.Where("i.status = ?", instancecontracts.InstanceStatusFailed)
	case "inactive":
		query = query.Where(
			"i.status NOT IN ?",
			[]string{
				instancecontracts.InstanceStatusRunning,
				instancecontracts.InstanceStatusCreating,
				instancecontracts.InstanceStatusExpired,
				instancecontracts.InstanceStatusFailed,
			},
		)
	}

	return query
}

type instanceMetadata struct {
	Title      string
	Category   string
	Difficulty string
	FlagType   string
}

func buildInstanceMetadata(contestMode, serviceSnapshot, serviceName, challengeTitle, category, difficulty, flagType string) instanceMetadata {
	metadata := instanceMetadata{
		Title:      challengeTitle,
		Category:   category,
		Difficulty: difficulty,
		FlagType:   flagType,
	}
	if contestMode != persistedContestModeAWD {
		return metadata
	}

	snapshot, err := decodeContestAWDServiceSnapshotReadModel(serviceSnapshot)
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
