package runtime_test

import (
	"context"

	runtimecmd "ctf-platform/internal/module/container_runtime/application/commands"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
)

func runtimeCleanupTarget(instance *instanceentity.Instance) runtimecontracts.RuntimeCleanupTarget {
	if instance == nil {
		return runtimecontracts.RuntimeCleanupTarget{}
	}
	return runtimecontracts.RuntimeCleanupTarget{
		InstanceID:     instance.ID,
		NodeID:         instance.NodeID,
		ContainerID:    instance.ContainerID,
		NetworkID:      instance.NetworkID,
		HostPort:       instance.HostPort,
		RuntimeDetails: instance.RuntimeDetails,
	}
}

type runtimeTestCleanerAdapter struct {
	cleaner *runtimecmd.RuntimeCleanupService
}

func newRuntimeTestCleanerAdapter(cleaner *runtimecmd.RuntimeCleanupService) *runtimeTestCleanerAdapter {
	if cleaner == nil {
		return nil
	}
	return &runtimeTestCleanerAdapter{cleaner: cleaner}
}

func (a *runtimeTestCleanerAdapter) CleanupRuntime(ctx context.Context, instance *instanceentity.Instance) error {
	if a == nil || a.cleaner == nil {
		return nil
	}
	return a.cleaner.CleanupRuntime(ctx, runtimeCleanupTarget(instance))
}

func (a *runtimeTestCleanerAdapter) RemoveContainer(ctx context.Context, containerID string) error {
	if a == nil || a.cleaner == nil {
		return nil
	}
	return a.cleaner.RemoveContainer(ctx, containerID)
}
