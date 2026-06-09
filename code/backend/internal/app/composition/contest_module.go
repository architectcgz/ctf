package composition

import (
	"context"

	"go.uber.org/zap"

	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contestports "ctf-platform/internal/module/contest/ports"
	contestruntime "ctf-platform/internal/module/contest/runtime"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
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
		CheckerRunner:         runtime.ContestCheckerRunner,
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
	var cleanupService interface {
		CleanupRuntime(ctx context.Context, instance *instancecontracts.Instance) error
	} = newContestEndedRuntimeCleanupAdapter(runtimecmd.NewRuntimeCleanupService(
		runtime.runtime.CleanupRuntime,
		runtimeRepo,
		logger.Named("contest_ended_runtime_cleanup_service"),
	))
	if runtime.nodeRouter != nil {
		cleanupService = newContestEndedRuntimeCleanupRouterAdapter(runtime.nodeRouter)
	}
	awdRepo := contestinfra.NewAWDRepository(root.DB())
	return contestinfra.NewContestEndedRuntimeCleaner(awdRepo, awdRepo, cleanupService, runtimeRepo)
}

type contestEndedRuntimeCleanupAdapter struct {
	cleaner *runtimecmd.RuntimeCleanupService
}

func newContestEndedRuntimeCleanupAdapter(cleaner *runtimecmd.RuntimeCleanupService) *contestEndedRuntimeCleanupAdapter {
	if cleaner == nil {
		return nil
	}
	return &contestEndedRuntimeCleanupAdapter{cleaner: cleaner}
}

func (a *contestEndedRuntimeCleanupAdapter) CleanupRuntime(ctx context.Context, instance *instancecontracts.Instance) error {
	if a == nil || a.cleaner == nil {
		return nil
	}
	return a.cleaner.CleanupRuntime(ctx, runtimeCleanupTargetFromInstance(instance))
}

type contestEndedRuntimeCleanupRouterAdapter struct {
	router *runtimeNodeExecutionRouter
}

func newContestEndedRuntimeCleanupRouterAdapter(router *runtimeNodeExecutionRouter) *contestEndedRuntimeCleanupRouterAdapter {
	if router == nil {
		return nil
	}
	return &contestEndedRuntimeCleanupRouterAdapter{router: router}
}

func (a *contestEndedRuntimeCleanupRouterAdapter) CleanupRuntime(ctx context.Context, instance *instancecontracts.Instance) error {
	if a == nil || a.router == nil {
		return nil
	}
	return a.router.CleanupRuntime(ctx, runtimeCleanupTargetFromInstance(instance))
}
