package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	authModule "ctf-platform/internal/module/auth"
	"ctf-platform/pkg/errcode"
	jwtpkg "ctf-platform/pkg/jwt"
	"ctf-platform/pkg/response"
)

func Auth(tokenService authModule.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractBearerToken(c.GetHeader("Authorization"))
		if tokenString == "" {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}

		claims, err := tokenService.ParseToken(tokenString)
		if err != nil {
			response.FromError(c, mapAuthError(err))
			c.Abort()
			return
		}
		if claims.TokenType != jwtpkg.TokenTypeAccess {
			response.Error(c, errcode.ErrTokenInvalid)
			c.Abort()
			return
		}

		revoked, err := tokenService.IsRevoked(c.Request.Context(), claims.ID)
		if err != nil {
			response.FromError(c, errcode.ErrInternal.WithCause(err))
			c.Abort()
			return
		}
		if revoked {
			response.Error(c, errcode.ErrTokenRevoked)
			c.Abort()
			return
		}

		authctx.SetCurrentUser(c, authctx.CurrentUser{
			UserID:    claims.UserID,
			Username:  claims.Username,
			Role:      claims.Role,
			JTI:       claims.ID,
			ExpiresAt: claims.ExpiresAt.Time,
		})
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func MustCurrentUser(c *gin.Context) authctx.CurrentUser {
	return authctx.MustCurrentUser(c)
}

func extractBearerToken(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func mapAuthError(err error) error {
	switch {
	case errors.Is(err, jwtpkg.ErrExpiredToken):
		return errcode.ErrAccessTokenExpired
	case errors.Is(err, jwtpkg.ErrInvalidToken):
		return errcode.ErrTokenInvalid
	default:
		return errcode.ErrUnauthorized
	}
}
