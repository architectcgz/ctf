package commands

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	authports "ctf-platform/internal/module/auth/ports"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/pkg/errcode"
)

const (
	defaultCASValidatePath = "/serviceValidate"
)

type CASService interface {
	Authenticate(ctx context.Context, ticket string) (*dto.LoginResp, *authcontracts.Session, error)
}

type casService struct {
	config       config.CASConfig
	users        casUserRepository
	tokenService authcontracts.TokenService
	log          *zap.Logger
	validator    authports.CASTicketValidator
}

type casUserRepository interface {
	identitycontracts.UserLookupRepository
	identitycontracts.UserWriteRepository
	identitycontracts.UserProfileRepository
}

func NewCASService(cfg config.CASConfig, users casUserRepository, tokenService authcontracts.TokenService, log *zap.Logger, validator authports.CASTicketValidator) CASService {
	if log == nil {
		log = zap.NewNop()
	}

	return &casService{
		config:       cfg,
		users:        users,
		tokenService: tokenService,
		log:          log,
		validator:    validator,
	}
}

func (s *casService) Authenticate(ctx context.Context, ticket string) (*dto.LoginResp, *authcontracts.Session, error) {
	if !s.config.Enabled {
		return nil, nil, errcode.ErrCASDisabled
	}
	if !s.isConfigured() {
		return nil, nil, errcode.ErrCASNotConfigured
	}
	if s.users == nil || s.tokenService == nil || s.validator == nil {
		return nil, nil, errcode.ErrCASNotImplemented
	}

	principal, err := s.validateTicket(ctx, ticket)
	if err != nil {
		return nil, nil, err
	}

	user, err := s.syncUser(ctx, principal)
	if err != nil {
		return nil, nil, err
	}

	return s.issueLoginResp(ctx, user)
}

func (s *casService) validateTicket(ctx context.Context, ticket string) (*authports.CASPrincipal, error) {
	validateURL, err := s.buildValidateURL(ticket)
	if err != nil {
		return nil, errcode.ErrCASNotConfigured.WithCause(err)
	}

	principal, err := s.validator.ValidateTicket(ctx, validateURL)
	if err != nil {
		if errors.Is(err, authports.ErrCASTicketInvalid) {
			return nil, errcode.ErrCASTicketInvalid
		}
		return nil, errcode.ErrServiceUnavailable.WithCause(err)
	}
	return principal, nil
}

func (s *casService) syncUser(ctx context.Context, principal *authports.CASPrincipal) (*model.User, error) {
	user, err := s.users.FindByUsername(ctx, principal.Username)
	if err != nil {
		if !errors.Is(err, identitycontracts.ErrUserNotFound) {
			s.log.Error("auth_cas_find_user_failed", zap.String("username", principal.Username), zap.Error(err))
			return nil, errcode.ErrInternal.WithCause(err)
		}
		if !s.config.AutoProvision {
			return nil, errcode.ErrCASUserNotProvisioned
		}

		user = &model.User{
			Username:  principal.Username,
			Name:      principal.Name,
			Email:     principal.Email,
			StudentNo: principal.StudentNo,
			TeacherNo: principal.TeacherNo,
			Role:      model.RoleStudent,
			ClassName: principal.ClassName,
			Status:    model.UserStatusActive,
		}
		if err := user.SetPassword(randomPassword()); err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		if err := s.users.Create(ctx, user); err != nil {
			return nil, s.mapUserSyncError(err)
		}
		return user, nil
	}

	if user.Status == model.UserStatusBanned {
		return nil, errcode.ErrAccountDisabled
	}
	if user.Status == model.UserStatusLocked && (user.LockedUntil == nil || time.Now().Before(*user.LockedUntil)) {
		return nil, errcode.ErrAccountLocked
	}

	changed := s.mergePrincipal(user, principal)
	if user.Status == model.UserStatusLocked || user.FailedLoginAttempts > 0 || user.LastFailedLoginAt != nil || user.LockedUntil != nil {
		user.Status = model.UserStatusActive
		user.FailedLoginAttempts = 0
		user.LastFailedLoginAt = nil
		user.LockedUntil = nil
		changed = true
	}
	if !changed {
		return user, nil
	}
	if err := s.users.UpdateProfile(ctx, user); err != nil {
		return nil, s.mapUserSyncError(err)
	}
	return user, nil
}

func (s *casService) mergePrincipal(user *model.User, principal *authports.CASPrincipal) bool {
	changed := false
	if principal.Name != "" && user.Name != principal.Name {
		user.Name = principal.Name
		changed = true
	}
	if principal.Email != "" && user.Email != principal.Email {
		user.Email = principal.Email
		changed = true
	}
	if principal.ClassName != "" && user.ClassName != principal.ClassName {
		user.ClassName = principal.ClassName
		changed = true
	}
	if principal.StudentNo != "" && user.StudentNo != principal.StudentNo {
		user.StudentNo = principal.StudentNo
		changed = true
	}
	if principal.TeacherNo != "" && user.TeacherNo != principal.TeacherNo {
		user.TeacherNo = principal.TeacherNo
		changed = true
	}
	return changed
}

func (s *casService) issueLoginResp(ctx context.Context, user *model.User) (*dto.LoginResp, *authcontracts.Session, error) {
	session, err := s.tokenService.CreateSession(ctx, user.ID, user.Username, user.Role)
	if err != nil {
		s.log.Error("auth_cas_create_session_failed", zap.String("username", user.Username), zap.Int64("user_id", user.ID), zap.Error(err))
		return nil, nil, errcode.ErrInternal.WithCause(err)
	}

	return authCommandResponseMapperInst.ToLoginRespPtr(loginRespSource{User: buildAuthUser(user)}), session, nil
}

func (s *casService) mapUserSyncError(err error) error {
	switch {
	case errors.Is(err, identitycontracts.ErrUsernameExists):
		return errcode.ErrUsernameExists
	case errors.Is(err, identitycontracts.ErrEmailExists):
		return errcode.ErrEmailExists
	case errors.Is(err, identitycontracts.ErrStudentNoExists):
		return errcode.ErrStudentNoExists
	case errors.Is(err, identitycontracts.ErrTeacherNoExists):
		return errcode.ErrTeacherNoExists
	case errors.Is(err, identitycontracts.ErrRoleNotFound):
		return errcode.ErrInternal.WithCause(err)
	default:
		return errcode.ErrInternal.WithCause(err)
	}
}

func (s *casService) isConfigured() bool {
	return strings.TrimSpace(s.config.BaseURL) != "" && strings.TrimSpace(s.config.ServiceURL) != ""
}

func (s *casService) buildValidateURL(ticket string) (string, error) {
	validatePath := strings.TrimSpace(s.config.ValidatePath)
	if validatePath == "" {
		validatePath = defaultCASValidatePath
	}
	return s.buildCASURL(validatePath, ticket)
}

func (s *casService) buildCASURL(pathValue, ticket string) (string, error) {
	base, err := url.Parse(strings.TrimRight(s.config.BaseURL, "/"))
	if err != nil {
		return "", err
	}
	base.Path = strings.TrimRight(base.Path, "/") + pathValue

	query := base.Query()
	query.Set("service", s.config.ServiceURL)
	if ticket != "" {
		query.Set("ticket", ticket)
	}
	base.RawQuery = query.Encode()
	return base.String(), nil
}

func randomPassword() string {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("cas_%d", time.Now().UnixNano())
	}
	return "cas_" + hex.EncodeToString(buf)
}
