package composition

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	runtimecmd "ctf-platform/internal/module/container_runtime/application/commands"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contestports "ctf-platform/internal/module/contest/ports"
	contestruntime "ctf-platform/internal/module/contest/runtime"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceinfra "ctf-platform/internal/module/instance/infrastructure"
	runtimestate "ctf-platform/internal/module/runtime/contracts"
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

	allocationRepo := runtimeinfra.NewAllocationRepository(root.DB())
	runtimeAWDRepo := runtimeinfra.NewAWDRepository(root.DB())
	instanceRepo := instanceinfra.NewRepository(root.DB())
	var cleanupService interface {
		CleanupRuntime(ctx context.Context, instance *instancecontracts.Instance) error
	} = newContestEndedRuntimeCleanupAdapter(runtimecmd.NewRuntimeCleanupService(
		runtime.runtime.CleanupRuntime,
		allocationRepo,
		logger.Named("contest_ended_runtime_cleanup_service"),
	))
	if runtime.nodeRouter != nil {
		cleanupService = newContestEndedRuntimeCleanupRouterAdapter(runtime.nodeRouter)
	}
	contestAWDRepo := contestinfra.NewAWDRepository(root.DB())
	return contestinfra.NewContestEndedRuntimeCleaner(
		contestAWDRepo,
		contestAWDRepo,
		cleanupService,
		newContestEndedRuntimeStateStore(root.DB(), instanceRepo, allocationRepo, runtimeAWDRepo),
	)
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

type contestEndedRuntimeStateStoreAdapter struct {
	db             *gorm.DB
	instanceRepo   *instanceinfra.Repository
	allocationRepo *runtimeinfra.AllocationRepository
	awdRepo        *runtimeinfra.AWDRepository
}

func newContestEndedRuntimeStateStore(db *gorm.DB, instanceRepo *instanceinfra.Repository, allocationRepo *runtimeinfra.AllocationRepository, awdRepo *runtimeinfra.AWDRepository) *contestEndedRuntimeStateStoreAdapter {
	if db == nil || instanceRepo == nil || allocationRepo == nil || awdRepo == nil {
		return nil
	}
	return &contestEndedRuntimeStateStoreAdapter{
		db:             db,
		instanceRepo:   instanceRepo,
		allocationRepo: allocationRepo,
		awdRepo:        awdRepo,
	}
}

func (a *contestEndedRuntimeStateStoreAdapter) ExpireInstanceRuntime(ctx context.Context, id int64) error {
	if a == nil {
		return nil
	}
	return withInstanceRuntimeLifecycleTx(ctx, a.db, a.instanceRepo, a.allocationRepo, func(instanceTx *instanceinfra.Repository, allocationTx *runtimeinfra.AllocationRepository) error {
		release, err := instanceTx.ExpireInstanceRuntime(ctx, id)
		if err != nil || release == nil {
			return err
		}
		return allocationTx.ReleaseRuntimeAllocationsForInstance(ctx, release.InstanceID, release.HostPort)
	})
}

func (a *contestEndedRuntimeStateStoreAdapter) FindAWDDefenseWorkspace(ctx context.Context, contestID, teamID, serviceID int64) (*runtimestate.AWDDefenseWorkspace, error) {
	if a == nil || a.awdRepo == nil {
		return nil, nil
	}
	return a.awdRepo.FindAWDDefenseWorkspace(ctx, contestID, teamID, serviceID)
}

func (a *contestEndedRuntimeStateStoreAdapter) UpsertAWDDefenseWorkspace(ctx context.Context, workspace *runtimestate.AWDDefenseWorkspace) error {
	if a == nil || a.awdRepo == nil {
		return nil
	}
	return a.awdRepo.UpsertAWDDefenseWorkspace(ctx, workspace)
}

func (a *contestEndedRuntimeStateStoreAdapter) FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error {
	if a == nil || a.awdRepo == nil {
		return nil
	}
	return a.awdRepo.FinishActiveAWDServiceOperationForInstance(ctx, instanceID, status, errorMessage, finishedAt)
}
