package runtime

import (
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsqry "ctf-platform/internal/module/ops/application/queries"
	opsinfra "ctf-platform/internal/module/ops/infrastructure"
	opsports "ctf-platform/internal/module/ops/ports"
	platformevents "ctf-platform/internal/platform/events"
	websocketpkg "ctf-platform/pkg/websocket"
)

type Module struct {
	AuditService        auditlog.Recorder
	AuditHandler        *opshttp.AuditHandler
	DashboardHandler    *opshttp.DashboardHandler
	NotificationHandler *opshttp.NotificationHandler
	RiskHandler         *opshttp.RiskHandler
	WebSocketManager    *websocketpkg.Manager

	notificationBuilder func(authcontracts.TokenService) *opshttp.NotificationHandler
}

type Deps struct {
	Config       *config.Config
	Logger       *zap.Logger
	DB           *gorm.DB
	Cache        *redislib.Client
	Events       platformevents.Bus
	RuntimeQuery opsports.RuntimeQuery
	RuntimeStats opsports.RuntimeStatsProvider
}

type moduleDeps struct {
	input Deps
	// auditRepo        opsports.AuditRepository
	auditRepo interface {
		opsports.AuditCommandRepository
		opsports.AuditQueryRepository
	}
	// riskRepo         opsports.RiskRepository
	riskRepo interface {
		opsports.RiskSubmitEventRepository
		opsports.RiskLoginEventRepository
	}
	// notificationRepo opsports.NotificationRepository
	notificationRepo interface {
		opsports.NotificationCommandRepository
		opsports.NotificationQueryRepository
	}
	runtimeQuery     opsports.RuntimeQuery
	runtimeStats     opsports.RuntimeStatsProvider
	webSocketManager *websocketpkg.Manager
}

func Build(deps Deps) *Module {
	internalDeps := newModuleDeps(deps)
	auditHandler, auditService := buildAuditHandler(internalDeps)
	dashboardHandler := buildDashboardHandler(internalDeps)
	riskHandler := buildRiskHandler(internalDeps)

	return &Module{
		AuditService:     auditService,
		AuditHandler:     auditHandler,
		DashboardHandler: dashboardHandler,
		RiskHandler:      riskHandler,
		WebSocketManager: internalDeps.webSocketManager,
		notificationBuilder: func(tokenService authcontracts.TokenService) *opshttp.NotificationHandler {
			return buildNotificationHandler(internalDeps, tokenService)
		},
	}
}

func (m *Module) BindNotificationHandler(tokenService authcontracts.TokenService) {
	if m == nil || m.notificationBuilder == nil {
		return
	}
	m.NotificationHandler = m.notificationBuilder(tokenService)
}

func newModuleDeps(deps Deps) moduleDeps {
	cfg := deps.Config
	log := deps.Logger

	return moduleDeps{
		input:            deps,
		auditRepo:        opsinfra.NewAuditRepository(deps.DB),
		riskRepo:         opsinfra.NewRiskRepository(deps.DB),
		notificationRepo: opsinfra.NewNotificationRepository(deps.DB),
		runtimeQuery:     deps.RuntimeQuery,
		runtimeStats:     deps.RuntimeStats,
		webSocketManager: websocketpkg.NewManager(cfg.WebSocket, log.Named("websocket_manager")),
	}
}

func buildAuditHandler(deps moduleDeps) (*opshttp.AuditHandler, auditlog.Recorder) {
	cfg := deps.input.Config
	log := deps.input.Logger

	auditCommandService := opscmd.NewAuditService(deps.auditRepo, log.Named("audit_command_service"))
	auditQueryService := opsqry.NewAuditService(deps.auditRepo, cfg.Pagination, log.Named("audit_query_service"))
	return opshttp.NewAuditHandler(auditQueryService), auditCommandService
}

func buildDashboardHandler(deps moduleDeps) *opshttp.DashboardHandler {
	cfg := deps.input.Config
	log := deps.input.Logger

	dashboardService := opsqry.NewDashboardService(
		deps.runtimeQuery,
		deps.runtimeStats,
		deps.input.Cache,
		cfg,
		log.Named("dashboard_service"),
	)
	return opshttp.NewDashboardHandler(dashboardService)
}

func buildRiskHandler(deps moduleDeps) *opshttp.RiskHandler {
	log := deps.input.Logger
	riskService := opsqry.NewRiskService(deps.riskRepo, log.Named("risk_service"))
	return opshttp.NewRiskHandler(riskService)
}

func buildNotificationHandler(deps moduleDeps, tokenService authcontracts.TokenService) *opshttp.NotificationHandler {
	cfg := deps.input.Config
	log := deps.input.Logger

	notificationCommandService := opscmd.NewNotificationService(
		deps.notificationRepo,
		cfg.Pagination,
		deps.webSocketManager,
		log.Named("notification_command_service"),
	)
	notificationQueryService := opsqry.NewNotificationService(
		deps.notificationRepo,
		cfg.Pagination,
		log.Named("notification_query_service"),
	)
	notificationCommandService.RegisterPracticeEventConsumers(deps.input.Events)
	notificationCommandService.RegisterChallengeEventConsumers(deps.input.Events)
	return opshttp.NewNotificationHandler(
		notificationCommandService,
		notificationQueryService,
		tokenService,
		deps.webSocketManager,
		log.Named("notification_handler"),
	)
}
