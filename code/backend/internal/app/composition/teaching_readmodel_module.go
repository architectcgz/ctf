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

func BuildTeachingReadmodelModule(root *Root, assessment *AssessmentModule) *TeachingReadmodelModule {
	log := root.Logger()
	db := root.DB()

	repo := readmodelinfra.NewRepository(db)
	service := readmodelqueries.NewQueryService(repo, assessment.Recommendations, log.Named("teaching_readmodel_query_service"))

	return &TeachingReadmodelModule{
		Handler: teachinghttp.NewHandler(service),
		Query:   service,
	}
}
