package composition

import (
	teachingreadmodel "ctf-platform/internal/module/teaching_readmodel"
	teachinghttp "ctf-platform/internal/module/teaching_readmodel/api/http"
	readmodelapp "ctf-platform/internal/module/teaching_readmodel/application"
	readmodelinfra "ctf-platform/internal/module/teaching_readmodel/infrastructure"
)

type TeacherModule struct {
	Handler *teachinghttp.Handler
	Query   teachingreadmodel.TeachingQuery
}

func BuildTeacherModule(root *Root, assessment *AssessmentModule) *TeacherModule {
	log := root.Logger()
	db := root.DB()

	repo := readmodelinfra.NewRepository(db)
	service := readmodelapp.NewQueryService(repo, assessment.RecommendationService, log.Named("teacher_query_service"))

	return &TeacherModule{
		Handler: teachinghttp.NewHandler(service),
		Query:   teachingreadmodel.NewModule(repo),
	}
}
