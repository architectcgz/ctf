package middleware

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/pkg/response"
)

func Recovery(log *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		requestID, _ := c.Get(RequestIDKey)
		log.Error("panic_recovered",
			zap.Any("panic", recovered),
			zap.String("request_id", valueOf(requestID)),
			zap.String("path", c.Request.URL.Path),
		)
		response.InternalError(c)
	})
}
