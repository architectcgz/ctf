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

		fields := []zap.Field{
			zap.String("request_id", c.GetString(RequestIDKey)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		}
		if err := lastRequestError(c); err != nil {
			fields = append(fields, zap.Error(err))
		}

		status := c.Writer.Status()
		switch {
		case status >= 500:
			log.Error("http_request", fields...)
		case status >= 400 && len(c.Errors) > 0:
			log.Warn("http_request", fields...)
		default:
			log.Info("http_request", fields...)
		}
	}
}

func lastRequestError(c *gin.Context) error {
	if len(c.Errors) == 0 {
		return nil
	}

	return c.Errors.Last().Err
}
