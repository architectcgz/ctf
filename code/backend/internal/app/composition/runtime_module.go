package composition

import (
	"context"

	runtimeModule "ctf-platform/internal/module/runtime"
	runtimeinfra "ctf-platform/internal/module/runtimeinfra"
)

type RuntimeModule struct {
	Handler    *runtimeModule.Handler
	Query      runtimeModule.RuntimeQuery
	Repository runtimeModule.InstanceRepository
	Service    runtimeModule.RuntimeFacade

	service *runtimeModule.Module
}

func BuildRuntimeModule(root *Root, infra *RuntimeInfraModule) *RuntimeModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()
	repo := runtimeModule.NewRepository(db)
	baseService := runtimeModule.NewService(repo, infra.Engine, &cfg.Container, log.Named("runtime_service"))
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
		runtimeModule.NewProxyTicketService(cache, &cfg.Container),
		cfg.Container.ProxyBodyPreviewSize,
	)

	return &RuntimeModule{
		Query:      runtimeModule.NewQuery(repo),
		Repository: repo,
		Service:    service,
		service:    service,
	}
}

func (m *RuntimeModule) BuildHandler(root *Root, system *SystemModule) {
	if m == nil {
		return
	}

	cfg := root.Config()
	m.Handler = runtimeModule.NewHandler(
		m.service,
		system.AuditService,
		runtimeModule.ProxyCookieConfig{
			Secure:   cfg.Auth.RefreshCookieSecure,
			SameSite: cfg.Auth.CookieSameSite(),
		},
	)
}
