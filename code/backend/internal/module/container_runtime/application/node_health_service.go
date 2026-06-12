package application

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.uber.org/zap"

	runtimeentity "ctf-platform/internal/module/container_runtime/entity"
)

const (
	defaultNodeHealthPollInterval = 10 * time.Second
	defaultNodeHealthProbeTimeout = 2 * time.Second
	defaultNodeHealthStaleAfter   = 30 * time.Second
	defaultNodeHealthFailures     = 3
)

type NodeHealthOptions struct {
	PollInterval     time.Duration
	ProbeTimeout     time.Duration
	StaleAfter       time.Duration
	FailureThreshold int
}

type NodeHealthRepository interface {
	ListHealthCheckNodes(ctx context.Context) ([]runtimeentity.RuntimeNode, error)
	MarkNodeHeartbeat(ctx context.Context, nodeID int64, healthStatus, capacitySnapshot string, seenAt time.Time) (*runtimeentity.RuntimeNode, error)
	MarkNodeOffline(ctx context.Context, nodeID int64, updatedAt time.Time) (*runtimeentity.RuntimeNode, error)
}

type NodeHealthProbe interface {
	ListManagedContainerStats(ctx context.Context, node runtimeentity.RuntimeNode) ([]ManagedContainerStat, error)
}

type NodeOfflineHandler func(ctx context.Context, node runtimeentity.RuntimeNode) error

type NodeHealthService struct {
	repo           NodeHealthRepository
	probe          NodeHealthProbe
	options        NodeHealthOptions
	logger         *zap.Logger
	now            func() time.Time
	offlineHandler NodeOfflineHandler

	failures       map[int64]int
	offlineHandled map[int64]bool
}

func NewNodeHealthService(repo NodeHealthRepository, probe NodeHealthProbe, options NodeHealthOptions, logger *zap.Logger) *NodeHealthService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &NodeHealthService{
		repo:           repo,
		probe:          probe,
		options:        normalizeNodeHealthOptions(options),
		logger:         logger,
		now:            func() time.Time { return time.Now().UTC() },
		failures:       make(map[int64]int),
		offlineHandled: make(map[int64]bool),
	}
}

func (s *NodeHealthService) SetNow(now func() time.Time) *NodeHealthService {
	if s == nil {
		return nil
	}
	if now != nil {
		s.now = now
	}
	return s
}

func (s *NodeHealthService) SetOfflineHandler(handler NodeOfflineHandler) *NodeHealthService {
	if s == nil {
		return nil
	}
	s.offlineHandler = handler
	return s
}

func (s *NodeHealthService) EvaluateOnce(ctx context.Context) error {
	ctx = normalizeContext(ctx)
	if s == nil || s.repo == nil {
		return nil
	}

	nodes, err := s.repo.ListHealthCheckNodes(ctx)
	if err != nil {
		return err
	}
	now := s.nowUTC()
	for _, node := range nodes {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if node.ID <= 0 {
			continue
		}
		if err := s.evaluateNode(ctx, node, now); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Warn("runtime_node_health_evaluate_failed",
				zap.Int64("node_id", node.ID),
				zap.String("node_name", node.Name),
				zap.Error(err))
		}
	}
	return nil
}

func (s *NodeHealthService) Run(ctx context.Context) {
	if ctx == nil {
		return
	}
	for {
		if err := s.EvaluateOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Warn("runtime_node_health_loop_failed", zap.Error(err))
		}

		timer := time.NewTimer(s.options.PollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (s *NodeHealthService) evaluateNode(ctx context.Context, node runtimeentity.RuntimeNode, now time.Time) error {
	if s == nil || s.probe == nil {
		return nil
	}

	probeCtx, cancel := context.WithTimeout(ctx, s.options.ProbeTimeout)
	defer cancel()
	stats, err := s.probe.ListManagedContainerStats(probeCtx, node)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		s.failures[node.ID]++
		if node.HealthStatus == runtimeentity.RuntimeNodeHealthOffline || s.shouldMarkOffline(node, now) || s.failures[node.ID] >= s.options.FailureThreshold {
			updatedNode, markErr := s.repo.MarkNodeOffline(ctx, node.ID, now)
			if markErr != nil {
				return markErr
			}
			if s.offlineHandled[node.ID] {
				return nil
			}
			offlineNode := node
			if updatedNode != nil {
				offlineNode = *updatedNode
			}
			if err := s.handleNodeOffline(ctx, offlineNode); err != nil {
				return err
			}
			s.offlineHandled[node.ID] = true
		}
		return nil
	}

	s.failures[node.ID] = 0
	delete(s.offlineHandled, node.ID)
	snapshot, err := buildNodeCapacitySnapshot(stats)
	if err != nil {
		return err
	}
	_, err = s.repo.MarkNodeHeartbeat(ctx, node.ID, runtimeentity.RuntimeNodeHealthReady, snapshot, now)
	return err
}

func (s *NodeHealthService) handleNodeOffline(ctx context.Context, node runtimeentity.RuntimeNode) error {
	if s == nil || s.offlineHandler == nil {
		return nil
	}
	return s.offlineHandler(ctx, node)
}

func (s *NodeHealthService) shouldMarkOffline(node runtimeentity.RuntimeNode, now time.Time) bool {
	if s == nil || s.options.StaleAfter <= 0 || node.LastSeenAt == nil || node.LastSeenAt.IsZero() {
		return false
	}
	return now.UTC().Sub(node.LastSeenAt.UTC()) > s.options.StaleAfter
}

func (s *NodeHealthService) nowUTC() time.Time {
	if s == nil || s.now == nil {
		return time.Now().UTC()
	}
	return s.now().UTC()
}

func normalizeNodeHealthOptions(options NodeHealthOptions) NodeHealthOptions {
	if options.PollInterval <= 0 {
		options.PollInterval = defaultNodeHealthPollInterval
	}
	if options.ProbeTimeout <= 0 {
		options.ProbeTimeout = defaultNodeHealthProbeTimeout
	}
	if options.StaleAfter <= 0 {
		options.StaleAfter = defaultNodeHealthStaleAfter
	}
	if options.FailureThreshold <= 0 {
		options.FailureThreshold = defaultNodeHealthFailures
	}
	return options
}

type nodeCapacitySnapshot struct {
	Containers       int     `json:"containers"`
	MemoryUsage      int64   `json:"memory_usage"`
	MemoryLimit      int64   `json:"memory_limit"`
	MaxCPUPercent    float64 `json:"max_cpu_percent"`
	MaxMemoryPercent float64 `json:"max_memory_percent"`
	UpdatedAt        string  `json:"updated_at,omitempty"`
}

func buildNodeCapacitySnapshot(stats []ManagedContainerStat) (string, error) {
	snapshot := nodeCapacitySnapshot{Containers: len(stats)}
	for _, item := range stats {
		snapshot.MemoryUsage += item.MemoryUsage
		snapshot.MemoryLimit += item.MemoryLimit
		if item.CPUPercent > snapshot.MaxCPUPercent {
			snapshot.MaxCPUPercent = item.CPUPercent
		}
		if item.MemoryPercent > snapshot.MaxMemoryPercent {
			snapshot.MaxMemoryPercent = item.MemoryPercent
		}
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}
