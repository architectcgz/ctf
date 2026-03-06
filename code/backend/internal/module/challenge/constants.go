package challenge

import "time"

const (
	// 分页默认值
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100

	// 缓存 TTL
	SolvedCountCacheTTL = 5 * time.Minute
)
