package composition

import (
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

func runtimeCleanupTargetFromInstance(instance *instancecontracts.Instance) runtimecontracts.RuntimeCleanupTarget {
	if instance == nil {
		return runtimecontracts.RuntimeCleanupTarget{}
	}
	return runtimecontracts.RuntimeCleanupTarget{
		InstanceID:     instance.ID,
		RuntimeNodeID:  instance.RuntimeNodeID,
		ContainerID:    instance.ContainerID,
		NetworkID:      instance.NetworkID,
		HostPort:       instance.HostPort,
		RuntimeDetails: instance.RuntimeDetails,
	}
}
