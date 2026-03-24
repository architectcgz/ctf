package application

import "time"

// Config 描述 challenge 应用服务运行所需的配置项。
type Config struct {
	SolvedCountCacheTTL time.Duration
}
