package composition

import (
	"go.uber.org/zap"

	"ctf-platform/internal/auditlog"
	websocketpkg "ctf-platform/internal/infrastructure/websocket"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opsruntime "ctf-platform/internal/module/ops/runtime"
)

type OpsModule struct {
	AuditService        auditlog.Recorder
	AuditHandler        *opshttp.AuditHandler
	DashboardHandler    *opshttp.DashboardHandler
	NotificationHandler *opshttp.NotificationHandler
	RiskHandler         *opshttp.RiskHandler
	WebSocketManager    *websocketpkg.Manager

	runtime *opsruntime.Module
}

func BuildOpsModule(root *Root, runtime *ContainerRuntimeModule) *OpsModule {
	module := opsruntime.Build(opsruntime.Deps{
		Config:                root.Config(),
		Logger:                root.Logger(),
		DB:                    root.DB(),
		Cache:                 root.Cache(),
		Events:                root.Events,
		RuntimeQuery:          runtime.OpsRuntimeQuery,
		RuntimeStats:          runtime.OpsRuntimeStatsProvider,
		ContestRealtimeOutbox: contestinfra.NewRealtimeOutboxRepository(root.DB()),
	})
	for _, job := range module.BackgroundJobs {
		root.RegisterBackgroundJob(NewLoopBackgroundJob(job.Name, job.Run))
	}
	return &OpsModule{
		AuditService:     module.AuditService,
		AuditHandler:     module.AuditHandler,
		DashboardHandler: module.DashboardHandler,
		RiskHandler:      module.RiskHandler,
		WebSocketManager: module.WebSocketManager,
		runtime:          module,
	}
}

func (m *OpsModule) BuildNotificationHandler(root *Root, tokenService authcontracts.TokenService) {
	if m == nil || m.runtime == nil {
		return
	}
	m.runtime.BindNotificationHandler(tokenService)
	m.NotificationHandler = m.runtime.NotificationHandler
}

func NamedAuditLogger(log *zap.Logger) *zap.Logger {
	return log.Named("audit_middleware")
}
