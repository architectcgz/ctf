package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/authctx"
	response "ctf-platform/internal/httpresponse"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

func Auth(tokenService authcontracts.TokenService, cookieName string, users ...identitycontracts.UserLookupRepository) gin.HandlerFunc {
	var userRepo identitycontracts.UserLookupRepository
	if len(users) > 0 {
		userRepo = users[0]
	}

	return func(c *gin.Context) {
		sessionID, err := c.Cookie(cookieName)
		if err != nil || sessionID == "" {
			response.Error(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		session, err := tokenService.GetSession(c.Request.Context(), sessionID)
		if err != nil {
			response.FromError(c, err)
			c.Abort()
			return
		}

		username := session.Username
		role := session.Role
		if userRepo != nil {
			user, err := userRepo.FindByID(c.Request.Context(), session.UserID)
			if err != nil {
				if errors.Is(err, identitycontracts.ErrUserNotFound) {
					response.Error(c, authcontracts.ErrAccessTokenExpired)
				} else {
					response.FromError(c, apperror.ErrInternal.WithCause(err))
				}
				c.Abort()
				return
			}
			username = user.Username
			role = user.Role
		}

		authctx.SetCurrentUser(c, authctx.CurrentUser{
			UserID:    session.UserID,
			Username:  username,
			Role:      role,
			SessionID: session.ID,
			ExpiresAt: session.ExpiresAt,
		})
		c.Next()
	}
}

func MustCurrentUser(c *gin.Context) authctx.CurrentUser {
	return authctx.MustCurrentUser(c)
}
