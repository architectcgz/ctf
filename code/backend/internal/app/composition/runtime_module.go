package composition

import (
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestports "ctf-platform/internal/module/contest/ports"
	opsports "ctf-platform/internal/module/ops/ports"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	runtimemodule "ctf-platform/internal/module/runtime/runtime"
	"go.uber.org/zap"
)

type ContainerRuntimeModule struct {
	ChallengeImageRuntime   challengeports.ImageRuntime
	ChallengeRuntimeProbe   challengeports.ChallengeRuntimeProbe
	OpsRuntimeQuery         opsports.RuntimeQuery
	OpsRuntimeStatsProvider opsports.RuntimeStatsProvider
	ContestContainerFiles   contestports.AWDContainerFileWriter

	runtime *runtimemodule.Module
}

type RuntimeModule = ContainerRuntimeModule

type runtimeCapabilityEngine interface {
	runtimeports.ContainerProvisioningRuntime
	runtimeports.ContainerCleanupRuntime
	runtimeports.ContainerFileRuntime
	runtimeports.ContainerImageRuntime
	runtimeports.ManagedContainerInventory
	runtimeports.ManagedContainerStatsReader
	runtimeports.ContainerInteractiveExecutor
}

func BuildContainerRuntimeModule(root *Root) *ContainerRuntimeModule {
	cfg := root.Config()
	log := root.Logger()
	engine := buildRuntimeEngine(root)
	module := runtimemodule.Build(runtimemodule.Deps{
		Config:                    cfg,
		Logger:                    log,
		DB:                        root.DB(),
		Cache:                     root.Cache(),
		ProvisioningRuntime:       engine,
		CleanupRuntime:            engine,
		FileRuntime:               engine,
		ImageRuntime:              engine,
		ManagedContainerInventory: engine,
		ManagedContainerStats:     engine,
		InteractiveExecutor:       engine,
	})

	for _, job := range module.BackgroundJobs {
		root.RegisterBackgroundJob(NewBackgroundJob(job.Name, job.Start, job.Stop))
	}

	return &ContainerRuntimeModule{
		ChallengeImageRuntime:   module.ChallengeImageRuntime,
		ChallengeRuntimeProbe:   module.ChallengeRuntimeProbe,
		OpsRuntimeQuery:         module.OpsRuntimeQuery,
		OpsRuntimeStatsProvider: module.OpsRuntimeStatsProvider,
		ContestContainerFiles:   module.ContestContainerFiles,
		runtime:                 module,
	}
}

func BuildRuntimeModule(root *Root) *RuntimeModule {
	return BuildContainerRuntimeModule(root)
}

func buildRuntimeEngine(root *Root) runtimeCapabilityEngine {
	if root == nil {
		return nil
	}

	cfg := root.Config()
	log := root.Logger()
	if cfg == nil {
		return nil
	}
	if cfg.App.Env == "test" {
		if log != nil {
			log.Info("runtime_engine_enabled_with_test_adapter_for_router")
		}
		return newTestRuntimeEngine(log.Named("runtime_test_engine"))
	}

	engine, err := runtimeinfra.NewEngine(&cfg.Container)
	if err != nil {
		if log != nil {
			log.Warn("runtime_engine_init_failed_for_router", zap.Error(err))
		}
		return nil
	}
	return engine
}
