package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const (
	// DynamicFlagHashLength 动态 Flag 哈希截取长度
	DynamicFlagHashLength = 32
	// RandomStringLength 随机字符串长度（盐值/Nonce）
	RandomStringLength = 32
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

// generateRandomString 生成随机字符串（内部函数）
func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateSalt 生成随机盐值
func GenerateSalt() (string, error) {
	return generateRandomString(RandomStringLength)
}

// GenerateNonce 生成实例随机值
// 注意：此函数应在 B18（实例启动）任务中调用，生成的 nonce 存储到 instances.nonce 字段
func GenerateNonce() (string, error) {
	return generateRandomString(RandomStringLength)
}
