package composition

import (
	practiceruntime "ctf-platform/internal/module/practice/runtime"
)

type PracticeModule = practiceruntime.Module

func BuildPracticeModule(root *Root, challenge *ChallengeModule, runtime *RuntimeModule, assessment *AssessmentModule) *PracticeModule {
	module := practiceruntime.Build(practiceruntime.Deps{
		AppContext:     root.Context(),
		Config:         root.Config(),
		Logger:         root.Logger(),
		DB:             root.DB(),
		Cache:          root.Cache(),
		Events:         root.Events,
		InstanceRepo:   runtime.PracticeInstanceRepository,
		RuntimeService: runtime.PracticeRuntimeService,
		ChallengeRepo:  challenge.Catalog,
		ImageStore:     challenge.ImageStore,
		Assessment:     assessment.ProfileService,
	})
	for _, job := range module.BackgroundJobs {
		root.RegisterBackgroundJob(NewLoopBackgroundJob(job.Name, job.Run))
	}
	return module
}
