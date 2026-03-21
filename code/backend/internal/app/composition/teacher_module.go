package composition

import (
	teacherModule "ctf-platform/internal/module/teacher"
	teachingreadmodel "ctf-platform/internal/module/teaching_readmodel"
)

type TeacherModule struct {
	Handler *teacherModule.Handler
	Query   teachingreadmodel.TeachingQuery
}

func BuildTeacherModule(root *Root, assessment *AssessmentModule) *TeacherModule {
	log := root.Logger()
	db := root.DB()

	repo := teacherModule.NewRepository(db)
	service := teacherModule.NewService(repo, assessment.RecommendationService, log.Named("teacher_service"))

	return &TeacherModule{
		Handler: teacherModule.NewHandler(service),
		Query:   teachingreadmodel.NewModule(repo),
	}
}
