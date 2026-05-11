package composition

import contestruntime "ctf-platform/internal/module/contest/runtime"

type ContestModule = contestruntime.Module

func BuildContestModule(root *Root, challenge *ChallengeModule, runtime *ContainerRuntimeModule) *ContestModule {
	module := contestruntime.Build(contestruntime.Deps{
		Config:                root.Config(),
		Logger:                root.Logger(),
		DB:                    root.DB(),
		Cache:                 root.Cache(),
		Events:                root.Events,
		ChallengeCatalog:      challenge.Catalog,
		AWDChallengeQueryRepo: challenge.AWDChallengeQuery,
		ImageRepo:             challenge.ImageStore,
		FlagValidator:         challenge.FlagValidator,
		ContainerFiles:        runtime.ContestContainerFiles,
		RuntimeProbe:          runtime.ChallengeRuntimeProbe,
	})
	for _, job := range module.BackgroundJobs {
		root.RegisterBackgroundJob(NewLoopBackgroundJob(job.Name, job.Run))
	}
	return module
}
