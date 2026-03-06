package dto

// DashboardStats 仪表盘统计数据
type DashboardStats struct {
	OnlineUsers      int64           `json:"online_users"`      // 在线用户数
	ActiveContainers int64           `json:"active_containers"` // 活跃容器数
	CPUUsage         float64         `json:"cpu_usage"`         // CPU 使用率 (%)
	MemoryUsage      float64         `json:"memory_usage"`      // 内存使用率 (%)
	ContainerStats   []ContainerStat `json:"container_stats"`   // 容器资源统计
	Alerts           []ResourceAlert `json:"alerts"`            // 资源告警
}

// ContainerStat 容器资源统计
type ContainerStat struct {
	ContainerID   string  `json:"container_id"`
	ContainerName string  `json:"container_name"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	MemoryUsage   int64   `json:"memory_usage"` // 字节
	MemoryLimit   int64   `json:"memory_limit"` // 字节
}

// ResourceAlert 资源告警
type ResourceAlert struct {
	ContainerID string  `json:"container_id"`
	Type        string  `json:"type"`      // cpu / memory
	Value       float64 `json:"value"`     // 使用率
	Threshold   float64 `json:"threshold"` // 阈值
	Message     string  `json:"message"`
}
