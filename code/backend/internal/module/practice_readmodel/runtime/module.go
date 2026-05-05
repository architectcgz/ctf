package runtime

import (
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	practicereadmodelhttp "ctf-platform/internal/module/practice_readmodel/api/http"
	practicereadmodelqueries "ctf-platform/internal/module/practice_readmodel/application/queries"
	practicereadmodelinfra "ctf-platform/internal/module/practice_readmodel/infrastructure"
	practicereadmodelports "ctf-platform/internal/module/practice_readmodel/ports"
)

type Module struct {
	Handler *practicereadmodelhttp.Handler
	Query   practicereadmodelqueries.Service
}

type Deps struct {
	Config *config.Config
	Logger *zap.Logger
	DB     *gorm.DB
	Cache  *redislib.Client
}

type moduleDeps struct {
	input Deps
	// repo  practicereadmodelports.QueryRepository
	repo interface {
		practicereadmodelports.ProgressQueryRepository
		practicereadmodelports.TimelineQueryRepository
	}
}

func Build(deps Deps) *Module {
	internalDeps := newModuleDeps(deps)
	service := buildQueryService(internalDeps)

	return &Module{
		Handler: practicereadmodelhttp.NewHandler(service),
		Query:   service,
	}
}

func newModuleDeps(deps Deps) moduleDeps {
	return moduleDeps{
		input: deps,
		repo:  practicereadmodelinfra.NewRepository(deps.DB),
	}
}

func buildQueryService(deps moduleDeps) practicereadmodelqueries.Service {
	cfg := deps.input.Config
	log := deps.input.Logger
	return practicereadmodelqueries.NewQueryService(deps.repo, deps.input.Cache, cfg.Cache.ProgressTTL, log.Named("practice_readmodel_query_service"))
}
