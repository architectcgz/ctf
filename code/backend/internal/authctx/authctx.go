package authctx

import (
	"time"

	"github.com/gin-gonic/gin"
)

const CurrentUserKey = "current_user"

type CurrentUser struct {
	UserID    int64
	Username  string
	Role      string
	SessionID string
	ExpiresAt time.Time
}

func SetCurrentUser(c *gin.Context, user CurrentUser) {
	c.Set(CurrentUserKey, user)
}

func MustCurrentUser(c *gin.Context) CurrentUser {
	value, _ := c.Get(CurrentUserKey)
	user, _ := value.(CurrentUser)
	return user
}
