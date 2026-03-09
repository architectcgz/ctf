package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/pkg/response"
)

func Recovery(log *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		panicErr := fmt.Errorf("panic: %v", recovered)
		_ = c.Error(panicErr)

		log.Error("panic_recovered",
			zap.Any("panic", recovered),
			zap.String("request_id", c.GetString(RequestIDKey)),
			zap.String("path", c.Request.URL.Path),
			zap.ByteString("stack", debug.Stack()),
		)
		response.InternalError(c)
	})
}
