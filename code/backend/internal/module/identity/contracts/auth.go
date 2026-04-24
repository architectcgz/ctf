package contracts

import (
	"context"
	"time"

	authcontracts "ctf-platform/internal/module/auth/contracts"
	"ctf-platform/internal/model"
)

type Authenticator interface {
	authcontracts.TokenService
}

type UserListFilter struct {
	Keyword   string
	StudentNo string
	TeacherNo string
	Role      string
	Status    string
	ClassName string
	Offset    int
	Limit     int
}

type UserRepository interface {
	List(ctx context.Context, filter UserListFilter) ([]*model.User, int64, error)
	FindByID(ctx context.Context, userID int64) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, userID int64) error
	UpdatePassword(ctx context.Context, userID int64, newHash string) error
	UpdateLoginState(ctx context.Context, userID int64, failedAttempts int, lastFailedAt, lockedUntil *time.Time, status string) error
	UpdateProfile(ctx context.Context, user *model.User) error
}
