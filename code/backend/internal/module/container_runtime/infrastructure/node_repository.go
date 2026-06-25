package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimeentity "ctf-platform/internal/module/container_runtime/entity"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"

	"gorm.io/gorm"
)

type RuntimeNodeRepository struct {
	db *gorm.DB
}

func NewRuntimeNodeRepository(db *gorm.DB) *RuntimeNodeRepository {
	if db == nil {
		return nil
	}
	return &RuntimeNodeRepository{db: db}
}

func (r *RuntimeNodeRepository) dbWithContext(ctx context.Context) *gorm.DB {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.WithContext(ctx)
}

func (r *RuntimeNodeRepository) EnsureDefaultNode(ctx context.Context, spec runtimecontracts.RuntimeNodeBootstrapSpec) (*runtimeentity.RuntimeNode, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}

	name := strings.TrimSpace(spec.Name)
	if name == "" {
		name = "default"
	}
	endpoint := strings.TrimSpace(spec.Endpoint)
	if endpoint == "" {
		endpoint = "local://docker"
	}
	publicHost := strings.TrimSpace(spec.PublicHost)
	accessHost := strings.TrimSpace(spec.AccessHost)
	now := time.Now().UTC()

	var node runtimeentity.RuntimeNode
	err := r.dbWithContext(ctx).Where("name = ?", name).First(&node).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		node = runtimeentity.RuntimeNode{
			Name:             name,
			Endpoint:         endpoint,
			PublicHost:       publicHost,
			AccessHost:       accessHost,
			TLSIdentity:      strings.TrimSpace(spec.TLSIdentity),
			Schedulable:      spec.Schedulable,
			Labels:           "{}",
			HealthStatus:     runtimeentity.RuntimeNodeHealthUnknown,
			CapacitySnapshot: "{}",
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := r.dbWithContext(ctx).Create(&node).Error; err != nil {
			return nil, err
		}
		return &node, nil
	}
	if err != nil {
		return nil, err
	}

	updates := map[string]any{
		"endpoint":     endpoint,
		"public_host":  publicHost,
		"access_host":  accessHost,
		"tls_identity": strings.TrimSpace(spec.TLSIdentity),
		"schedulable":  spec.Schedulable,
		"updated_at":   now,
	}
	if err := r.dbWithContext(ctx).Model(&runtimeentity.RuntimeNode{}).
		Where("id = ?", node.ID).
		Updates(updates).Error; err != nil {
		return nil, err
	}
	node.Endpoint = endpoint
	node.PublicHost = publicHost
	node.AccessHost = accessHost
	node.TLSIdentity = strings.TrimSpace(spec.TLSIdentity)
	node.Schedulable = spec.Schedulable
	node.UpdatedAt = now
	return &node, nil
}

func (r *RuntimeNodeRepository) FindFirstSchedulableNode(ctx context.Context) (*runtimeentity.RuntimeNode, error) {
	if r == nil || r.db == nil {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}

	var node runtimeentity.RuntimeNode
	if err := r.dbWithContext(ctx).
		Where("schedulable = ?", true).
		Order("id ASC").
		First(&node).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, runtimeports.ErrRuntimeNodeUnavailable
		}
		return nil, err
	}
	return &node, nil
}

func (r *RuntimeNodeRepository) ListSchedulableHealthyNodes(ctx context.Context, staleThreshold time.Duration, now time.Time) ([]runtimeentity.RuntimeNode, error) {
	if r == nil || r.db == nil {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}

	nodes := make([]runtimeentity.RuntimeNode, 0)
	query := r.dbWithContext(ctx).
		Where("schedulable = ?", true).
		Where("health_status IN ?", []string{
			runtimeentity.RuntimeNodeHealthReady,
			runtimeentity.RuntimeNodeHealthDegraded,
		})
	if staleThreshold > 0 {
		if now.IsZero() {
			now = time.Now().UTC()
		}
		query = query.
			Where("last_seen_at IS NOT NULL").
			Where("last_seen_at >= ?", now.UTC().Add(-staleThreshold))
	}
	if err := query.Order("id ASC").Find(&nodes).Error; err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}
	return nodes, nil
}

func (r *RuntimeNodeRepository) FindByID(ctx context.Context, nodeID int64) (*runtimeentity.RuntimeNode, error) {
	if r == nil || r.db == nil || nodeID <= 0 {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}

	var node runtimeentity.RuntimeNode
	if err := r.dbWithContext(ctx).Where("id = ?", nodeID).First(&node).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, runtimeports.ErrRuntimeNodeUnavailable
		}
		return nil, err
	}
	return &node, nil
}

func (r *RuntimeNodeRepository) FindHealthyByID(ctx context.Context, nodeID int64, staleThreshold time.Duration, now time.Time) (*runtimeentity.RuntimeNode, error) {
	if r == nil || r.db == nil || nodeID <= 0 {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}

	var node runtimeentity.RuntimeNode
	query := r.dbWithContext(ctx).
		Where("id = ?", nodeID).
		Where("health_status IN ?", []string{
			runtimeentity.RuntimeNodeHealthReady,
			runtimeentity.RuntimeNodeHealthDegraded,
		})
	if staleThreshold > 0 {
		if now.IsZero() {
			now = time.Now().UTC()
		}
		query = query.
			Where("last_seen_at IS NOT NULL").
			Where("last_seen_at >= ?", now.UTC().Add(-staleThreshold))
	}
	if err := query.First(&node).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, runtimeports.ErrRuntimeNodeUnavailable
		}
		return nil, err
	}
	return &node, nil
}

func (r *RuntimeNodeRepository) FindSchedulableNodeByName(ctx context.Context, name string) (*runtimeentity.RuntimeNode, error) {
	if r == nil || r.db == nil {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}

	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}

	var node runtimeentity.RuntimeNode
	if err := r.dbWithContext(ctx).
		Where("name = ? AND schedulable = ?", trimmedName, true).
		First(&node).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, runtimeports.ErrRuntimeNodeUnavailable
		}
		return nil, err
	}
	return &node, nil
}

func (r *RuntimeNodeRepository) FindSchedulableHealthyNodeByName(ctx context.Context, name string, staleThreshold time.Duration, now time.Time) (*runtimeentity.RuntimeNode, error) {
	if r == nil || r.db == nil {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}

	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}

	var node runtimeentity.RuntimeNode
	query := r.dbWithContext(ctx).
		Where("name = ? AND schedulable = ?", trimmedName, true).
		Where("health_status IN ?", []string{
			runtimeentity.RuntimeNodeHealthReady,
			runtimeentity.RuntimeNodeHealthDegraded,
		})
	if staleThreshold > 0 {
		if now.IsZero() {
			now = time.Now().UTC()
		}
		query = query.
			Where("last_seen_at IS NOT NULL").
			Where("last_seen_at >= ?", now.UTC().Add(-staleThreshold))
	}
	if err := query.First(&node).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, runtimeports.ErrRuntimeNodeUnavailable
		}
		return nil, err
	}
	return &node, nil
}

func (r *RuntimeNodeRepository) ListHealthCheckNodes(ctx context.Context) ([]runtimeentity.RuntimeNode, error) {
	if r == nil || r.db == nil {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}

	nodes := make([]runtimeentity.RuntimeNode, 0)
	if err := r.dbWithContext(ctx).
		Order("id ASC").
		Find(&nodes).Error; err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}
	return nodes, nil
}

func (r *RuntimeNodeRepository) ListSchedulableNodes(ctx context.Context) ([]runtimeentity.RuntimeNode, error) {
	if r == nil || r.db == nil {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}

	nodes := make([]runtimeentity.RuntimeNode, 0)
	if err := r.dbWithContext(ctx).
		Where("schedulable = ?", true).
		Order("id ASC").
		Find(&nodes).Error; err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}
	return nodes, nil
}

func (r *RuntimeNodeRepository) MarkNodeHeartbeat(ctx context.Context, nodeID int64, healthStatus, capacitySnapshot string, seenAt time.Time) (*runtimeentity.RuntimeNode, error) {
	if r == nil || r.db == nil || nodeID <= 0 {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}
	if seenAt.IsZero() {
		seenAt = time.Now().UTC()
	}
	seenAt = seenAt.UTC()
	healthStatus = normalizeRuntimeNodeHealthStatus(healthStatus)
	capacitySnapshot = strings.TrimSpace(capacitySnapshot)
	if capacitySnapshot == "" {
		capacitySnapshot = "{}"
	}

	result := r.dbWithContext(ctx).Model(&runtimeentity.RuntimeNode{}).
		Where("id = ?", nodeID).
		Updates(map[string]any{
			"health_status":     healthStatus,
			"capacity_snapshot": capacitySnapshot,
			"last_seen_at":      seenAt,
			"updated_at":        seenAt,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}
	return r.FindByID(ctx, nodeID)
}

func (r *RuntimeNodeRepository) MarkNodeOffline(ctx context.Context, nodeID int64, updatedAt time.Time) (*runtimeentity.RuntimeNode, error) {
	if r == nil || r.db == nil || nodeID <= 0 {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	updatedAt = updatedAt.UTC()
	result := r.dbWithContext(ctx).Model(&runtimeentity.RuntimeNode{}).
		Where("id = ?", nodeID).
		Updates(map[string]any{
			"health_status": runtimeentity.RuntimeNodeHealthOffline,
			"updated_at":    updatedAt,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}
	return r.FindByID(ctx, nodeID)
}

func normalizeRuntimeNodeHealthStatus(status string) string {
	switch strings.TrimSpace(status) {
	case runtimeentity.RuntimeNodeHealthReady:
		return runtimeentity.RuntimeNodeHealthReady
	case runtimeentity.RuntimeNodeHealthDegraded:
		return runtimeentity.RuntimeNodeHealthDegraded
	case runtimeentity.RuntimeNodeHealthOffline:
		return runtimeentity.RuntimeNodeHealthOffline
	default:
		return runtimeentity.RuntimeNodeHealthUnknown
	}
}

type defaultRuntimeNodeSelector struct {
	repo            *RuntimeNodeRepository
	defaultNodeName string
	staleThreshold  time.Duration
}

func NewDefaultRuntimeNodeSelector(repo *RuntimeNodeRepository, defaultNodeName string, staleThreshold ...time.Duration) runtimeports.RuntimeNodeSelector {
	if repo == nil {
		return nil
	}
	threshold := time.Duration(0)
	if len(staleThreshold) > 0 {
		threshold = staleThreshold[0]
	}
	return &defaultRuntimeNodeSelector{
		repo:            repo,
		defaultNodeName: strings.TrimSpace(defaultNodeName),
		staleThreshold:  threshold,
	}
}

func (s *defaultRuntimeNodeSelector) SelectDefaultNode(ctx context.Context) (*runtimecontracts.RuntimeNodeBinding, error) {
	if s == nil || s.repo == nil {
		return nil, runtimeports.ErrRuntimeNodeUnavailable
	}

	var (
		node *runtimeentity.RuntimeNode
		err  error
	)
	if s.staleThreshold > 0 {
		if s.defaultNodeName != "" {
			node, err = s.repo.FindSchedulableHealthyNodeByName(ctx, s.defaultNodeName, s.staleThreshold, time.Now().UTC())
			if err == nil {
				return runtimeNodeBindingFromEntity(node), nil
			}
			if !errors.Is(err, runtimeports.ErrRuntimeNodeUnavailable) {
				return nil, err
			}
		}
		nodes, err := s.repo.ListSchedulableHealthyNodes(ctx, s.staleThreshold, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		return runtimeNodeBindingFromEntity(&nodes[0]), nil
	}

	if s.defaultNodeName != "" {
		node, err = s.repo.FindSchedulableNodeByName(ctx, s.defaultNodeName)
	} else {
		node, err = s.repo.FindFirstSchedulableNode(ctx)
	}
	if err != nil {
		return nil, err
	}
	return runtimeNodeBindingFromEntity(node), nil
}

func runtimeNodeBindingFromEntity(node *runtimeentity.RuntimeNode) *runtimecontracts.RuntimeNodeBinding {
	if node == nil {
		return nil
	}
	return &runtimecontracts.RuntimeNodeBinding{
		NodeID:   node.ID,
		NodeName: node.Name,
	}
}
