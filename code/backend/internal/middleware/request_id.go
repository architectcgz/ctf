package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "request_id"

var fallbackRequestIDCounter atomic.Uint64

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}

		c.Set(RequestIDKey, requestID)
		c.Request.Header.Set("X-Request-ID", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func newRequestID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("req_fallback_%d_%d", time.Now().UnixNano(), fallbackRequestIDCounter.Add(1))
	}
	return "req_" + hex.EncodeToString(buf)
}
