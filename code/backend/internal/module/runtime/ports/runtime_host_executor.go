package ports

// RuntimeHostExecutor 收口宿主机执行面的本地/远端 adapter 能力边界。
type RuntimeHostExecutor interface {
	ContainerProvisioningRuntime
	ContainerCleanupRuntime
	ContainerFileRuntime
	ContainerImageRuntime
	ManagedContainerInventory
	ManagedContainerStatsReader
	ContainerInteractiveExecutor
}
