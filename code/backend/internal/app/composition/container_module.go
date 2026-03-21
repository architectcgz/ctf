package composition

import (
	"go.uber.org/zap"

	"ctf-platform/internal/module/container"
	"ctf-platform/internal/module/runtime"
)

type ContainerModule struct {
	Handler            *container.Handler
	ProxyTicketService *container.ProxyTicketService
	Repository         *container.Repository
	Service            *runtime.Module
}

func BuildContainerModule(root *Root) (*ContainerModule, error) {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	repo := container.NewRepository(db)
	var runtimeEngine *container.Engine
	if cfg.App.Env == "test" {
		log.Info("container_engine_disabled_in_test_env_for_router")
	} else if engine, err := container.NewEngine(&cfg.Container); err != nil {
		log.Warn("container_engine_init_failed_for_router", zap.Error(err))
	} else {
		runtimeEngine = engine
	}

	service := runtime.NewModule(container.NewService(repo, runtimeEngine, &cfg.Container, log.Named("container_service")))

	return &ContainerModule{
		ProxyTicketService: container.NewProxyTicketService(cache, &cfg.Container),
		Repository:         repo,
		Service:            service,
	}, nil
}

func (m *ContainerModule) BuildHandler(root *Root, system *SystemModule) {
	if m == nil {
		return
	}

	cfg := root.Config()
	m.Handler = container.NewHandler(
		m.Service.Service,
		m.ProxyTicketService,
		system.AuditService,
		container.ProxyCookieConfig{
			Secure:   cfg.Auth.RefreshCookieSecure,
			SameSite: cfg.Auth.CookieSameSite(),
		},
	)
}
