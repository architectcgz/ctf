package contracts

import (
	"context"
	"time"

	identityentity "ctf-platform/internal/module/identity/entity"
)

type User = identityentity.User
type UserRole = identityentity.UserRole
type Role = identityentity.Role

const (
	UserStatusActive   = identityentity.UserStatusActive
	UserStatusInactive = identityentity.UserStatusInactive
	UserStatusLocked   = identityentity.UserStatusLocked
	UserStatusBanned   = identityentity.UserStatusBanned

	RoleStudent = identityentity.RoleStudent
	RoleTeacher = identityentity.RoleTeacher
	RoleAdmin   = identityentity.RoleAdmin
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
	List(ctx context.Context, filter UserListFilter) ([]*User, int64, error)
}

type UserLookupRepository interface {
	FindByID(ctx context.Context, userID int64) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
}

type UserWriteRepository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, userID int64) error
}

type UserPasswordRepository interface {
	UpdatePassword(ctx context.Context, userID int64, newHash string) error
}

type UserLoginStateRepository interface {
	UpdateLoginState(ctx context.Context, userID int64, failedAttempts int, lastFailedAt, lockedUntil *time.Time, status string) error
}

type UserProfileRepository interface {
	UpdateProfile(ctx context.Context, user *User) error
}

type UserRepository interface {
	UserListRepository
	UserLookupRepository
	UserWriteRepository
	UserPasswordRepository
	UserLoginStateRepository
	UserProfileRepository
}
