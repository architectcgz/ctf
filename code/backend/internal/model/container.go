package model

// ContainerConfig 容器配置
type ContainerConfig struct {
	Image          string
	Name           string
	Env            []string
	Ports          map[string]string
	Labels         map[string]string
	Resources      *ResourceLimits
	Security       *SecurityConfig
	Network        string
	NetworkAliases []string
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	CPUQuota  float64 // CPU 核心数，如 0.5 表示 0.5 核
	Memory    int64   // 内存限制（字节）
	PidsLimit int64   // 进程数限制
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	ReadonlyRootfs bool
	CapDrop        []string
	CapAdd         []string
	SecurityOpt    []string
	User           string
}
