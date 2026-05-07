package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func BuildAWDCheckerToken(contestID, teamID, serviceID, challengeID int64, secret string) string {
	return buildAWDCheckerScopedToken(secret, "runtime", contestID, teamID, serviceID, challengeID)
}

func BuildAWDCheckerPreviewToken(contestID, serviceID, challengeID int64, secret string) string {
	return buildAWDCheckerScopedToken(secret, "preview", contestID, 0, serviceID, challengeID)
}

func buildAWDCheckerScopedToken(secret, scope string, contestID, teamID, serviceID, challengeID int64) string {
	secret = strings.TrimSpace(secret)
	scope = strings.TrimSpace(scope)
	if secret == "" || scope == "" {
		return ""
	}
	message := fmt.Sprintf("awd:checker:%s:%d:%d:%d:%d", scope, contestID, teamID, serviceID, challengeID)
	hash := hmac.New(sha256.New, []byte(secret))
	_, _ = hash.Write([]byte(message))
	return hex.EncodeToString(hash.Sum(nil))
}
