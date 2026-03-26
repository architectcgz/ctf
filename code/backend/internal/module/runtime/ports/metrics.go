package ports

import "time"

type ManagedContainerStat struct {
	ContainerID   string
	ContainerName string
	CPUPercent    float64
	MemoryPercent float64
	MemoryUsage   int64
	MemoryLimit   int64
}

type ManagedContainer struct {
	ID        string
	Name      string
	CreatedAt time.Time
}
