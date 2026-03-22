package composition

import (
	"context"

	runtimeModule "ctf-platform/internal/module/runtime"
	runtimehttp "ctf-platform/internal/module/runtime/api/http"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	runtimeinfra "ctf-platform/internal/module/runtimeinfra"
)

type RuntimeModule struct {
	Handler *runtimehttp.Handler

	service *runtimeModule.Module
	query   *runtimeapp.QueryService
}

func BuildRuntimeModule(root *Root, infra *RuntimeInfraModule) *RuntimeModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()
	repo := runtimeinfrarepo.NewRepository(db)
	baseService := runtimeModule.NewService(repo, infra.Engine, &cfg.Container, log.Named("runtime_service"))
	instanceService := runtimeapp.NewInstanceService(repo, baseService, &cfg.Container, log.Named("runtime_instance_service"))
	proxyTicketService := runtimeapp.NewProxyTicketService(runtimeinfrarepo.NewProxyTicketStore(cache), cfg.Container.ProxyTicketTTL)
	cleaner := runtimeinfra.NewCleaner(baseService, cache, cfg.Container.CleanupLockTTL, log.Named("runtime_cleaner"))
	root.RegisterBackgroundJob(NewBackgroundJob(
		"runtime_cleaner",
		func(context.Context) error {
			return cleaner.Start(cfg.Container.CleanupInterval)
		},
		cleaner.Stop,
	))

	service := runtimeModule.NewModule(
		baseService,
		repo,
		instanceService,
		proxyTicketService,
		cfg.Container.ProxyBodyPreviewSize,
	)

	return &RuntimeModule{
		service: service,
		query:   runtimeapp.NewQueryService(repo),
	}
}

func (m *RuntimeModule) BuildHandler(root *Root, system *SystemModule) {
	if m == nil {
		return
	}

	cfg := root.Config()
	m.Handler = runtimehttp.NewHandler(
		m.service,
		system.AuditService,
		runtimehttp.CookieConfig{
			Secure:   cfg.Auth.RefreshCookieSecure,
			SameSite: cfg.Auth.CookieSameSite(),
		},
	)
}
