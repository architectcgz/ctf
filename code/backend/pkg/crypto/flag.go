package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// GenerateDynamicFlag 生成动态 Flag
// 算法: HMAC-SHA256(globalSecret, "userID:challengeID:nonce")
func GenerateDynamicFlag(userID, challengeID int64, globalSecret, nonce string) string {
	message := fmt.Sprintf("%d:%d:%s", userID, challengeID, nonce)
	h := hmac.New(sha256.New, []byte(globalSecret))
	h.Write([]byte(message))
	hash := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("flag{%s}", hash[:32])
}

// HashStaticFlag 对静态 Flag 进行哈希
func HashStaticFlag(flag, salt string) string {
	h := sha256.New()
	h.Write([]byte(flag + salt))
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateFlag 验证 Flag
func ValidateFlag(input, expected string) bool {
	return input == expected
}

// GenerateSalt 生成随机盐值
func GenerateSalt() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateNonce 生成实例随机值
func GenerateNonce() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
