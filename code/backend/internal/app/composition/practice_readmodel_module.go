package composition

import (
	practicereadmodelhttp "ctf-platform/internal/module/practice_readmodel/api/http"
	practicereadmodelqueries "ctf-platform/internal/module/practice_readmodel/application/queries"
	practicereadmodelinfra "ctf-platform/internal/module/practice_readmodel/infrastructure"
	practicereadmodelports "ctf-platform/internal/module/practice_readmodel/ports"
)

type PracticeReadmodelModule struct {
	Handler *practicereadmodelhttp.Handler
	Query   practicereadmodelqueries.Service
}

type practiceReadmodelModuleDeps struct {
	repo practicereadmodelports.QueryRepository
}

func BuildPracticeReadmodelModule(root *Root) *PracticeReadmodelModule {
	log := root.Logger()
	cfg := root.Config()
	cache := root.Cache()

	deps := buildPracticeReadmodelModuleDeps(root)
	service := practicereadmodelqueries.NewQueryService(deps.repo, cache, cfg.Cache.ProgressTTL, log.Named("practice_readmodel_query_service"))

	return &PracticeReadmodelModule{
		Handler: practicereadmodelhttp.NewHandler(service),
		Query:   service,
	}
}

func buildPracticeReadmodelModuleDeps(root *Root) practiceReadmodelModuleDeps {
	return practiceReadmodelModuleDeps{
		repo: practicereadmodelinfra.NewRepository(root.DB()),
	}
}
