package auth

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

type Service interface {
	Register(ctx context.Context, req *dto.RegisterReq) (*dto.LoginResp, *TokenPair, error)
	Login(ctx context.Context, req *dto.LoginReq) (*dto.LoginResp, *TokenPair, error)
	GetProfile(ctx context.Context, userID int64) (*dto.AuthUser, error)
	ChangePassword(ctx context.Context, userID int64, req *dto.ChangePasswordReq) error
	ValidatePassword(user *model.User, password string) bool
}

type service struct {
	repo         Repository
	tokenService TokenService
	log          *zap.Logger
}

func NewService(repo Repository, tokenService TokenService, log *zap.Logger) Service {
	if log == nil {
		log = zap.NewNop()
	}

	return &service{
		repo:         repo,
		tokenService: tokenService,
		log:          log,
	}
}

func (s *service) Register(ctx context.Context, req *dto.RegisterReq) (*dto.LoginResp, *TokenPair, error) {
	s.log.Info("auth_register_attempt", zap.String("username", req.Username))

	user := &model.User{
		Username:  req.Username,
		Email:     req.Email,
		ClassName: req.ClassName,
		Role:      model.RoleStudent,
		Status:    model.UserStatusActive,
	}
	if err := user.SetPassword(req.Password); err != nil {
		s.log.Error("auth_register_password_hash_failed", zap.String("username", req.Username), zap.Error(err))
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}

	if err := s.repo.Create(ctx, user); err != nil {
		switch {
		case errors.Is(err, ErrUsernameExists):
			s.log.Warn("auth_register_failed_username_exists", zap.String("username", req.Username))
			return nil, nil, errcode.ErrUsernameExists
		case errors.Is(err, ErrEmailExists):
			s.log.Warn("auth_register_failed_email_exists", zap.String("username", req.Username), zap.String("email", req.Email))
			return nil, nil, errcode.ErrEmailExists
		case errors.Is(err, ErrRoleNotFound):
			s.log.Error("auth_register_failed_role_missing", zap.String("username", req.Username), zap.String("role", user.Role))
			return nil, nil, errcode.ErrInternal.WithCause(err)
		default:
			s.log.Error("auth_register_failed", zap.String("username", req.Username), zap.Error(err))
			return nil, nil, errcode.ErrInternal.WithCause(err)
		}
	}

	s.log.Info("auth_register_succeeded", zap.String("username", user.Username), zap.Int64("user_id", user.ID))
	return s.issueLoginResp(user)
}

func (s *service) Login(ctx context.Context, req *dto.LoginReq) (*dto.LoginResp, *TokenPair, error) {
	s.log.Info("auth_login_attempt", zap.String("username", req.Username))

	user, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			s.log.Warn("auth_login_failed_user_not_found", zap.String("username", req.Username))
			return nil, nil, errcode.ErrInvalidCredentials
		}
		s.log.Error("auth_login_failed_lookup", zap.String("username", req.Username), zap.Error(err))
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}

	if !s.ValidatePassword(user, req.Password) {
		s.log.Warn("auth_login_failed_invalid_password", zap.String("username", req.Username), zap.Int64("user_id", user.ID))
		return nil, nil, errcode.ErrInvalidCredentials
	}
	if user.Status == model.UserStatusBanned {
		s.log.Warn("auth_login_failed_account_disabled", zap.String("username", req.Username), zap.Int64("user_id", user.ID))
		return nil, nil, errcode.ErrAccountDisabled
	}
	if user.Status == model.UserStatusLocked {
		s.log.Warn("auth_login_failed_account_locked", zap.String("username", req.Username), zap.Int64("user_id", user.ID))
		return nil, nil, errcode.ErrAccountLocked
	}

	s.log.Info("auth_login_succeeded", zap.String("username", user.Username), zap.Int64("user_id", user.ID))
	return s.issueLoginResp(user)
}

func (s *service) GetProfile(ctx context.Context, userID int64) (*dto.AuthUser, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, errcode.ErrUnauthorized
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	profile := buildAuthUser(user)
	return &profile, nil
}

func (s *service) ChangePassword(ctx context.Context, userID int64, req *dto.ChangePasswordReq) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return errcode.ErrUnauthorized
		}
		return errcode.ErrInternal.WithCause(err)
	}

	if !s.ValidatePassword(user, req.OldPassword) {
		s.log.Warn("auth_change_password_failed_old_password_invalid", zap.Int64("user_id", userID))
		return errcode.ErrOldPasswordInvalid
	}
	if req.OldPassword == req.NewPassword {
		s.log.Warn("auth_change_password_failed_password_unchanged", zap.Int64("user_id", userID))
		return errcode.ErrPasswordUnchanged
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		s.log.Error("auth_change_password_hash_failed", zap.Int64("user_id", userID), zap.Error(err))
		return errcode.ErrInternal.WithCause(err)
	}
	if err := s.repo.UpdatePassword(ctx, userID, user.PasswordHash); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return errcode.ErrUnauthorized
		}
		s.log.Error("auth_change_password_update_failed", zap.Int64("user_id", userID), zap.Error(err))
		return errcode.ErrInternal.WithCause(err)
	}

	s.log.Info("auth_change_password_succeeded", zap.Int64("user_id", userID))
	return nil
}

func (s *service) ValidatePassword(user *model.User, password string) bool {
	return user.CheckPassword(password)
}

func (s *service) issueLoginResp(user *model.User) (*dto.LoginResp, *TokenPair, error) {
	tokens, err := s.tokenService.IssueTokens(user.ID, user.Username, user.Role)
	if err != nil {
		s.log.Error("auth_issue_token_failed", zap.String("username", user.Username), zap.Int64("user_id", user.ID), zap.Error(err))
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.LoginResp{
		AccessToken: tokens.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(tokens.AccessTokenTTL.Seconds()),
		User:        buildAuthUser(user),
	}, tokens, nil
}

func buildAuthUser(user *model.User) dto.AuthUser {
	var name *string
	if user.Name != "" {
		name = &user.Name
	}
	var className *string
	if user.ClassName != "" {
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
