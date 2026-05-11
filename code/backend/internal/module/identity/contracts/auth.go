package contracts

import (
	"context"
	"time"

	"ctf-platform/internal/model"
)

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

type UserListRepository interface {
	List(ctx context.Context, filter UserListFilter) ([]*model.User, int64, error)
}

type UserLookupRepository interface {
	FindByID(ctx context.Context, userID int64) (*model.User, error)
	FindByUsername(ctx context.Context, username string) (*model.User, error)
}

type UserWriteRepository interface {
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, userID int64) error
}

type UserPasswordRepository interface {
	UpdatePassword(ctx context.Context, userID int64, newHash string) error
}

type UserLoginStateRepository interface {
	UpdateLoginState(ctx context.Context, userID int64, failedAttempts int, lastFailedAt, lockedUntil *time.Time, status string) error
}

type UserProfileRepository interface {
	UpdateProfile(ctx context.Context, user *model.User) error
}

type UserRepository interface {
	UserListRepository
	UserLookupRepository
	UserWriteRepository
	UserPasswordRepository
	UserLoginStateRepository
	UserProfileRepository
}
