package contracts

type ContainerConfig struct {
	Image          string
	Name           string
	Env            []string
	Command        []string
	WorkingDir     string
	Ports          map[string]string
	Mounts         []ContainerMount
	Labels         map[string]string
	Resources      *ResourceLimits
	Security       *SecurityConfig
	Network        string
	NetworkAliases []string
}

type ContainerMount struct {
	Source   string
	Target   string
	ReadOnly bool
}

type ResourceLimits struct {
	CPUQuota  float64
	Memory    int64
	PidsLimit int64
}

type SecurityConfig struct {
	ReadonlyRootfs bool
	CapDrop        []string
	CapAdd         []string
	SecurityOpt    []string
	User           string
}
