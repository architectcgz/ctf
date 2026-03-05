package config

// ContainerConfig 容器配置
type ContainerConfig struct {
	PortRangeStart int `mapstructure:"port_range_start"`
	PortRangeEnd   int `mapstructure:"port_range_end"`
}

// Config 全局配置
type Config struct {
	Container ContainerConfig `mapstructure:"container"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Container: ContainerConfig{
			PortRangeStart: 30000,
			PortRangeEnd:   40000,
		},
	}
}
