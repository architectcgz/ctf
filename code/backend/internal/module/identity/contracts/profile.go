package contracts

import (
	"context"

	"ctf-platform/internal/dto"
)

type ProfileQueryService interface {
	GetProfile(ctx context.Context, userID int64) (*dto.AuthUser, error)
}

type ProfileCommandService interface {
	ChangePassword(ctx context.Context, userID int64, req *dto.ChangePasswordReq) error
}
