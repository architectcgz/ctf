package composition

import (
	"go.uber.org/zap"

	"ctf-platform/internal/module/ops"
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

func BuildSystemModule(root *Root, container *ContainerModule) *SystemModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	auditRepo := system.NewAuditRepository(db)
	auditService := system.NewAuditService(auditRepo, cfg.Pagination, log.Named("audit_service"))
	dashboardService := system.NewDashboardService(
		container.Query,
		container.Service,
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
