package composition

import (
	"go.uber.org/zap"

	"ctf-platform/internal/auditlog"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsqry "ctf-platform/internal/module/ops/application/queries"
	opsinfra "ctf-platform/internal/module/ops/infrastructure"
	opsports "ctf-platform/internal/module/ops/ports"
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

type opsModuleDeps struct {
	auditRepo        opsports.AuditRepository
	riskRepo         opsports.RiskRepository
	runtimeQuery     opsports.RuntimeQuery
	runtimeStats     opsports.RuntimeStatsProvider
	webSocketManager *websocketpkg.Manager
}

type opsNotificationDeps struct {
	notificationRepo opsports.NotificationRepository
	broadcaster      opsports.NotificationBroadcaster
	webSocketManager *websocketpkg.Manager
}

func BuildOpsModule(root *Root, runtime *RuntimeModule) *OpsModule {
	deps := buildOpsModuleDeps(root, runtime)
	auditHandler, auditService := buildOpsAuditHandler(root, deps)
	dashboardHandler := buildOpsDashboardHandler(root, deps)
	riskHandler := buildOpsRiskHandler(root, deps)

	return &OpsModule{
		AuditService:     auditService,
		AuditHandler:     auditHandler,
		DashboardHandler: dashboardHandler,
		RiskHandler:      riskHandler,
		WebSocketManager: deps.webSocketManager,
	}
}

func (m *OpsModule) BuildNotificationHandler(root *Root, tokenService authcontracts.TokenService) {
	if m == nil {
		return
	}

	deps := buildOpsNotificationDeps(root, m.WebSocketManager)
	m.NotificationHandler = buildOpsNotificationHandler(root, deps, tokenService)
}

func NamedAuditLogger(log *zap.Logger) *zap.Logger {
	return log.Named("audit_middleware")
}

func buildOpsModuleDeps(root *Root, runtime *RuntimeModule) opsModuleDeps {
	cfg := root.Config()
	log := root.Logger()
	return opsModuleDeps{
		auditRepo:        opsinfra.NewAuditRepository(root.DB()),
		riskRepo:         opsinfra.NewRiskRepository(root.DB()),
		runtimeQuery:     runtime.OpsRuntimeQuery,
		runtimeStats:     runtime.OpsRuntimeStatsProvider,
		webSocketManager: websocketpkg.NewManager(cfg.WebSocket, log.Named("websocket_manager")),
	}
}

func buildOpsAuditHandler(root *Root, deps opsModuleDeps) (*opshttp.AuditHandler, auditlog.Recorder) {
	cfg := root.Config()
	log := root.Logger()
	auditCommandService := opscmd.NewAuditService(deps.auditRepo, log.Named("audit_command_service"))
	auditQueryService := opsqry.NewAuditService(deps.auditRepo, cfg.Pagination, log.Named("audit_query_service"))
	return opshttp.NewAuditHandler(auditQueryService), auditCommandService
}

func buildOpsDashboardHandler(root *Root, deps opsModuleDeps) *opshttp.DashboardHandler {
	cfg := root.Config()
	log := root.Logger()
	dashboardService := opsqry.NewDashboardService(
		deps.runtimeQuery,
		deps.runtimeStats,
		root.Cache(),
		cfg,
		log.Named("dashboard_service"),
	)
	return opshttp.NewDashboardHandler(dashboardService)
}

func buildOpsRiskHandler(root *Root, deps opsModuleDeps) *opshttp.RiskHandler {
	log := root.Logger()
	riskService := opsqry.NewRiskService(deps.riskRepo, log.Named("risk_service"))
	return opshttp.NewRiskHandler(riskService)
}

func buildOpsNotificationDeps(root *Root, manager *websocketpkg.Manager) opsNotificationDeps {
	return opsNotificationDeps{
		notificationRepo: opsinfra.NewNotificationRepository(root.DB()),
		broadcaster:      manager,
		webSocketManager: manager,
	}
}

func buildOpsNotificationHandler(root *Root, deps opsNotificationDeps, tokenService authcontracts.TokenService) *opshttp.NotificationHandler {
	cfg := root.Config()
	log := root.Logger()
	notificationCommandService := opscmd.NewNotificationService(
		deps.notificationRepo,
		cfg.Pagination,
		deps.broadcaster,
		log.Named("notification_command_service"),
	)
	notificationQueryService := opsqry.NewNotificationService(
		deps.notificationRepo,
		cfg.Pagination,
		log.Named("notification_query_service"),
	)
	notificationCommandService.RegisterPracticeEventConsumers(root.Events)
	return opshttp.NewNotificationHandler(
		notificationCommandService,
		notificationQueryService,
		tokenService,
		deps.webSocketManager,
		log.Named("notification_handler"),
	)
}
