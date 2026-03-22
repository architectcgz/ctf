package composition

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/module/ops"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	"ctf-platform/internal/module/system"
	websocketpkg "ctf-platform/pkg/websocket"
)

type SystemModule struct {
	AuditService        ops.AuditRecorder
	AuditHandler        *system.AuditHandler
	DashboardHandler    *system.DashboardHandler
	NotificationHandler *system.NotificationHandler
	RiskHandler         *system.RiskHandler
	WebSocketManager    *websocketpkg.Manager
}

type runtimeSystemQuery interface {
	CountRunning() (int64, error)
}

type runtimeSystemStatsProvider interface {
	ListManagedContainerStats(ctx context.Context) ([]system.ManagedContainerStat, error)
}

func BuildSystemModule(root *Root, runtime *RuntimeModule) *SystemModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	auditRepo := system.NewAuditRepository(db)
	auditService := system.NewAuditService(auditRepo, cfg.Pagination, log.Named("audit_service"))
	dashboardService := system.NewDashboardService(
		runtime.system.query,
		runtime.system.statsProvider,
		cache,
		cfg,
		log.Named("dashboard_service"),
	)
	riskRepo := system.NewRiskRepository(db)
	riskService := system.NewRiskService(riskRepo, log.Named("risk_service"))

	return &SystemModule{
		AuditService:     ops.NewModule(auditService),
		AuditHandler:     system.NewAuditHandler(auditService),
		DashboardHandler: system.NewDashboardHandler(dashboardService),
		RiskHandler:      system.NewRiskHandler(riskService),
		WebSocketManager: websocketpkg.NewManager(cfg.WebSocket, log.Named("websocket_manager")),
	}
}

func (m *SystemModule) BuildNotificationHandler(root *Root, auth *AuthModule) {
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
		auth.TokenService,
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

func (p *systemRuntimeStatsProvider) ListManagedContainerStats(ctx context.Context) ([]system.ManagedContainerStat, error) {
	if p == nil || p.service == nil {
		return []system.ManagedContainerStat{}, nil
	}

	stats, err := p.service.ListManagedContainerStats(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]system.ManagedContainerStat, 0, len(stats))
	for _, item := range stats {
		result = append(result, system.ManagedContainerStat{
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
