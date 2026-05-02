package composition

import (
	practicereadmodelruntime "ctf-platform/internal/module/practice_readmodel/runtime"
)

type PracticeReadmodelModule = practicereadmodelruntime.Module

func BuildPracticeReadmodelModule(root *Root) *PracticeReadmodelModule {
	return practicereadmodelruntime.Build(practicereadmodelruntime.Deps{
		Config: root.Config(),
		Logger: root.Logger(),
		DB:     root.DB(),
		Cache:  root.Cache(),
	})
}
