package composition

import (
	teachingqueryruntime "ctf-platform/internal/module/teaching_query/runtime"
)

type TeachingQueryModule = teachingqueryruntime.Module

func BuildTeachingQueryModule(root *Root, assessment *AssessmentModule) *TeachingQueryModule {
	return teachingqueryruntime.Build(teachingqueryruntime.Deps{
		Config:          root.Config(),
		Logger:          root.Logger(),
		DB:              root.DB(),
		Recommendations: assessment.Recommendations,
	})
}
