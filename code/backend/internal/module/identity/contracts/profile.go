package contracts

import "context"

type ProfileUser struct {
	ID        int64
	Username  string
	Role      string
	Avatar    *string
	Name      *string
	ClassName *string
}

type ProfileQueryService interface {
	GetProfile(ctx context.Context, userID int64) (*ProfileUser, error)
}

type ChangePasswordInput struct {
	OldPassword string
	NewPassword string
}

type ProfileCommandService interface {
	ChangePassword(ctx context.Context, userID int64, req ChangePasswordInput) error
}
