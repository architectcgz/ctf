package composition

import (
	teachinghttp "ctf-platform/internal/module/teaching_readmodel/api/http"
	readmodelqueries "ctf-platform/internal/module/teaching_readmodel/application/queries"
	readmodelinfra "ctf-platform/internal/module/teaching_readmodel/infrastructure"
)

type TeachingReadmodelModule struct {
	Handler *teachinghttp.Handler
	Query   readmodelqueries.Service
}

type teachingReadmodelModuleDeps struct {
	repo *readmodelinfra.Repository
}

func BuildTeachingReadmodelModule(root *Root, assessment *AssessmentModule) *TeachingReadmodelModule {
	log := root.Logger()

	deps := buildTeachingReadmodelModuleDeps(root)
	service := readmodelqueries.NewQueryService(
		deps.repo,
		assessment.Recommendations,
		root.Config().Pagination,
		log.Named("teaching_readmodel_query_service"),
	)

	return &TeachingReadmodelModule{
		Handler: teachinghttp.NewHandler(service),
		Query:   service,
	}
}

func buildTeachingReadmodelModuleDeps(root *Root) teachingReadmodelModuleDeps {
	return teachingReadmodelModuleDeps{
		repo: readmodelinfra.NewRepository(root.DB()),
	}
}
