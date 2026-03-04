package middleware

import (
	"strconv"
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
