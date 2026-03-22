package composition

import (
	practicereadmodel "ctf-platform/internal/module/practice_readmodel"
	practicereadmodelhttp "ctf-platform/internal/module/practice_readmodel/api/http"
	practicereadmodelapp "ctf-platform/internal/module/practice_readmodel/application"
	practicereadmodelinfra "ctf-platform/internal/module/practice_readmodel/infrastructure"
)

type PracticeReadmodelModule struct {
	Handler *practicereadmodelhttp.Handler
	Query   practicereadmodel.PracticeQuery
}

func BuildPracticeReadmodelModule(root *Root) *PracticeReadmodelModule {
	log := root.Logger()
	cfg := root.Config()
	db := root.DB()
	cache := root.Cache()

	repo := practicereadmodelinfra.NewRepository(db)
	service := practicereadmodelapp.NewQueryService(repo, cache, cfg.Cache.ProgressTTL, log.Named("practice_readmodel_query_service"))
	module := practicereadmodel.NewModule(service)

	return &PracticeReadmodelModule{
		Handler: practicereadmodelhttp.NewHandler(module),
		Query:   module,
	}
}
