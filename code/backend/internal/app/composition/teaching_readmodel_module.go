package composition

import (
	teachingreadmodel "ctf-platform/internal/module/teaching_readmodel"
	teachinghttp "ctf-platform/internal/module/teaching_readmodel/api/http"
	readmodelapp "ctf-platform/internal/module/teaching_readmodel/application"
	readmodelinfra "ctf-platform/internal/module/teaching_readmodel/infrastructure"
)

type TeachingReadmodelModule struct {
	Handler *teachinghttp.Handler
	Query   teachingreadmodel.TeachingQuery
}

func BuildTeachingReadmodelModule(root *Root, assessment *AssessmentModule) *TeachingReadmodelModule {
	log := root.Logger()
	db := root.DB()

	repo := readmodelinfra.NewRepository(db)
	service := readmodelapp.NewQueryService(repo, assessment.RecommendationService, log.Named("teaching_readmodel_query_service"))

	return &TeachingReadmodelModule{
		Handler: teachinghttp.NewHandler(service),
		Query:   service,
	}
}
