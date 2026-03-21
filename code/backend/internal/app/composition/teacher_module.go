package composition

import teacherModule "ctf-platform/internal/module/teacher"

type TeacherModule struct {
	Handler *teacherModule.Handler
}

func BuildTeacherModule(root *Root, assessment *AssessmentModule) *TeacherModule {
	log := root.Logger()
	db := root.DB()

	repo := teacherModule.NewRepository(db)
	service := teacherModule.NewService(repo, assessment.RecommendationService, log.Named("teacher_service"))

	return &TeacherModule{
		Handler: teacherModule.NewHandler(service),
	}
}
