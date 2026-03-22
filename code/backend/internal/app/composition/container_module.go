package composition

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/module/container"
)

type ContainerModule struct {
	Repository *container.Repository
	Service    *container.Service
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

	service := container.NewService(repo, runtimeEngine, &cfg.Container, log.Named("container_service"))
	cleaner := container.NewCleaner(service, cache, cfg.Container.CleanupLockTTL, log.Named("container_cleaner"))
	root.RegisterBackgroundJob(NewBackgroundJob(
		"container_cleaner",
		func(context.Context) error {
			return cleaner.Start(cfg.Container.CleanupInterval)
		},
		cleaner.Stop,
	))

	return &ContainerModule{
		Repository: repo,
		Service:    service,
	}, nil
}
