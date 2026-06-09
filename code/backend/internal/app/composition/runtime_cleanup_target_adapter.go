package composition

import (
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

func runtimeCleanupTargetFromInstance(instance *instancecontracts.Instance) runtimecontracts.RuntimeCleanupTarget {
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
