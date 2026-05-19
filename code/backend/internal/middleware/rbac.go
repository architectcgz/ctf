package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/apperror"
	response "ctf-platform/internal/httpresponse"
)

var roleLevels = map[string]int{
	"student": 1,
	"teacher": 2,
	"admin":   3,
}

func RequireRole(roles ...string) gin.HandlerFunc {
	minLevel := 0
	for _, role := range roles {
		level := roleLevels[role]
		if level > minLevel {
			minLevel = level
		}
	}

	return func(c *gin.Context) {
		currentUser := MustCurrentUser(c)
		if roleLevels[currentUser.Role] < minLevel {
			response.Error(c, apperror.ErrForbidden)
			c.Abort()
			return
		}
		c.Next()
	}
}

func RoleGuardPing(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		response.SuccessWithStatus(c, http.StatusOK, gin.H{
			"scope":  scope,
			"status": "ok",
		})
	}
}
