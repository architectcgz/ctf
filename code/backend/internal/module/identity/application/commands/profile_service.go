package commands

import (
	"context"
	"errors"

	"go.uber.org/zap"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/pkg/errcode"
)

type ProfileService struct {
	users profileCommandRepository
	log   *zap.Logger
}

type profileCommandRepository interface {
	identitycontracts.UserLookupRepository
	identitycontracts.UserPasswordRepository
}

var _ identitycontracts.ProfileCommandService = (*ProfileService)(nil)

func NewProfileService(users profileCommandRepository, log *zap.Logger) *ProfileService {
	if log == nil {
		log = zap.NewNop()
	}
	return &ProfileService{
		users: users,
		log:   log,
	}
}

func (s *ProfileService) ChangePassword(ctx context.Context, userID int64, req identitycontracts.ChangePasswordInput) error {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, identitycontracts.ErrUserNotFound) {
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
		if errors.Is(err, identitycontracts.ErrUserNotFound) {
			return errcode.ErrUnauthorized
		}
		s.log.Error("identity_change_password_update_failed", zap.Int64("user_id", userID), zap.Error(err))
		return errcode.ErrInternal.WithCause(err)
	}

	s.log.Info("identity_change_password_succeeded", zap.Int64("user_id", userID))
	return nil
}
