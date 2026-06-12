package runtime

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/config"
	websocketpkg "ctf-platform/internal/infrastructure/websocket"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	contestports "ctf-platform/internal/module/contest/ports"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsqry "ctf-platform/internal/module/ops/application/queries"
	opsinfra "ctf-platform/internal/module/ops/infrastructure"
	opsports "ctf-platform/internal/module/ops/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type BackgroundJob struct {
	Name string
	Run  func(context.Context)
}

type Module struct {
	AuditService        auditlog.Recorder
	AuditHandler        *opshttp.AuditHandler
	DashboardHandler    *opshttp.DashboardHandler
	NotificationHandler *opshttp.NotificationHandler
	RiskHandler         *opshttp.RiskHandler
	WebSocketManager    *websocketpkg.Manager
	BackgroundJobs      []BackgroundJob

	notificationBuilder        func(authcontracts.TokenService) (*opshttp.NotificationHandler, *opscmd.NotificationService)
	notificationCommandService *opscmd.NotificationService
}

type Deps struct {
	Config                *config.Config
	Logger                *zap.Logger
	DB                    *gorm.DB
	Cache                 *redislib.Client
	Events                platformevents.Bus
	RuntimeQuery          opsports.RuntimeQuery
	RuntimeStats          opsports.RuntimeStatsProvider
	ContestRealtimeOutbox contestports.ContestRealtimeOutboxRepository
	OutboxHandlers        *platformevents.OutboxHandlerRegistry
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
		opsports.NotificationOutboxTxManager
	}
	contestRealtimeOutbox contestports.ContestRealtimeOutboxRepository
	outboxHandlers        *platformevents.OutboxHandlerRegistry
	dashboardState        opsports.DashboardStateStore
	runtimeQuery          opsports.RuntimeQuery
	runtimeStats          opsports.RuntimeStatsProvider
	webSocketManager      *websocketpkg.Manager
}

func Build(deps Deps) *Module {
	internalDeps := newModuleDeps(deps)
	auditHandler, auditService := buildAuditHandler(internalDeps)
	dashboardHandler := buildDashboardHandler(internalDeps)
	riskHandler := buildRiskHandler(internalDeps)
	backgroundJobs := registerContestRealtimeConsumers(internalDeps)

	return &Module{
		AuditService:     auditService,
		AuditHandler:     auditHandler,
		DashboardHandler: dashboardHandler,
		RiskHandler:      riskHandler,
		WebSocketManager: internalDeps.webSocketManager,
		BackgroundJobs:   backgroundJobs,
		notificationBuilder: func(tokenService authcontracts.TokenService) (*opshttp.NotificationHandler, *opscmd.NotificationService) {
			return buildNotificationHandler(internalDeps, tokenService)
		},
	}
}

func newModuleDeps(deps Deps) moduleDeps {
	cfg := deps.Config
	log := deps.Logger

	return moduleDeps{
		input:                 deps,
		auditRepo:             opsinfra.NewAuditRepository(deps.DB),
		riskRepo:              opsinfra.NewRiskRepository(deps.DB),
		notificationRepo:      opsinfra.NewNotificationRepository(deps.DB),
		contestRealtimeOutbox: deps.ContestRealtimeOutbox,
		outboxHandlers:        deps.OutboxHandlers,
		dashboardState:        opsinfra.NewDashboardStateStore(deps.Cache, deps.Config, log.Named("dashboard_state_store")),
		runtimeQuery:          deps.RuntimeQuery,
		runtimeStats:          deps.RuntimeStats,
		webSocketManager:      websocketpkg.NewManager(cfg.WebSocket, log.Named("websocket_manager")),
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
		deps.dashboardState,
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
