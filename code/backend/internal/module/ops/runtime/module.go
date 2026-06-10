package runtime

import (
	"context"
	"os"
	"strings"
	"time"

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

	notificationBuilder func(authcontracts.TokenService) *opshttp.NotificationHandler
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
	contestRealtimeOutbox contestports.ContestRealtimeOutboxRepository
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
		input:                 deps,
		auditRepo:             opsinfra.NewAuditRepository(deps.DB),
		riskRepo:              opsinfra.NewRiskRepository(deps.DB),
		notificationRepo:      opsinfra.NewNotificationRepository(deps.DB),
		contestRealtimeOutbox: deps.ContestRealtimeOutbox,
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

func registerContestRealtimeConsumers(deps moduleDeps) []BackgroundJob {
	stream := opsinfra.NewContestRealtimeStream(deps.input.Cache, deps.webSocketManager, deps.input.Logger.Named("contest_realtime_stream"))
	relayService := opscmd.NewContestRealtimeService(stream)
	relayService.RegisterContestEventConsumers(deps.input.Events)

	dispatcher := opscmd.NewContestRealtimeOutboxDispatcher(deps.contestRealtimeOutbox, stream, deps.input.Logger.Named("contest_realtime_outbox_dispatcher"))
	consumerID := contestRealtimeConsumerID()
	return []BackgroundJob{
		{Name: "contest_realtime_outbox_dispatcher", Run: dispatcher.Run},
		{
			Name: "contest_realtime_stream_consumer",
			Run: func(ctx context.Context) {
				runContestRealtimeConsumer(ctx, stream, consumerID, deps.input.Logger.Named("contest_realtime_stream_consumer"))
			},
		},
	}
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

func runContestRealtimeConsumer(ctx context.Context, stream interface {
	ConsumeOnce(context.Context, string) error
}, consumerID string, logger *zap.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := stream.ConsumeOnce(ctx, consumerID); err != nil {
			if logger != nil {
				logger.Warn("consume contest realtime stream failed", zap.Error(err))
			}
			timer := time.NewTimer(250 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}
}

func contestRealtimeConsumerID() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "contest-realtime-consumer"
	}
	name := strings.TrimSpace(hostname)
	if name == "" {
		return "contest-realtime-consumer"
	}
	return name
}
