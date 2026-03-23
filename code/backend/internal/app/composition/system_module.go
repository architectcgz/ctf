package composition

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/module/identity"
	"ctf-platform/internal/module/ops"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	"ctf-platform/internal/module/system"
	websocketpkg "ctf-platform/pkg/websocket"
)

type SystemModule struct {
	AuditService        ops.AuditRecorder
	AuditHandler        ops.AuditLogHandler
	DashboardHandler    ops.DashboardHandler
	NotificationHandler *system.NotificationHandler
	RiskHandler         ops.RiskHandler
	WebSocketManager    *websocketpkg.Manager
}

type runtimeSystemQuery interface {
	CountRunning() (int64, error)
}

type runtimeSystemStatsProvider interface {
	ListManagedContainerStats(ctx context.Context) ([]ops.ManagedContainerStat, error)
}

func BuildSystemModule(root *Root, runtime *RuntimeModule) *SystemModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	auditRepo := ops.NewAuditRepository(db)
	auditService := ops.NewAuditService(auditRepo, cfg.Pagination, log.Named("audit_service"))
	dashboardService := ops.NewDashboardService(
		runtime.system.query,
		runtime.system.statsProvider,
		cache,
		cfg,
		log.Named("dashboard_service"),
	)
	riskRepo := ops.NewRiskRepository(db)
	riskService := ops.NewRiskService(riskRepo, log.Named("risk_service"))

	return &SystemModule{
		AuditService:     ops.NewModule(auditService),
		AuditHandler:     ops.NewAuditHandler(auditService),
		DashboardHandler: ops.NewDashboardHandler(dashboardService),
		RiskHandler:      ops.NewRiskHandler(riskService),
		WebSocketManager: websocketpkg.NewManager(cfg.WebSocket, log.Named("websocket_manager")),
	}
}

func (m *SystemModule) BuildNotificationHandler(root *Root, tokenService identity.Authenticator) {
	if m == nil {
		return
	}

	cfg := root.Config()
	log := root.Logger()
	db := root.DB()

	notificationRepo := system.NewNotificationRepository(db)
	notificationService := system.NewNotificationService(
		notificationRepo,
		cfg.Pagination,
		m.WebSocketManager,
		log.Named("notification_service"),
	)
	notificationService.RegisterPracticeEventConsumers(root.Events)
	m.NotificationHandler = system.NewNotificationHandler(
		notificationService,
		tokenService,
		m.WebSocketManager,
		log.Named("notification_handler"),
	)
}

func NamedAuditLogger(log *zap.Logger) *zap.Logger {
	return log.Named("audit_middleware")
}

type systemRuntimeStatsProvider struct {
	service *runtimeapp.ContainerStatsService
}

func newSystemRuntimeStatsProvider(service *runtimeapp.ContainerStatsService) *systemRuntimeStatsProvider {
	return &systemRuntimeStatsProvider{service: service}
}

func (p *systemRuntimeStatsProvider) ListManagedContainerStats(ctx context.Context) ([]ops.ManagedContainerStat, error) {
	if p == nil || p.service == nil {
		return []ops.ManagedContainerStat{}, nil
	}

	stats, err := p.service.ListManagedContainerStats(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ops.ManagedContainerStat, 0, len(stats))
	for _, item := range stats {
		result = append(result, ops.ManagedContainerStat{
			ContainerID:   item.ContainerID,
			ContainerName: item.ContainerName,
			CPUPercent:    item.CPUPercent,
			MemoryPercent: item.MemoryPercent,
			MemoryUsage:   item.MemoryUsage,
			MemoryLimit:   item.MemoryLimit,
		})
	}
	return result, nil
}
