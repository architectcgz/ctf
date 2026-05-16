package composition

import (
	practiceruntime "ctf-platform/internal/module/practice/runtime"
)

type PracticeModule = practiceruntime.Module

func BuildPracticeModule(root *Root, challenge *ChallengeModule, instance *InstanceModule) *PracticeModule {
	module := practiceruntime.Build(practiceruntime.Deps{
		AppContext:     root.Context(),
		Config:         root.Config(),
		Logger:         root.Logger(),
		DB:             root.DB(),
		Cache:          root.Cache(),
		Events:         root.Events,
		InstanceRepo:   instance.PracticeInstanceRepository,
		RuntimeService: instance.PracticeRuntimeService,
		ChallengeRepo:  challenge.Catalog,
		ImageStore:     challenge.ImageStore,
	})
	if instance != nil && module != nil && module.AWDDesiredRuntimeReconciler != nil {
		instance.SetAWDDesiredRuntimeReconciler(module.AWDDesiredRuntimeReconciler)
	}
	for _, job := range module.BackgroundJobs {
		root.RegisterBackgroundJob(NewLoopBackgroundJob(job.Name, job.Run))
	}
	return module
}
