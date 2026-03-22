package composition

import runtimeModule "ctf-platform/internal/module/runtime"

type RuntimeModule struct {
	Handler    *runtimeModule.Handler
	Query      runtimeModule.RuntimeQuery
	Repository runtimeModule.InstanceRepository
	Service    runtimeModule.RuntimeFacade

	service *runtimeModule.Module
}

func BuildRuntimeModule(root *Root, container *ContainerModule) *RuntimeModule {
	cfg := root.Config()
	cache := root.Cache()

	service := runtimeModule.NewModule(
		container.Service,
		runtimeModule.NewProxyTicketService(cache, &cfg.Container),
		cfg.Container.ProxyBodyPreviewSize,
	)

	return &RuntimeModule{
		Query:      runtimeModule.NewQuery(container.Repository),
		Repository: container.Repository,
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
