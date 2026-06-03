package infrastructure

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type Repository struct {
	db *gorm.DB
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

func (r *Repository) FindUserByID(ctx context.Context, userID int64) (*identitycontracts.User, error) {
	var user identitycontracts.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

func (r *Repository) FindChallengeByID(ctx context.Context, challengeID int64) (*challengecontracts.RecommendationChallenge, error) {
	var challenge challengecontracts.RecommendationChallenge
	if err := r.dbWithContext(ctx).
		Table("challenges").
		Select("id, title, category, category AS recommendation_dimension, difficulty, points").
		Where("id = ?", challengeID).
		Take(&challenge).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}

func (r *Repository) FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*instancecontracts.Instance, error) {
	var instance instancecontracts.Instance
	err := r.dbWithContext(ctx).Where("user_id = ? AND contest_id IS NULL AND team_id IS NULL AND challenge_id = ? AND status IN ?", userID, challengeID,
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

func (r *Repository) FindByContestUserID(ctx context.Context, contestID, userID int64) ([]*instancecontracts.Instance, error) {
	var instances []*instancecontracts.Instance
	err := r.dbWithContext(ctx).Where("contest_id = ? AND user_id = ? AND team_id IS NULL AND status IN ?", contestID, userID,
		[]string{instancecontracts.InstanceStatusPending, instancecontracts.InstanceStatusCreating, instancecontracts.InstanceStatusRunning}).
		Order("created_at DESC").
		Find(&instances).Error
	return instances, err
}

func (r *Repository) FindByContestUserAndChallenge(ctx context.Context, contestID, userID, challengeID int64) (*instancecontracts.Instance, error) {
	var instance instancecontracts.Instance
	err := r.dbWithContext(ctx).Where("contest_id = ? AND user_id = ? AND team_id IS NULL AND challenge_id = ? AND status IN ?",
		contestID, userID, challengeID, []string{instancecontracts.InstanceStatusPending, instancecontracts.InstanceStatusCreating, instancecontracts.InstanceStatusRunning}).
		First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) FindByContestTeamID(ctx context.Context, contestID, teamID int64) ([]*instancecontracts.Instance, error) {
	var instances []*instancecontracts.Instance
	err := r.dbWithContext(ctx).Where("contest_id = ? AND team_id = ? AND status IN ?", contestID, teamID,
		[]string{instancecontracts.InstanceStatusPending, instancecontracts.InstanceStatusCreating, instancecontracts.InstanceStatusRunning}).
		Order("created_at DESC").
		Find(&instances).Error
	return instances, err
}

func (r *Repository) FindByContestTeamAndChallenge(ctx context.Context, contestID, teamID, challengeID int64) (*instancecontracts.Instance, error) {
	var instance instancecontracts.Instance
	err := r.dbWithContext(ctx).Where("contest_id = ? AND team_id = ? AND challenge_id = ? AND status IN ?",
		contestID, teamID, challengeID, []string{instancecontracts.InstanceStatusPending, instancecontracts.InstanceStatusCreating, instancecontracts.InstanceStatusRunning}).
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
	return r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("id = ?", instanceID).
		Updates(map[string]any{
			"expires_at": expiresAt,
			"updated_at": time.Now(),
		}).Error
}

func (r *Repository) UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error {
	_, err := r.updateStatusAndReleasePortWithCurrentStatus(ctx, id, nil, status)
	return err
}

func (r *Repository) FailProvisioning(ctx context.Context, id int64) (bool, error) {
	return r.updateStatusAndReleasePortWithCurrentStatus(ctx, id, []string{instancecontracts.InstanceStatusCreating}, instancecontracts.InstanceStatusFailed)
}

func (r *Repository) updateStatusAndReleasePortWithCurrentStatus(ctx context.Context, id int64, currentStatuses []string, status string) (bool, error) {
	if id <= 0 {
		return false, nil
	}

	changed := false
	err := r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var instance instancecontracts.Instance
		query := tx.Select("id", "host_port").Where("id = ?", id)
		if len(currentStatuses) > 0 {
			query = query.Where("status IN ?", currentStatuses)
		}
		if err := query.First(&instance).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) && len(currentStatuses) > 0 {
				return nil
			}
			return err
		}

		updates := map[string]any{
			"status":     status,
			"updated_at": time.Now(),
		}
		if status == instancecontracts.InstanceStatusStopped || status == instancecontracts.InstanceStatusExpired {
			updates["destroyed_at"] = time.Now()
			updates["host_port"] = 0
			updates["container_id"] = ""
			updates["network_id"] = ""
			updates["runtime_details"] = ""
			updates["access_url"] = ""
		}
		updateQuery := tx.Model(&instancecontracts.Instance{}).
			Where("id = ?", id)
		if len(currentStatuses) > 0 {
			updateQuery = updateQuery.Where("status IN ?", currentStatuses)
		}
		result := updateQuery.Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if len(currentStatuses) > 0 && result.RowsAffected == 0 {
			return nil
		}

		deleteQuery := tx.Where("instance_id = ?", id)
		if instance.HostPort > 0 {
			deleteQuery = deleteQuery.Or("port = ?", instance.HostPort)
		}
		if err := deleteQuery.Delete(&runtimeentity.PortAllocation{}).Error; err != nil {
			return err
		}
		if err := tx.Where("instance_id = ?", id).Delete(&runtimeentity.NetworkAllocation{}).Error; err != nil {
			return err
		}
		changed = true
		return nil
	})
	return changed, err
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

func (r *Repository) FinalizeStoppedRuntime(ctx context.Context, id int64) error {
	return r.finalizeInstanceRuntime(ctx, id, instancecontracts.InstanceStatusStopped)
}

func (r *Repository) ExpireInstanceRuntime(ctx context.Context, id int64) error {
	return r.finalizeInstanceRuntime(ctx, id, instancecontracts.InstanceStatusExpired)
}

func (r *Repository) finalizeInstanceRuntime(ctx context.Context, id int64, status string) error {
	if id <= 0 {
		return nil
	}

	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var instance instancecontracts.Instance
		if err := tx.Select("id", "host_port").Where("id = ?", id).First(&instance).Error; err != nil {
			return err
		}

		now := time.Now().UTC()
		if err := tx.Model(&instancecontracts.Instance{}).
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
			return err
		}

		deleteQuery := tx.Where("instance_id = ?", id)
		if instance.HostPort > 0 {
			deleteQuery = deleteQuery.Or("port = ?", instance.HostPort)
		}
		if err := deleteQuery.Delete(&runtimeentity.PortAllocation{}).Error; err != nil {
			return err
		}
		return tx.Where("instance_id = ?", id).Delete(&runtimeentity.NetworkAllocation{}).Error
	})
}

func (r *Repository) UpdateRuntime(ctx context.Context, instance *instancecontracts.Instance) error {
	_, err := r.PersistProvisionedRuntime(ctx, instance)
	return err
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
			"host_port":       instance.HostPort,
			"container_id":    instance.ContainerID,
			"network_id":      instance.NetworkID,
			"runtime_details": instance.RuntimeDetails,
			"access_url":      instance.AccessURL,
			"status":          instance.Status,
			"updated_at":      time.Now(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) FindAWDDefenseWorkspace(ctx context.Context, contestID, teamID, serviceID int64) (*runtimeentity.AWDDefenseWorkspace, error) {
	var workspace runtimeentity.AWDDefenseWorkspace
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

func (r *Repository) UpsertAWDDefenseWorkspace(ctx context.Context, workspace *runtimeentity.AWDDefenseWorkspace) error {
	if workspace == nil {
		return nil
	}

	if workspace.WorkspaceRevision <= 0 {
		workspace.WorkspaceRevision = 1
	}
	if strings.TrimSpace(workspace.Status) == "" {
		workspace.Status = runtimeentity.AWDDefenseWorkspaceStatusPending
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
		var workspace runtimeentity.AWDDefenseWorkspace
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("contest_id = ? AND team_id = ? AND service_id = ?", contestID, teamID, serviceID).
			First(&workspace).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			return tx.Create(&runtimeentity.AWDDefenseWorkspace{
				ContestID:         contestID,
				TeamID:            teamID,
				ServiceID:         serviceID,
				InstanceID:        instanceID,
				WorkspaceRevision: 1,
				Status:            runtimeentity.AWDDefenseWorkspaceStatusProvisioning,
				SeedSignature:     seedSignature,
				CreatedAt:         now,
				UpdatedAt:         now,
			}).Error
		}

		return tx.Model(&runtimeentity.AWDDefenseWorkspace{}).
			Where("id = ?", workspace.ID).
			Updates(map[string]any{
				"instance_id":        instanceID,
				"workspace_revision": workspace.WorkspaceRevision + 1,
				"status":             runtimeentity.AWDDefenseWorkspaceStatusProvisioning,
				"container_id":       "",
				"seed_signature":     seedSignature,
				"updated_at":         now,
			}).Error
	})
}

func (r *Repository) FindAccessibleByIDForUser(ctx context.Context, instanceID, userID int64) (*instancecontracts.Instance, error) {
	var instance instancecontracts.Instance
	query := r.db.WithContext(ctx).
		Table("instances AS inst").
		Select("inst.*").
		Joins("LEFT JOIN team_members AS tm ON tm.team_id = inst.team_id AND tm.contest_id = inst.contest_id AND tm.user_id = ?", userID).
		Joins("LEFT JOIN contest_registrations AS reg ON reg.contest_id = inst.contest_id AND reg.user_id = ? AND reg.status = ?", userID, contestcontracts.ContestRegistrationStatusApproved)
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

func (r *Repository) FindVisibleByUser(ctx context.Context, userID int64) ([]*instancecontracts.Instance, error) {
	var instances []*instancecontracts.Instance
	query := r.db.WithContext(ctx).
		Table("instances AS inst").
		Select("DISTINCT inst.*").
		Joins("LEFT JOIN team_members AS tm ON tm.team_id = inst.team_id AND tm.contest_id = inst.contest_id AND tm.user_id = ?", userID).
		Joins("LEFT JOIN contest_registrations AS reg ON reg.contest_id = inst.contest_id AND reg.user_id = ? AND reg.status = ?", userID, contestcontracts.ContestRegistrationStatusApproved)
	query = joinAWDActiveScopeControls(query, "inst.contest_id", "inst.team_id", "inst.service_id", "visible_team_retired_ctl", "visible_service_disabled_ctl")
	err := applyAWDActiveScopeFilter(query, "inst.service_id", "visible_team_retired_ctl", "visible_service_disabled_ctl").
		Where("inst.status IN ?", []string{
			instancecontracts.InstanceStatusPending,
			instancecontracts.InstanceStatusCreating,
			instancecontracts.InstanceStatusRunning,
			instancecontracts.InstanceStatusStopping,
			instancecontracts.InstanceStatusFailed,
			instancecontracts.InstanceStatusExpired,
		}).
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
	query := r.db.WithContext(ctx).
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
		Joins("LEFT JOIN contest_registrations AS reg ON reg.contest_id = inst.contest_id AND reg.user_id = ? AND reg.status = ?", userID, contestcontracts.ContestRegistrationStatusApproved)
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
		Where("(co.mode IS NULL OR co.mode <> ? OR cas.id IS NOT NULL)", contestcontracts.ContestModeAWD).
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

func (r *Repository) FindExpired(ctx context.Context) ([]*instancecontracts.Instance, error) {
	var instances []*instancecontracts.Instance
	err := r.dbWithContext(ctx).Where("status = ? AND expires_at < ?",
		instancecontracts.InstanceStatusRunning, time.Now()).
		Find(&instances).Error
	return instances, err
}

func (r *Repository) ListRecoverableActiveInstances(ctx context.Context) ([]*instancecontracts.Instance, error) {
	var instances []*instancecontracts.Instance
	err := r.dbWithContext(ctx).
		Where("status IN ?", []string{instancecontracts.InstanceStatusCreating, instancecontracts.InstanceStatusRunning}).
		Where("expires_at > ?", time.Now()).
		Order("updated_at ASC, id ASC").
		Find(&instances).Error
	return instances, err
}

func (r *Repository) ListStoppingInstances(ctx context.Context, updatedBefore time.Time) ([]*instancecontracts.Instance, error) {
	var instances []*instancecontracts.Instance
	query := r.dbWithContext(ctx).
		Where("status = ?", instancecontracts.InstanceStatusStopping)
	if !updatedBefore.IsZero() {
		query = query.Where("updated_at <= ?", updatedBefore)
	}
	err := query.Order("updated_at ASC, id ASC").Find(&instances).Error
	return instances, err
}

func (r *Repository) FindRunningAWDDefenseWorkspaceByInstanceID(ctx context.Context, instanceID int64) (*runtimeentity.AWDDefenseWorkspace, error) {
	if instanceID <= 0 {
		return nil, nil
	}

	var workspace runtimeentity.AWDDefenseWorkspace
	err := r.dbWithContext(ctx).
		Where("instance_id = ?", instanceID).
		Where("status = ?", runtimeentity.AWDDefenseWorkspaceStatusRunning).
		Where("container_id <> ''").
		First(&workspace).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "no such table") || strings.Contains(lowerErr, "does not exist") {
			return nil, nil
		}
		return nil, err
	}
	return &workspace, nil
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
			[]string{instancecontracts.InstanceStatusCreating, instancecontracts.InstanceStatusRunning},
			time.Now(),
		).
		Updates(map[string]any{
			"status":          instancecontracts.InstanceStatusPending,
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

func (r *Repository) CreateAWDServiceOperation(ctx context.Context, operation *runtimeentity.AWDServiceOperation) error {
	return r.dbWithContext(ctx).Create(operation).Error
}

func (r *Repository) FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error {
	if instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Model(&runtimeentity.AWDServiceOperation{}).
		Where("instance_id = ? AND status IN ?", instanceID, []string{
			runtimeentity.AWDServiceOperationStatusRequested,
			runtimeentity.AWDServiceOperationStatusProvisioning,
			runtimeentity.AWDServiceOperationStatusRecovering,
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
		Model(&runtimeentity.AWDServiceOperation{}).
		Where("id = ?", operationID).
		Updates(map[string]any{
			"status":        status,
			"error_message": errorMessage,
			"finished_at":   finishedAt,
			"updated_at":    time.Now(),
		}).Error
}

func (r *Repository) ListTeacherInstances(ctx context.Context, filter runtimeports.TeacherInstanceFilter) (*runtimeports.TeacherInstancePage, error) {
	rows := make([]teacherInstanceRow, 0)
	now := time.Now()

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
		Where("(co.mode IS NULL OR co.mode <> ? OR cas.id IS NOT NULL)", contestcontracts.ContestModeAWD).
		Where("u.role = ? AND u.deleted_at IS NULL", identitycontracts.RoleStudent)

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
	return &runtimeports.TeacherInstancePage{
		List:  items,
		Total: summary.TotalCount,
		Summary: runtimeports.TeacherInstanceListSummary{
			TotalCount:        summary.TotalCount,
			RunningCount:      summary.RunningCount,
			ExpiringSoonCount: summary.ExpiringSoonCount,
			WarningCount:      summary.WarningCount,
		},
	}, nil
}

func applyTeacherInstanceQueryFilters(query *gorm.DB, filter runtimeports.TeacherInstanceFilter, now time.Time) *gorm.DB {
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
	if contestMode != contestcontracts.ContestModeAWD {
		return metadata
	}

	snapshot, err := contestcontracts.DecodeContestAWDServiceSnapshot(serviceSnapshot)
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
	return r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"expires_at":   expiresAt,
			"extend_count": extendCount,
		}).Error
}

func (r *Repository) AtomicExtend(ctx context.Context, id int64, userID int64, maxExtends int, duration time.Duration) error {
	result := r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("id = ? AND user_id = ? AND status = ? AND extend_count < ?",
			id, userID, instancecontracts.InstanceStatusRunning, maxExtends).
		Updates(map[string]interface{}{
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

func (r *Repository) AtomicExtendByID(ctx context.Context, id int64, maxExtends int, duration time.Duration) error {
	result := r.db.WithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("id = ? AND status = ? AND extend_count < ?",
			id, instancecontracts.InstanceStatusRunning, maxExtends).
		Updates(map[string]interface{}{
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

func (r *Repository) CountRunning(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("status = ?", instancecontracts.InstanceStatusRunning).
		Count(&count).Error
	return count, err
}

func (r *Repository) ListPendingInstances(ctx context.Context, limit int) ([]*instancecontracts.Instance, error) {
	if limit <= 0 {
		return []*instancecontracts.Instance{}, nil
	}

	instances := make([]*instancecontracts.Instance, 0, limit)
	err := r.db.WithContext(ctx).
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
	result := r.db.WithContext(ctx).Model(&instancecontracts.Instance{}).
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
	err := r.db.WithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("status IN ?", statuses).
		Count(&count).Error
	return count, err
}

func (r *Repository) ReserveAvailablePort(ctx context.Context, start, end int) (int, error) {
	return r.ReserveAvailablePortExcluding(ctx, start, end, 0)
}

func (r *Repository) ReserveAvailablePortExcluding(ctx context.Context, start, end, excludedPort int) (int, error) {
	for port := start; port < end; port++ {
		if excludedPort > 0 && port == excludedPort {
			continue
		}
		reserved, err := r.tryReservePort(ctx, port)
		if err != nil {
			return 0, err
		}
		if reserved {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port in range %d-%d", start, end)
}

func (r *Repository) tryReservePort(ctx context.Context, port int) (bool, error) {
	now := time.Now().UTC()
	result := r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "port"}},
		DoUpdates: clause.Assignments(map[string]any{
			"updated_at": now,
		}),
		Where: clause.Where{Exprs: []clause.Expression{
			clause.Expr{SQL: "port_allocations.instance_id IS NULL"},
		}},
	}).Create(&runtimeentity.PortAllocation{
		Port:      port,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if result.Error != nil {
		if isPortAllocationConflict(result.Error) {
			return false, nil
		}
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) BindReservedPort(ctx context.Context, port int, instanceID int64) error {
	return r.dbWithContext(ctx).Model(&runtimeentity.PortAllocation{}).
		Where("port = ?", port).
		Updates(map[string]any{
			"instance_id": instanceID,
			"updated_at":  time.Now(),
		}).Error
}

func (r *Repository) ReleaseReservedPort(ctx context.Context, port int) error {
	if port <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("port = ? AND instance_id IS NULL", port).
		Delete(&runtimeentity.PortAllocation{}).Error
}

func (r *Repository) ReleasePortForInstance(ctx context.Context, port int, instanceID int64) error {
	if port <= 0 || instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("port = ? AND instance_id = ?", port, instanceID).
		Delete(&runtimeentity.PortAllocation{}).Error
}

func (r *Repository) ReserveAvailableSubnet(ctx context.Context, baseCIDR string, subnetMask int) (string, error) {
	return r.ReserveAvailableSubnetExcluding(ctx, baseCIDR, subnetMask, nil)
}

func (r *Repository) ReserveAvailableSubnetForInstance(ctx context.Context, baseCIDR string, subnetMask int, instanceID int64, networkKey string) (string, error) {
	return r.ReserveAvailableSubnetForInstanceExcluding(ctx, baseCIDR, subnetMask, instanceID, networkKey, nil)
}

func (r *Repository) ReserveAvailableSubnetExcluding(ctx context.Context, baseCIDR string, subnetMask int, excludedSubnets []string) (string, error) {
	return r.reserveAvailableSubnet(ctx, baseCIDR, subnetMask, 0, "", excludedSubnets)
}

func (r *Repository) ReserveAvailableSubnetForInstanceExcluding(ctx context.Context, baseCIDR string, subnetMask int, instanceID int64, networkKey string, excludedSubnets []string) (string, error) {
	return r.reserveAvailableSubnet(ctx, baseCIDR, subnetMask, instanceID, networkKey, excludedSubnets)
}

func (r *Repository) reserveAvailableSubnet(ctx context.Context, baseCIDR string, subnetMask int, instanceID int64, networkKey string, excludedSubnets []string) (string, error) {
	normalizedKey := strings.TrimSpace(networkKey)
	excludedSet := make(map[string]struct{}, len(excludedSubnets))
	for _, subnet := range excludedSubnets {
		subnet = strings.TrimSpace(subnet)
		if subnet == "" {
			continue
		}
		excludedSet[subnet] = struct{}{}
	}
	if instanceID > 0 && normalizedKey != "" {
		existing, err := r.findSubnetAllocationByOwner(ctx, instanceID, normalizedKey)
		if err != nil {
			return "", err
		}
		if existing != "" {
			if _, excluded := excludedSet[existing]; excluded {
				existing = ""
			}
		}
		if existing != "" {
			return existing, nil
		}
	}

	candidates, err := subnetCandidates(baseCIDR, subnetMask)
	if err != nil {
		return "", err
	}
	for _, subnet := range candidates {
		if _, excluded := excludedSet[subnet]; excluded {
			continue
		}
		reserved, reserveErr := r.tryReserveSubnet(ctx, subnet, instanceID, normalizedKey)
		if reserveErr != nil {
			return "", reserveErr
		}
		if reserved {
			return subnet, nil
		}
		if instanceID > 0 && normalizedKey != "" {
			existing, findErr := r.findSubnetAllocationByOwner(ctx, instanceID, normalizedKey)
			if findErr != nil {
				return "", findErr
			}
			if existing != "" {
				if _, excluded := excludedSet[existing]; excluded {
					moved, moveErr := r.moveSubnetAllocationForOwner(ctx, instanceID, normalizedKey, existing, subnet)
					if moveErr != nil {
						return "", moveErr
					}
					if moved {
						return subnet, nil
					}
					continue
				}
				return existing, nil
			}
		}
	}
	return "", fmt.Errorf("no available subnet in %s with /%d", baseCIDR, subnetMask)
}

func (r *Repository) tryReserveSubnet(ctx context.Context, subnet string, instanceID int64, networkKey string) (bool, error) {
	now := time.Now().UTC()
	allocation := &runtimeentity.NetworkAllocation{
		Subnet:     subnet,
		NetworkKey: networkKey,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if instanceID > 0 {
		allocation.InstanceID = &instanceID
	}

	result := r.dbWithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(allocation)
	if result.Error != nil {
		if isNetworkAllocationConflict(result.Error) {
			return false, nil
		}
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) moveSubnetAllocationForOwner(ctx context.Context, instanceID int64, networkKey, currentSubnet, targetSubnet string) (bool, error) {
	if instanceID <= 0 || strings.TrimSpace(networkKey) == "" {
		return false, nil
	}
	currentSubnet = strings.TrimSpace(currentSubnet)
	targetSubnet = strings.TrimSpace(targetSubnet)
	if currentSubnet == "" || targetSubnet == "" || currentSubnet == targetSubnet {
		return false, nil
	}

	result := r.dbWithContext(ctx).
		Model(&runtimeentity.NetworkAllocation{}).
		Where("instance_id = ? AND network_key = ? AND subnet = ?", instanceID, networkKey, currentSubnet).
		Updates(map[string]any{
			"subnet":     targetSubnet,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		if isNetworkAllocationConflict(result.Error) {
			return false, nil
		}
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *Repository) findSubnetAllocationByOwner(ctx context.Context, instanceID int64, networkKey string) (string, error) {
	if instanceID <= 0 || strings.TrimSpace(networkKey) == "" {
		return "", nil
	}

	var allocation runtimeentity.NetworkAllocation
	err := r.dbWithContext(ctx).
		Where("instance_id = ? AND network_key = ?", instanceID, networkKey).
		First(&allocation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return allocation.Subnet, nil
}

func (r *Repository) ReleaseReservedSubnet(ctx context.Context, subnet string) error {
	subnet = strings.TrimSpace(subnet)
	if subnet == "" {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("subnet = ? AND instance_id IS NULL", subnet).
		Delete(&runtimeentity.NetworkAllocation{}).Error
}

func (r *Repository) ReleaseSubnetForInstance(ctx context.Context, subnet string, instanceID int64) error {
	subnet = strings.TrimSpace(subnet)
	if subnet == "" || instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("subnet = ? AND instance_id = ?", subnet, instanceID).
		Delete(&runtimeentity.NetworkAllocation{}).Error
}

func (r *Repository) IsHostPortReusableForRestart(ctx context.Context, instanceID int64, hostPort int) (bool, error) {
	if instanceID <= 0 || hostPort <= 0 {
		return false, nil
	}

	var allocation runtimeentity.PortAllocation
	if err := r.dbWithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("port = ?", hostPort).
		First(&allocation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	if allocation.InstanceID == nil {
		return false, nil
	}
	return *allocation.InstanceID == instanceID, nil
}

func (r *Repository) SyncInstanceHostPortForRestart(ctx context.Context, instanceID int64, hostPort int, preserveHostPort bool) (int, error) {
	if instanceID <= 0 {
		return 0, nil
	}
	if !preserveHostPort {
		return 0, r.releaseAllPortsForInstance(ctx, instanceID)
	}

	boundPort, err := r.findLatestBoundPortForInstance(ctx, instanceID)
	if err != nil {
		return 0, err
	}
	if boundPort > 0 {
		hostPort = boundPort
	}
	if hostPort <= 0 {
		return 0, nil
	}
	if err := r.ensurePortBoundToInstance(ctx, hostPort, instanceID); err != nil {
		return 0, err
	}
	return hostPort, nil
}

func (r *Repository) findLatestBoundPortForInstance(ctx context.Context, instanceID int64) (int, error) {
	if instanceID <= 0 {
		return 0, nil
	}

	var allocation runtimeentity.PortAllocation
	err := r.dbWithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("instance_id = ?", instanceID).
		Order("updated_at DESC, port DESC").
		First(&allocation).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return allocation.Port, nil
}

func (r *Repository) ensurePortBoundToInstance(ctx context.Context, port int, instanceID int64) error {
	if port <= 0 || instanceID <= 0 {
		return nil
	}

	allocation := &runtimeentity.PortAllocation{
		Port:       port,
		InstanceID: &instanceID,
	}
	if err := r.dbWithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(allocation).Error; err != nil {
		return err
	}

	var stored runtimeentity.PortAllocation
	if err := r.dbWithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("port = ?", port).
		First(&stored).Error; err != nil {
		return err
	}
	if stored.InstanceID != nil && *stored.InstanceID != instanceID {
		return fmt.Errorf("host port %d is allocated to instance %d", port, *stored.InstanceID)
	}
	if stored.InstanceID == nil {
		return r.dbWithContext(ctx).Model(&runtimeentity.PortAllocation{}).
			Where("port = ?", port).
			Updates(map[string]any{
				"instance_id": instanceID,
				"updated_at":  time.Now().UTC(),
			}).Error
	}
	return nil
}

func (r *Repository) releaseAllPortsForInstance(ctx context.Context, instanceID int64) error {
	if instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Where("instance_id = ?", instanceID).
		Delete(&runtimeentity.PortAllocation{}).Error
}

func (r *Repository) ListActiveContainerIDs(ctx context.Context) ([]string, error) {
	var items []struct {
		ContainerID    string
		RuntimeDetails string
	}
	if err := r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("status IN ?", []string{
			instancecontracts.InstanceStatusCreating,
			instancecontracts.InstanceStatusRunning,
			instancecontracts.InstanceStatusStopping,
		}).
		Select("container_id, runtime_details").
		Scan(&items).Error; err != nil {
		return nil, err
	}
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		ids := []string{item.ContainerID}
		details, err := runtimecontracts.DecodeInstanceRuntimeDetails(item.RuntimeDetails)
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

	var workspaceItems []struct {
		ContainerID string
	}
	if err := r.dbWithContext(ctx).
		Table("awd_defense_workspaces AS ws").
		Joins("JOIN instances AS inst ON inst.id = ws.instance_id").
		Where("inst.status IN ?", []string{
			instancecontracts.InstanceStatusCreating,
			instancecontracts.InstanceStatusRunning,
			instancecontracts.InstanceStatusStopping,
		}).
		Where("ws.status = ? AND ws.container_id <> ''", runtimeentity.AWDDefenseWorkspaceStatusRunning).
		Select("ws.container_id").
		Scan(&workspaceItems).Error; err != nil {
		lowerErr := strings.ToLower(err.Error())
		if !strings.Contains(lowerErr, "no such table") && !strings.Contains(lowerErr, "does not exist") {
			return nil, err
		}
		return result, nil
	}
	for _, item := range workspaceItems {
		containerID := strings.TrimSpace(item.ContainerID)
		if containerID == "" {
			continue
		}
		if _, exists := seen[containerID]; exists {
			continue
		}
		seen[containerID] = struct{}{}
		result = append(result, containerID)
	}
	return result, nil
}

func (r *Repository) FindRuntimeNodeIDByContainerID(ctx context.Context, containerID string) (*int64, error) {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return nil, nil
	}

	type instanceContainerLookupRow struct {
		NodeID         *int64 `gorm:"column:node_id"`
		ContainerID    string `gorm:"column:container_id"`
		RuntimeDetails string `gorm:"column:runtime_details"`
	}

	rows := make([]instanceContainerLookupRow, 0)
	likePattern := "%" + containerID + "%"
	if err := r.dbWithContext(ctx).
		Model(&instancecontracts.Instance{}).
		Select("node_id, container_id, runtime_details").
		Where("container_id = ? OR runtime_details LIKE ?", containerID, likePattern).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		if row.ContainerID == containerID {
			return row.NodeID, nil
		}
		details, err := runtimecontracts.DecodeInstanceRuntimeDetails(row.RuntimeDetails)
		if err != nil {
			continue
		}
		for _, item := range details.Containers {
			if strings.TrimSpace(item.ContainerID) == containerID {
				return row.NodeID, nil
			}
		}
	}

	type workspaceContainerLookupRow struct {
		NodeID *int64 `gorm:"column:node_id"`
	}
	var workspace workspaceContainerLookupRow
	if err := r.dbWithContext(ctx).
		Table("awd_defense_workspaces AS ws").
		Joins("JOIN instances AS inst ON inst.id = ws.instance_id").
		Where("ws.container_id = ?", containerID).
		Select("inst.node_id").
		Take(&workspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "no such table") || strings.Contains(lowerErr, "does not exist") {
			return nil, nil
		}
		return nil, err
	}
	return workspace.NodeID, nil
}

func (r *Repository) ListInstancesNeedingACLHandleMigration(ctx context.Context) ([]instancecontracts.Instance, error) {
	type instanceACLMigrationRow struct {
		ID             int64  `gorm:"column:id"`
		NodeID         *int64 `gorm:"column:node_id"`
		RuntimeDetails string `gorm:"column:runtime_details"`
	}

	rows := make([]instanceACLMigrationRow, 0)
	if err := r.dbWithContext(ctx).
		Model(&instancecontracts.Instance{}).
		Where("destroyed_at IS NULL").
		Where("runtime_details <> ''").
		Select("id, node_id, runtime_details").
		Scan(&rows).Error; err != nil {
		lowerErr := strings.ToLower(err.Error())
		if strings.Contains(lowerErr, "no such table") || strings.Contains(lowerErr, "does not exist") {
			return nil, nil
		}
		return nil, err
	}

	result := make([]instancecontracts.Instance, 0, len(rows))
	for _, row := range rows {
		details, err := runtimecontracts.DecodeInstanceRuntimeDetails(row.RuntimeDetails)
		if err != nil || details.ACL != nil || len(details.ACLRules) == 0 {
			continue
		}
		result = append(result, instancecontracts.Instance{
			ID:             row.ID,
			NodeID:         row.NodeID,
			RuntimeDetails: row.RuntimeDetails,
		})
	}
	return result, nil
}

func (r *Repository) UpdateInstanceRuntimeDetails(ctx context.Context, instanceID int64, runtimeDetails string) error {
	if instanceID <= 0 {
		return nil
	}
	return r.dbWithContext(ctx).
		Model(&instancecontracts.Instance{}).
		Where("id = ?", instanceID).
		Updates(map[string]any{
			"runtime_details": runtimeDetails,
			"updated_at":      time.Now().UTC(),
		}).Error
}

func (r *Repository) ListAllocatedPorts(ctx context.Context) ([]int, error) {
	var ports []int
	if err := r.dbWithContext(ctx).Model(&runtimeentity.PortAllocation{}).Pluck("port", &ports).Error; err == nil {
		return ports, nil
	} else if !strings.Contains(strings.ToLower(err.Error()), "no such table") && !strings.Contains(strings.ToLower(err.Error()), "does not exist") {
		return nil, err
	}

	var accessURLs []string
	if err := r.dbWithContext(ctx).Model(&instancecontracts.Instance{}).
		Where("status IN ?", []string{instancecontracts.InstanceStatusCreating, instancecontracts.InstanceStatusRunning}).
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

func isNetworkAllocationConflict(err error) bool {
	return isPortAllocationConflict(err)
}

func subnetCandidates(baseCIDR string, subnetMask int) ([]string, error) {
	baseIP, baseNet, err := net.ParseCIDR(strings.TrimSpace(baseCIDR))
	if err != nil {
		return nil, fmt.Errorf("parse subnet base: %w", err)
	}
	ip4 := baseIP.To4()
	if ip4 == nil {
		return nil, fmt.Errorf("subnet base must be ipv4")
	}
	baseNet.IP = ip4

	basePrefix, bits := baseNet.Mask.Size()
	if bits != 32 {
		return nil, fmt.Errorf("subnet base must be ipv4")
	}
	if subnetMask <= basePrefix || subnetMask > 30 {
		return nil, fmt.Errorf("invalid subnet mask /%d for base %s", subnetMask, baseCIDR)
	}

	start := binary.BigEndian.Uint32(baseNet.IP)
	blockSize := uint32(1) << uint32(32-subnetMask)
	subnetCount := 1 << uint(subnetMask-basePrefix)
	result := make([]string, 0, subnetCount)
	for idx := 0; idx < subnetCount; idx++ {
		current := start + uint32(idx)*blockSize
		ip := make(net.IP, net.IPv4len)
		binary.BigEndian.PutUint32(ip, current)
		result = append(result, (&net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(subnetMask, 32),
		}).String())
	}
	return result, nil
}
