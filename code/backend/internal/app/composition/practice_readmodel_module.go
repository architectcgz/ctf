package composition

import (
	practicereadmodelhttp "ctf-platform/internal/module/practice_readmodel/api/http"
	practicereadmodelqueries "ctf-platform/internal/module/practice_readmodel/application/queries"
	practicereadmodelinfra "ctf-platform/internal/module/practice_readmodel/infrastructure"
)

type PracticeReadmodelModule struct {
	Handler *practicereadmodelhttp.Handler
	Query   practicereadmodelqueries.Service
}

func BuildPracticeReadmodelModule(root *Root) *PracticeReadmodelModule {
	log := root.Logger()
	cfg := root.Config()
	db := root.DB()
	cache := root.Cache()

	repo := practicereadmodelinfra.NewRepository(db)
	service := practicereadmodelqueries.NewQueryService(repo, cache, cfg.Cache.ProgressTTL, log.Named("practice_readmodel_query_service"))

	return &PracticeReadmodelModule{
		Handler: practicereadmodelhttp.NewHandler(service),
		Query:   service,
	}
}
