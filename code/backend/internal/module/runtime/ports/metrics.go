package ports

import instanceports "ctf-platform/internal/module/instance/ports"

type ManagedContainerStat struct {
	ContainerID   string
	ContainerName string
	CPUPercent    float64
	MemoryPercent float64
	MemoryUsage   int64
	MemoryLimit   int64
}

type ManagedContainer = instanceports.ManagedContainer

type ManagedContainerState = instanceports.ManagedContainerState
