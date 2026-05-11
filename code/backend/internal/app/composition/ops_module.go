package composition

import (
	"go.uber.org/zap"

	"ctf-platform/internal/auditlog"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opsruntime "ctf-platform/internal/module/ops/runtime"
	websocketpkg "ctf-platform/pkg/websocket"
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
		Config:       root.Config(),
		Logger:       root.Logger(),
		DB:           root.DB(),
		Cache:        root.Cache(),
		Events:       root.Events,
		RuntimeQuery: runtime.OpsRuntimeQuery,
		RuntimeStats: runtime.OpsRuntimeStatsProvider,
	})
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
