package commands

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

type Service interface {
	Register(ctx context.Context, req RegisterInput) (*LoginResp, *authcontracts.Session, error)
	Login(ctx context.Context, req LoginInput) (*LoginResp, *authcontracts.Session, error)
	ValidatePassword(user *identitycontracts.User, password string) bool
}

type service struct {
	users        authUserRepository
	tokenService authcontracts.TokenService
	log          *zap.Logger
	loginPolicy  config.RateLimitPolicyConfig
}

type authUserRepository interface {
	identitycontracts.UserLookupRepository
	identitycontracts.UserWriteRepository
	identitycontracts.UserLoginStateRepository
}

func NewService(users authUserRepository, tokenService authcontracts.TokenService, loginPolicy config.RateLimitPolicyConfig, log *zap.Logger) Service {
	if log == nil {
		log = zap.NewNop()
	}

	return &service{
		users:        users,
		tokenService: tokenService,
		log:          log,
		loginPolicy:  loginPolicy,
	}
}

func (s *service) Register(ctx context.Context, req RegisterInput) (*LoginResp, *authcontracts.Session, error) {
	s.log.Info("auth_register_attempt", zap.String("username", req.Username))

	user := &identitycontracts.User{
		Username:  req.Username,
		Email:     strings.TrimSpace(req.Email),
		ClassName: req.ClassName,
		Role:      identitycontracts.RoleStudent,
		Status:    identitycontracts.UserStatusActive,
	}
	if err := user.SetPassword(req.Password); err != nil {
		s.log.Error("auth_register_password_hash_failed", zap.String("username", req.Username), zap.Error(err))
		return nil, nil, apperror.ErrInternal.WithCause(err)
	}

	if err := s.users.Create(ctx, user); err != nil {
		switch {
		case errors.Is(err, identitycontracts.ErrUsernameExists):
			s.log.Warn("auth_register_failed_username_exists", zap.String("username", req.Username))
			return nil, nil, identitycontracts.ErrDuplicateUsername
		case errors.Is(err, identitycontracts.ErrEmailExists):
			s.log.Warn("auth_register_failed_email_exists", zap.String("username", req.Username), zap.String("email", req.Email))
			return nil, nil, identitycontracts.ErrDuplicateEmail
		case errors.Is(err, identitycontracts.ErrRoleNotFound):
			s.log.Error("auth_register_failed_role_missing", zap.String("username", req.Username), zap.String("role", user.Role))
			return nil, nil, apperror.ErrInternal.WithCause(err)
		default:
			s.log.Error("auth_register_failed", zap.String("username", req.Username), zap.Error(err))
			return nil, nil, apperror.ErrInternal.WithCause(err)
		}
	}

	s.log.Info("auth_register_succeeded", zap.String("username", user.Username), zap.Int64("user_id", user.ID))
	return s.issueLoginResp(ctx, user)
}

func (s *service) Login(ctx context.Context, req LoginInput) (*LoginResp, *authcontracts.Session, error) {
	s.log.Info("auth_login_attempt", zap.String("username", req.Username))

	user, err := s.users.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, identitycontracts.ErrUserNotFound) {
			s.log.Warn("auth_login_failed_user_not_found", zap.String("username", req.Username))
			return nil, nil, authcontracts.ErrInvalidCredentials
		}
		s.log.Error("auth_login_failed_lookup", zap.String("username", req.Username), zap.Error(err))
		return nil, nil, apperror.ErrInternal.WithCause(err)
	}

	if user.Status == identitycontracts.UserStatusBanned {
		s.log.Warn("auth_login_failed_account_disabled", zap.String("username", req.Username), zap.Int64("user_id", user.ID))
		return nil, nil, authcontracts.ErrAccountDisabled
	}
	if user.Status == identitycontracts.UserStatusLocked {
		if user.LockedUntil == nil || time.Now().Before(*user.LockedUntil) {
			s.log.Warn("auth_login_failed_account_locked", zap.String("username", req.Username), zap.Int64("user_id", user.ID))
			return nil, nil, authcontracts.ErrAccountLocked
		}
		if err := s.resetLoginTracking(ctx, user, identitycontracts.UserStatusActive); err != nil {
			s.log.Error("auth_login_failed_unlock_expired_lock", zap.String("username", req.Username), zap.Int64("user_id", user.ID), zap.Error(err))
			return nil, nil, apperror.ErrInternal.WithCause(err)
		}
	}

	if !s.ValidatePassword(user, req.Password) {
		locked, updateErr := s.recordFailedLogin(ctx, user, time.Now())
		if updateErr != nil {
			s.log.Error("auth_login_failed_record_attempt", zap.String("username", req.Username), zap.Int64("user_id", user.ID), zap.Error(updateErr))
			return nil, nil, apperror.ErrInternal.WithCause(updateErr)
		}
		s.log.Warn("auth_login_failed_invalid_password", zap.String("username", req.Username), zap.Int64("user_id", user.ID), zap.Bool("locked", locked))
		if locked {
			return nil, nil, authcontracts.ErrLoginTooFrequent
		}
		return nil, nil, authcontracts.ErrInvalidCredentials
	}

	if user.FailedLoginAttempts > 0 || user.LockedUntil != nil || user.Status == identitycontracts.UserStatusLocked {
		nextStatus := user.Status
		if nextStatus == identitycontracts.UserStatusLocked {
			nextStatus = identitycontracts.UserStatusActive
		}
		if err := s.resetLoginTracking(ctx, user, nextStatus); err != nil {
			s.log.Error("auth_login_failed_reset_attempts", zap.String("username", req.Username), zap.Int64("user_id", user.ID), zap.Error(err))
			return nil, nil, apperror.ErrInternal.WithCause(err)
		}
	}

	s.log.Info("auth_login_succeeded", zap.String("username", user.Username), zap.Int64("user_id", user.ID))
	return s.issueLoginResp(ctx, user)
}

func (s *service) ValidatePassword(user *identitycontracts.User, password string) bool {
	return user.CheckPassword(password)
}

func (s *service) issueLoginResp(ctx context.Context, user *identitycontracts.User) (*LoginResp, *authcontracts.Session, error) {
	session, err := s.tokenService.CreateSession(ctx, user.ID, user.Username, user.Role)
	if err != nil {
		s.log.Error("auth_create_session_failed", zap.String("username", user.Username), zap.Int64("user_id", user.ID), zap.Error(err))
		return nil, nil, apperror.ErrInternal.WithCause(err)
	}

	return authCommandResponseMapperInst.ToLoginRespPtr(loginRespSource{User: buildAuthUser(user)}), session, nil
}

func (s *service) recordFailedLogin(ctx context.Context, user *identitycontracts.User, now time.Time) (bool, error) {
	failedAttempts := user.FailedLoginAttempts
	if user.LastFailedLoginAt == nil || s.loginPolicy.Window <= 0 || now.Sub(*user.LastFailedLoginAt) > s.loginPolicy.Window {
		failedAttempts = 0
	}
	failedAttempts++

	var lockedUntil *time.Time
	status := user.Status
	locked := false
	if s.loginPolicy.Limit > 0 && failedAttempts >= s.loginPolicy.Limit {
		until := now.Add(s.loginPolicy.LockDuration)
		lockedUntil = &until
		status = identitycontracts.UserStatusLocked
		locked = true
	}

	lastFailedAt := &now
	if err := s.users.UpdateLoginState(ctx, user.ID, failedAttempts, lastFailedAt, lockedUntil, status); err != nil {
		return false, err
	}

	user.FailedLoginAttempts = failedAttempts
	user.LastFailedLoginAt = lastFailedAt
	user.LockedUntil = lockedUntil
	user.Status = status
	return locked, nil
}

func (s *service) resetLoginTracking(ctx context.Context, user *identitycontracts.User, status string) error {
	if err := s.users.UpdateLoginState(ctx, user.ID, 0, nil, nil, status); err != nil {
		return err
	}
	user.FailedLoginAttempts = 0
	user.LastFailedLoginAt = nil
	user.LockedUntil = nil
	user.Status = status
	return nil
}
