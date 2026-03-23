package application

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	identitymodule "ctf-platform/internal/module/identity"
	"ctf-platform/pkg/errcode"
)

type ProfileService struct {
	users identitymodule.UserRepository
	log   *zap.Logger
}

var _ identitymodule.ProfileService = (*ProfileService)(nil)

func NewProfileService(users identitymodule.UserRepository, log *zap.Logger) *ProfileService {
	if log == nil {
		log = zap.NewNop()
	}
	return &ProfileService{
		users: users,
		log:   log,
	}
}

func (s *ProfileService) GetProfile(ctx context.Context, userID int64) (*dto.AuthUser, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, identitymodule.ErrUserNotFound) {
			return nil, errcode.ErrUnauthorized
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	profile := buildAuthUser(user)
	return &profile, nil
}

func (s *ProfileService) ChangePassword(ctx context.Context, userID int64, req *dto.ChangePasswordReq) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, identitymodule.ErrUserNotFound) {
			return errcode.ErrUnauthorized
		}
		return errcode.ErrInternal.WithCause(err)
	}

	if !user.CheckPassword(req.OldPassword) {
		s.log.Warn("identity_change_password_failed_old_password_invalid", zap.Int64("user_id", userID))
		return errcode.ErrOldPasswordInvalid
	}
	if req.OldPassword == req.NewPassword {
		s.log.Warn("identity_change_password_failed_password_unchanged", zap.Int64("user_id", userID))
		return errcode.ErrPasswordUnchanged
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		s.log.Error("identity_change_password_hash_failed", zap.Int64("user_id", userID), zap.Error(err))
		return errcode.ErrInternal.WithCause(err)
	}
	if err := s.users.UpdatePassword(ctx, userID, user.PasswordHash); err != nil {
		if errors.Is(err, identitymodule.ErrUserNotFound) {
			return errcode.ErrUnauthorized
		}
		s.log.Error("identity_change_password_update_failed", zap.Int64("user_id", userID), zap.Error(err))
		return errcode.ErrInternal.WithCause(err)
	}

	s.log.Info("identity_change_password_succeeded", zap.Int64("user_id", userID))
	return nil
}

func buildAuthUser(user *model.User) dto.AuthUser {
	var name *string
	if strings.TrimSpace(user.Name) != "" {
		name = &user.Name
	}
	var className *string
	if strings.TrimSpace(user.ClassName) != "" {
		className = &user.ClassName
	}

	return dto.AuthUser{
		ID:        user.ID,
		Username:  user.Username,
		Role:      user.Role,
		Name:      name,
		ClassName: className,
	}
}
