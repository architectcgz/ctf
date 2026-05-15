package composition

import (
	"go.uber.org/zap"

	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contestports "ctf-platform/internal/module/contest/ports"
	contestruntime "ctf-platform/internal/module/contest/runtime"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
)

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
		EndedRuntimeCleaner:   buildContestEndedRuntimeCleaner(root, runtime),
	})
	for _, job := range module.BackgroundJobs {
		root.RegisterBackgroundJob(NewLoopBackgroundJob(job.Name, job.Run))
	}
	return module
}

func buildContestEndedRuntimeCleaner(root *Root, runtime *ContainerRuntimeModule) contestports.ContestEndedRuntimeCleaner {
	if root == nil || runtime == nil || runtime.runtime == nil {
		return nil
	}

	logger := root.Logger()
	if logger == nil {
		logger = zap.NewNop()
	}

	runtimeRepo := runtimeinfra.NewRepository(root.DB())
	cleanupService := runtimecmd.NewRuntimeCleanupService(
		runtime.runtime.CleanupRuntime,
		runtimeRepo,
		logger.Named("contest_ended_runtime_cleanup_service"),
	)
	awdRepo := contestinfra.NewAWDRepository(root.DB())
	return contestinfra.NewContestEndedRuntimeCleaner(awdRepo, awdRepo, cleanupService, runtimeRepo)
}
