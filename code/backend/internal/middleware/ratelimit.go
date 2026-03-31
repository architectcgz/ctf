package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ctf-platform/pkg/errcode"
	ratelimitpkg "ctf-platform/pkg/ratelimit"
	"ctf-platform/pkg/response"
)

func RateLimitByIP(checker *ratelimitpkg.Checker, scope string, limit int, window time.Duration) gin.HandlerFunc {
	return rateLimitMiddleware(checker, scope, limit, window, func(c *gin.Context) string {
		return c.ClientIP()
	})
}

func RateLimitByUser(checker *ratelimitpkg.Checker, scope string, limit int, window time.Duration) gin.HandlerFunc {
	return rateLimitMiddleware(checker, scope, limit, window, func(c *gin.Context) string {
		return strconv.FormatInt(MustCurrentUser(c).UserID, 10)
	})
}

func RateLimitByLoginPrincipalAndIP(checker *ratelimitpkg.Checker, scope string, limit int, window time.Duration) gin.HandlerFunc {
	return rateLimitMiddleware(checker, scope, limit, window, func(c *gin.Context) string {
		principal := strings.ToLower(strings.TrimSpace(loginRateLimitPrincipal(c)))
		if principal == "" {
			principal = "_"
		}
		return principal + ":" + c.ClientIP()
	})
}

func rateLimitMiddleware(checker *ratelimitpkg.Checker, scope string, limit int, window time.Duration, keyFunc func(c *gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := checker.CheckRate(c.Request.Context(), scope+":"+keyFunc(c), limit, window)
		if err != nil {
			response.FromError(c, errcode.ErrInternal.WithCause(err))
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(result.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))
		if result.RetryAfter > 0 {
			c.Header("Retry-After", strconv.FormatInt(int64(result.RetryAfter.Seconds()), 10))
		}

		if !result.Allowed {
			response.Error(c, errcode.ErrRateLimitExceeded)
			c.Abort()
			return
		}
		c.Next()
	}
}

func loginRateLimitPrincipal(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}

	if strings.Contains(strings.ToLower(c.GetHeader("Content-Type")), "application/json") {
		return loginPrincipalFromJSONBody(c)
	}
	return strings.TrimSpace(c.PostForm("username"))
}

func loginPrincipalFromJSONBody(c *gin.Context) string {
	if c.Request == nil || c.Request.Body == nil {
		return ""
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Request.Body = io.NopCloser(bytes.NewReader(nil))
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	if len(body) == 0 {
		return ""
	}

	var payload struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	return strings.TrimSpace(payload.Username)
}
