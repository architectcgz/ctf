package cache

import "fmt"

// Redis Key 前缀常量
const (
	KeyPrefixUserScore = "ctf:score:user"
	KeyPrefixRanking   = "ctf:ranking"
	KeyPrefixScoreLock = "ctf:lock:score"
)

// UserScoreKey 生成用户得分缓存键
func UserScoreKey(userID int64) string {
	return fmt.Sprintf("%s:%d", KeyPrefixUserScore, userID)
}

// RankingKey 生成排行榜缓存键
func RankingKey() string {
	return KeyPrefixRanking
}

// ScoreLockKey 生成计分分布式锁键
func ScoreLockKey(userID int64) string {
	return fmt.Sprintf("%s:%d", KeyPrefixScoreLock, userID)
}
