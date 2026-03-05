package model

import "time"

// ContainerConfig 容器配置
type ContainerConfig struct {
	Image     string
	Env       []string
	Ports     map[string]string // 容器端口:宿主机端口
	Resources *ResourceLimits
	Security  *SecurityConfig
	Network   string
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	Name      string // 网络名称，格式：ctf-{challenge_id}-{instance_id}
	NetworkID string // Docker 网络 ID
	Driver    string // 网络驱动，默认 bridge
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	CPUQuota  int64  // CPU 配额（微秒），如 50000 表示 0.5 核
	Memory    int64  // 内存限制（字节）
	PidsLimit int64  // 进程数限制
	DiskQuota string // 磁盘配额，如 "1G"
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	Privileged             bool     // 是否特权模式（强制 false）
	ReadonlyRootfs         bool     // 只读根文件系统
	NoNewPrivileges        bool     // 禁止新特权获取
	CapDrop                []string // 移除的 Capabilities
	CapAdd                 []string // 添加的 Capabilities
	SecurityOpt            []string // 安全选项（如 Seccomp）
	User                   string   // 运行用户（非 root）
}

// ContainerStatus 容器状态
type ContainerStatus struct {
	ID      string
	State   string // running, exited, paused 等
	StartAt time.Time
}

// ImageInfo 镜像信息
type ImageInfo struct {
	ID      string
	RepoTag string
	Size    int64
	Created time.Time
}
