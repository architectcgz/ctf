package flagcrypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

const (
	// DynamicFlagHashLength 动态 Flag 哈希截取长度
	DynamicFlagHashLength = 32
)

// GenerateDynamicFlag 生成动态 Flag
// 算法: HMAC-SHA256(globalSecret, "userID:challengeID:nonce")
// prefix: Flag 前缀，如 "flag"、"ctf"，为空时默认 "flag"
func GenerateDynamicFlag(userID, challengeID int64, globalSecret, nonce, prefix string) string {
	message := fmt.Sprintf("%d:%d:%s", userID, challengeID, nonce)
	h := hmac.New(sha256.New, []byte(globalSecret))
	h.Write([]byte(message))
	hash := hex.EncodeToString(h.Sum(nil))

	if prefix == "" {
		prefix = "flag"
	}
	return fmt.Sprintf("%s{%s}", prefix, hash[:DynamicFlagHashLength])
}

// HashStaticFlag 对静态 Flag 进行哈希
func HashStaticFlag(flag, salt string) string {
	h := sha256.New()
	h.Write([]byte(flag + salt))
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateFlag 验证 Flag（防时序攻击）
func ValidateFlag(input, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(input), []byte(expected)) == 1
}
