package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
)

// ParseChallengeID 解析路径参数中的 challenge ID
func ParseChallengeID() gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			response.FromError(c, errcode.ErrInvalidParams)
			c.Abort()
			return
		}
		c.Set("challenge_id", id)
		c.Next()
	}
}
