package composition

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/auditlog"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsqry "ctf-platform/internal/module/ops/application/queries"
	opsinfra "ctf-platform/internal/module/ops/infrastructure"
	opsports "ctf-platform/internal/module/ops/ports"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	websocketpkg "ctf-platform/pkg/websocket"
)

type OpsModule struct {
	AuditService        auditlog.Recorder
	AuditHandler        *opshttp.AuditHandler
	DashboardHandler    *opshttp.DashboardHandler
	NotificationHandler *opshttp.NotificationHandler
	RiskHandler         *opshttp.RiskHandler
	WebSocketManager    *websocketpkg.Manager
}

type runtimeOpsQuery interface {
	CountRunning() (int64, error)
}

type runtimeOpsStatsProvider interface {
	ListManagedContainerStats(ctx context.Context) ([]opsports.ManagedContainerStat, error)
}

func BuildOpsModule(root *Root, runtime *RuntimeModule) *OpsModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	auditRepo := opsinfra.NewAuditRepository(db)
	auditCommandService := opscmd.NewAuditService(auditRepo, log.Named("audit_command_service"))
	auditQueryService := opsqry.NewAuditService(auditRepo, cfg.Pagination, log.Named("audit_query_service"))
	dashboardService := opsqry.NewDashboardService(
		runtime.ops.query,
		runtime.ops.statsProvider,
		cache,
		cfg,
		log.Named("dashboard_service"),
	)
	riskRepo := opsinfra.NewRiskRepository(db)
	riskService := opsqry.NewRiskService(riskRepo, log.Named("risk_service"))

	return &OpsModule{
		AuditService:     auditCommandService,
		AuditHandler:     opshttp.NewAuditHandler(auditQueryService),
		DashboardHandler: opshttp.NewDashboardHandler(dashboardService),
		RiskHandler:      opshttp.NewRiskHandler(riskService),
		WebSocketManager: websocketpkg.NewManager(cfg.WebSocket, log.Named("websocket_manager")),
	}
}

func (m *OpsModule) BuildNotificationHandler(root *Root, tokenService authcontracts.TokenService) {
	if m == nil {
		return
	}

	cfg := root.Config()
	log := root.Logger()
	db := root.DB()

	notificationRepo := opsinfra.NewNotificationRepository(db)
	notificationCommandService := opscmd.NewNotificationService(
		notificationRepo,
		cfg.Pagination,
		m.WebSocketManager,
		log.Named("notification_command_service"),
	)
	notificationQueryService := opsqry.NewNotificationService(
		notificationRepo,
		cfg.Pagination,
		log.Named("notification_query_service"),
	)
	notificationCommandService.RegisterPracticeEventConsumers(root.Events)
	m.NotificationHandler = opshttp.NewNotificationHandler(
		notificationCommandService,
		notificationQueryService,
		tokenService,
		m.WebSocketManager,
		log.Named("notification_handler"),
	)
}

func NamedAuditLogger(log *zap.Logger) *zap.Logger {
	return log.Named("audit_middleware")
}

type opsRuntimeStatsProvider struct {
	service *runtimeapp.ContainerStatsService
}

func newOpsRuntimeStatsProvider(service *runtimeapp.ContainerStatsService) *opsRuntimeStatsProvider {
	return &opsRuntimeStatsProvider{service: service}
}

func (p *opsRuntimeStatsProvider) ListManagedContainerStats(ctx context.Context) ([]opsports.ManagedContainerStat, error) {
	if p == nil || p.service == nil {
		return []opsports.ManagedContainerStat{}, nil
	}

	stats, err := p.service.ListManagedContainerStats(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]opsports.ManagedContainerStat, 0, len(stats))
	for _, item := range stats {
		result = append(result, opsports.ManagedContainerStat{
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
