package constants

import "fmt"

// Redis Key 前缀常量
const (
	KeyPrefixUserProgress = "user:progress"
	KeyPrefixSubmitLimit  = "submit:limit"
)

// UserProgressKey 生成用户进度缓存 Key
func UserProgressKey(userID int64) string {
	return fmt.Sprintf("%s:%d", KeyPrefixUserProgress, userID)
}

// SubmitLimitKey 生成提交限流 Key
func SubmitLimitKey(userID, challengeID int64) string {
	return fmt.Sprintf("%s:%d:%d", KeyPrefixSubmitLimit, userID, challengeID)
}
