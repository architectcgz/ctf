package cache

import "fmt"

const (
	KeyPrefixChallenge = "challenge"
)

// ChallengeSolvedCountKey 靶场完成人数缓存键
func ChallengeSolvedCountKey(challengeID int64) string {
	return fmt.Sprintf("%s:solved_count:%d", KeyPrefixChallenge, challengeID)
}
