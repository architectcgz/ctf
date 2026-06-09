package infrastructure

import (
	"context"
	"errors"
	"strings"
	"time"

	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
	runtimeports "ctf-platform/internal/module/runtime/ports"

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
	now := time.Now().UTC()

	var node runtimeentity.RuntimeNode
	err := r.dbWithContext(ctx).Where("name = ?", name).First(&node).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		node = runtimeentity.RuntimeNode{
			Name:             name,
			Endpoint:         endpoint,
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

type defaultRuntimeNodeSelector struct {
	repo            *RuntimeNodeRepository
	defaultNodeName string
}

func NewDefaultRuntimeNodeSelector(repo *RuntimeNodeRepository, defaultNodeName string) runtimeports.RuntimeNodeSelector {
	if repo == nil {
		return nil
	}
	return &defaultRuntimeNodeSelector{
		repo:            repo,
		defaultNodeName: strings.TrimSpace(defaultNodeName),
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
	if s.defaultNodeName != "" {
		node, err = s.repo.FindSchedulableNodeByName(ctx, s.defaultNodeName)
	} else {
		node, err = s.repo.FindFirstSchedulableNode(ctx)
	}
	if err != nil {
		return nil, err
	}
	return &runtimecontracts.RuntimeNodeBinding{
		NodeID:   node.ID,
		NodeName: node.Name,
	}, nil
}
