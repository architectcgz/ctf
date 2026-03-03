package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AccessLog(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		requestID, _ := c.Get(RequestIDKey)
		log.Info("http_request",
			zap.String("request_id", valueOf(requestID)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

func valueOf(raw any) string {
	if v, ok := raw.(string); ok {
		return v
	}
	return ""
}
