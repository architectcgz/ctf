package model

// ContainerConfig 容器配置
type ContainerConfig struct {
	Image     string
	Env       []string
	Ports     map[string]string
	Resources *ResourceLimits
	Security  *SecurityConfig
	Network   string
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	CPUQuota  int64
	Memory    int64
	PidsLimit int64
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	ReadonlyRootfs bool
	CapDrop        []string
	CapAdd         []string
	SecurityOpt    []string
	User           string
}
