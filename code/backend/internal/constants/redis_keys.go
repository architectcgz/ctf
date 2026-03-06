package constants

import "fmt"

const (
	RedisNamespace = "ctf"
)

// WithNamespace 添加全局命名空间前缀
func WithNamespace(key string) string {
	return fmt.Sprintf("%s:%s", RedisNamespace, key)
}

// ChallengeSolvedCount 靶场完成人数缓存 Key
func ChallengeSolvedCount(challengeID int64) string {
	return WithNamespace(fmt.Sprintf("challenge:solved_count:%d", challengeID))
}
