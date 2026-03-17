package container

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

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type Repository struct {
	db *gorm.DB
}

type TeacherInstanceFilter struct {
	ClassName string
	Keyword   string
	StudentNo string
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithDB(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(instance *model.Instance) error {
	return r.CreateWithContext(context.Background(), instance)
}

func (r *Repository) CreateWithContext(ctx context.Context, instance *model.Instance) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return r.db.WithContext(ctx).Create(instance).Error
}

func (r *Repository) FindByID(id int64) (*model.Instance, error) {
	var instance model.Instance
	err := r.db.Where("id = ?", id).First(&instance).Error
	if err != nil {
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

func (r *Repository) FindByUserID(userID int64) ([]*model.Instance, error) {
	return r.FindByUserIDWithContext(context.Background(), userID)
}

func (r *Repository) FindByUserIDWithContext(ctx context.Context, userID int64) ([]*model.Instance, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	var instances []*model.Instance
	err := r.db.WithContext(ctx).Where("user_id = ? AND contest_id IS NULL AND team_id IS NULL AND status IN ?", userID,
		[]string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
		Order("created_at DESC").
		Find(&instances).Error
	return instances, err
}

func (r *Repository) FindByUserAndChallenge(userID, challengeID int64) (*model.Instance, error) {
	var instance model.Instance
	err := r.db.Where("user_id = ? AND contest_id IS NULL AND team_id IS NULL AND challenge_id = ? AND status IN ?", userID, challengeID,
		[]string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
		First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) FindByContestUserID(contestID, userID int64) ([]*model.Instance, error) {
	var instances []*model.Instance
	err := r.db.Where("contest_id = ? AND user_id = ? AND team_id IS NULL AND status IN ?", contestID, userID,
		[]string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
		Order("created_at DESC").
		Find(&instances).Error
	return instances, err
}

func (r *Repository) FindByContestUserAndChallenge(contestID, userID, challengeID int64) (*model.Instance, error) {
	var instance model.Instance
	err := r.db.Where("contest_id = ? AND user_id = ? AND team_id IS NULL AND challenge_id = ? AND status IN ?",
		contestID, userID, challengeID, []string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
		First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) FindByContestTeamID(contestID, teamID int64) ([]*model.Instance, error) {
	var instances []*model.Instance
	err := r.db.Where("contest_id = ? AND team_id = ? AND status IN ?", contestID, teamID,
		[]string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
		Order("created_at DESC").
		Find(&instances).Error
	return instances, err
}

func (r *Repository) FindByContestTeamAndChallenge(contestID, teamID, challengeID int64) (*model.Instance, error) {
	var instance model.Instance
	err := r.db.Where("contest_id = ? AND team_id = ? AND challenge_id = ? AND status IN ?",
		contestID, teamID, challengeID, []string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
		First(&instance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &instance, nil
}

func (r *Repository) UpdateStatus(id int64, status string) error {
	return r.UpdateStatusWithContext(context.Background(), id, status)
}

func (r *Repository) UpdateStatusWithContext(ctx context.Context, id int64, status string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	return r.db.WithContext(ctx).Model(&model.Instance{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *Repository) UpdateStatusAndReleasePort(id int64, status string) error {
	if id <= 0 {
		return nil
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		var instance model.Instance
		if err := tx.Select("id", "host_port").Where("id = ?", id).First(&instance).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Instance{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"status":     status,
				"updated_at": time.Now(),
			}).Error; err != nil {
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

func (r *Repository) UpdateRuntime(instance *model.Instance) error {
	return r.db.Model(&model.Instance{}).
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

func (r *Repository) FindAccessibleByIDForUser(ctx context.Context, instanceID, userID int64) (*model.Instance, error) {
	var instance model.Instance
	err := r.db.WithContext(ctx).
		Table("instances AS inst").
		Select("inst.*").
		Joins("LEFT JOIN team_members AS tm ON tm.team_id = inst.team_id AND tm.contest_id = inst.contest_id AND tm.user_id = ?", userID).
		Where("inst.id = ?", instanceID).
		Where("(inst.team_id IS NULL AND inst.user_id = ?) OR tm.user_id IS NOT NULL", userID).
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
		Where("inst.status IN ?", []string{model.InstanceStatusCreating, model.InstanceStatusRunning}).
		Where("(inst.team_id IS NULL AND inst.user_id = ?) OR tm.user_id IS NOT NULL", userID).
		Order("inst.created_at DESC").
		Scan(&instances).Error
	return instances, err
}

func (r *Repository) FindExpired() ([]*model.Instance, error) {
	var instances []*model.Instance
	err := r.db.Where("status = ? AND expires_at < ?",
		model.InstanceStatusRunning, time.Now()).
		Find(&instances).Error
	return instances, err
}

func (r *Repository) ListTeacherInstances(ctx context.Context, filter TeacherInstanceFilter) ([]dto.TeacherInstanceItem, error) {
	items := make([]dto.TeacherInstanceItem, 0)

	query := r.db.WithContext(ctx).
		Table("instances AS i").
		Select(strings.Join([]string{
			"i.id",
			"u.id AS student_id",
			"u.username AS student_name",
			"u.username AS student_username",
			"NULLIF(u.student_no, '') AS student_no",
			"u.class_name",
			"i.challenge_id",
			"c.title AS challenge_title",
			"i.status",
			"i.access_url",
			"i.expires_at",
			"i.extend_count",
			"i.max_extends",
			"i.created_at",
		}, ", ")).
		Joins("JOIN users u ON u.id = i.user_id").
		Joins("JOIN challenges c ON c.id = i.challenge_id").
		Where("i.status <> ?", model.InstanceStatusStopped).
		Where("u.role = ? AND u.deleted_at IS NULL", model.RoleStudent)

	if filter.ClassName != "" {
		query = query.Where("u.class_name = ?", filter.ClassName)
	}
	if filter.StudentNo != "" {
		query = query.Where("u.student_no = ?", filter.StudentNo)
	}
	if filter.Keyword != "" {
		pattern := "%" + strings.ToLower(filter.Keyword) + "%"
		query = query.Where("LOWER(u.username) LIKE ?", pattern)
	}

	if err := query.Order("i.created_at DESC").Scan(&items).Error; err != nil {
		return nil, fmt.Errorf("list teacher instances: %w", err)
	}
	return items, nil
}

func (r *Repository) UpdateExtend(id int64, expiresAt time.Time, extendCount int) error {
	return r.db.Model(&model.Instance{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"expires_at":   expiresAt,
			"extend_count": extendCount,
		}).Error
}

func (r *Repository) AtomicExtend(id int64, userID int64, maxExtends int, duration time.Duration) error {
	result := r.db.Model(&model.Instance{}).
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

func (r *Repository) AtomicExtendByID(id int64, maxExtends int, duration time.Duration) error {
	return r.AtomicExtendByIDWithContext(context.Background(), id, maxExtends, duration)
}

func (r *Repository) AtomicExtendByIDWithContext(ctx context.Context, id int64, maxExtends int, duration time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}

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

func (r *Repository) CountRunning() (int64, error) {
	var count int64
	err := r.db.Model(&model.Instance{}).
		Where("status = ?", model.InstanceStatusRunning).
		Count(&count).Error
	return count, err
}

func (r *Repository) ReserveAvailablePort(start, end int) (int, error) {
	for port := start; port < end; port++ {
		if err := r.db.Create(&model.PortAllocation{Port: port}).Error; err != nil {
			if isPortAllocationConflict(err) {
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

func (r *Repository) ReleasePort(port int) error {
	if port <= 0 {
		return nil
	}
	return r.db.Where("port = ?", port).Delete(&model.PortAllocation{}).Error
}

func (r *Repository) ListActiveContainerIDs() ([]string, error) {
	var items []struct {
		ContainerID    string
		RuntimeDetails string
	}
	if err := r.db.Model(&model.Instance{}).
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

func (r *Repository) ListAllocatedPorts() ([]int, error) {
	var ports []int
	if err := r.db.Model(&model.PortAllocation{}).Pluck("port", &ports).Error; err == nil {
		return ports, nil
	} else if !strings.Contains(strings.ToLower(err.Error()), "no such table") && !strings.Contains(strings.ToLower(err.Error()), "does not exist") {
		return nil, err
	}

	var accessURLs []string
	if err := r.db.Model(&model.Instance{}).
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
