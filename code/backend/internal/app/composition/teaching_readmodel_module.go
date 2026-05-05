package composition

import (
	readmodelruntime "ctf-platform/internal/module/teaching_readmodel/runtime"
)

type TeachingReadmodelModule = readmodelruntime.Module

func BuildTeachingReadmodelModule(root *Root, assessment *AssessmentModule) *TeachingReadmodelModule {
	return readmodelruntime.Build(readmodelruntime.Deps{
		Config:          root.Config(),
		Logger:          root.Logger(),
		DB:              root.DB(),
		Recommendations: assessment.Recommendations,
	})
}
