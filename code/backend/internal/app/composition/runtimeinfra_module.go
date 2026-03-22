package composition

import (
	"go.uber.org/zap"

	runtimeinfra "ctf-platform/internal/module/runtimeinfra"
)

type RuntimeInfraModule struct {
	Engine *runtimeinfra.Engine
}

func BuildRuntimeInfraModule(root *Root) (*RuntimeInfraModule, error) {
	cfg := root.Config()
	log := root.Logger()
	var runtimeEngine *runtimeinfra.Engine
	if cfg.App.Env == "test" {
		log.Info("runtime_engine_disabled_in_test_env_for_router")
	} else if engine, err := runtimeinfra.NewEngine(&cfg.Container); err != nil {
		log.Warn("runtime_engine_init_failed_for_router", zap.Error(err))
	} else {
		runtimeEngine = engine
	}

	return &RuntimeInfraModule{
		Engine: runtimeEngine,
	}, nil
}
