package runtime

import (
	containerruntime "ctf-platform/internal/module/container_runtime/runtime"
	contestports "ctf-platform/internal/module/contest/ports"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type BackgroundJob = containerruntime.BackgroundJob
type Deps = containerruntime.Deps

// Module preserves the old runtime module exported field surface while
// delegating construction to container_runtime/runtime.
type Module struct {
	BackgroundJobs []BackgroundJob

	ImageRuntime          *runtimeapp.ImageRuntimeService
	RuntimeStatsProvider  runtimeports.ManagedContainerStatsReader
	ContestContainerFiles contestports.AWDContainerFileWriter
	ProvisioningService   *runtimecmd.ProvisioningService
	CleanupService        *runtimecmd.RuntimeCleanupService

	ProvisioningRuntime       runtimeports.ContainerProvisioningRuntime
	CleanupRuntime            runtimeports.ContainerCleanupRuntime
	FileRuntime               runtimeports.ContainerFileRuntime
	ManagedContainerInventory runtimeports.ManagedContainerInventory
	InteractiveExecutor       runtimeports.ContainerInteractiveExecutor
}

func Build(deps Deps) *Module {
	module := containerruntime.Build(deps)
	if module == nil {
		return nil
	}
	return &Module{
		BackgroundJobs:            module.BackgroundJobs,
		ImageRuntime:              module.ImageRuntime,
		RuntimeStatsProvider:      module.RuntimeStatsProvider,
		ContestContainerFiles:     module.ContainerFiles,
		ProvisioningService:       module.ProvisioningService,
		CleanupService:            module.CleanupService,
		ProvisioningRuntime:       module.ProvisioningRuntime,
		CleanupRuntime:            module.CleanupRuntime,
		FileRuntime:               module.FileRuntime,
		ManagedContainerInventory: module.ManagedContainerInventory,
		InteractiveExecutor:       module.InteractiveExecutor,
	}
}
