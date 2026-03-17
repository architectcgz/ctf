package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/config"
	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
	jwtpkg "ctf-platform/pkg/jwt"
)

type mockRepository struct {
	createFn           func(ctx context.Context, user *model.User) error
	findByUsernameFn   func(ctx context.Context, username string) (*model.User, error)
	findByIDFn         func(ctx context.Context, userID int64) (*model.User, error)
	updatePasswordFn   func(ctx context.Context, userID int64, newHash string) error
	updateLoginStateFn func(ctx context.Context, userID int64, failedAttempts int, lastFailedAt, lockedUntil *time.Time, status string) error
	updateCASProfileFn func(ctx context.Context, user *model.User) error
}

func (m *mockRepository) Create(ctx context.Context, user *model.User) error {
	if m.createFn == nil {
		return nil
	}
	return m.createFn(ctx, user)
}

func (m *mockRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	if m.findByUsernameFn == nil {
		return nil, ErrUserNotFound
	}
	return m.findByUsernameFn(ctx, username)
}

func (m *mockRepository) FindByID(ctx context.Context, userID int64) (*model.User, error) {
	if m.findByIDFn == nil {
		return nil, ErrUserNotFound
	}
	return m.findByIDFn(ctx, userID)
}

func (m *mockRepository) UpdatePassword(ctx context.Context, userID int64, newHash string) error {
	if m.updatePasswordFn == nil {
		return nil
	}
	return m.updatePasswordFn(ctx, userID, newHash)
}

func (m *mockRepository) UpdateLoginState(ctx context.Context, userID int64, failedAttempts int, lastFailedAt, lockedUntil *time.Time, status string) error {
	if m.updateLoginStateFn == nil {
		return nil
	}
	return m.updateLoginStateFn(ctx, userID, failedAttempts, lastFailedAt, lockedUntil, status)
}

func (m *mockRepository) UpdateCASProfile(ctx context.Context, user *model.User) error {
	if m.updateCASProfileFn == nil {
		return nil
	}
	return m.updateCASProfileFn(ctx, user)
}

func TestServiceChangePasswordSuccess(t *testing.T) {
	t.Parallel()

	user := &model.User{
		ID:       1,
		Username: "alice_1",
		Role:     model.RoleStudent,
		Status:   model.UserStatusActive,
	}
	if err := user.SetPassword("Password123"); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}
	oldHash := user.PasswordHash

	var updatedHash string
	service := NewService(&mockRepository{
		findByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			if userID != user.ID {
				t.Fatalf("unexpected user id: %d", userID)
			}
			return user, nil
		},
		updatePasswordFn: func(ctx context.Context, userID int64, newHash string) error {
			if userID != user.ID {
				t.Fatalf("unexpected user id: %d", userID)
			}
			updatedHash = newHash
			return nil
		},
	}, &mockTokenService{}, config.RateLimitPolicyConfig{}, zap.NewNop())

	err := service.ChangePassword(context.Background(), user.ID, &dto.ChangePasswordReq{
		OldPassword: "Password123",
		NewPassword: "Password456",
	})
	if err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	if updatedHash == "" || updatedHash == oldHash {
		t.Fatalf("expected password hash to be updated")
	}
	if !user.CheckPassword("Password456") {
		t.Fatalf("expected user password to be replaced with new password")
	}
}

func TestServiceChangePasswordOldPasswordInvalid(t *testing.T) {
	t.Parallel()

	user := &model.User{ID: 1, Username: "alice_1"}
	if err := user.SetPassword("Password123"); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	service := NewService(&mockRepository{
		findByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			return user, nil
		},
		updatePasswordFn: func(ctx context.Context, userID int64, newHash string) error {
			t.Fatalf("UpdatePassword() should not be called")
			return nil
		},
	}, &mockTokenService{}, config.RateLimitPolicyConfig{}, zap.NewNop())

	err := service.ChangePassword(context.Background(), user.ID, &dto.ChangePasswordReq{
		OldPassword: "wrong-password",
		NewPassword: "Password456",
	})
	if !errors.Is(err, errcode.ErrOldPasswordInvalid) {
		t.Fatalf("expected old password invalid, got %v", err)
	}
}

func TestServiceChangePasswordRejectsSamePassword(t *testing.T) {
	t.Parallel()

	user := &model.User{ID: 1, Username: "alice_1"}
	if err := user.SetPassword("Password123"); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	service := NewService(&mockRepository{
		findByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			return user, nil
		},
		updatePasswordFn: func(ctx context.Context, userID int64, newHash string) error {
			t.Fatalf("UpdatePassword() should not be called")
			return nil
		},
	}, &mockTokenService{}, config.RateLimitPolicyConfig{}, zap.NewNop())

	err := service.ChangePassword(context.Background(), user.ID, &dto.ChangePasswordReq{
		OldPassword: "Password123",
		NewPassword: "Password123",
	})
	if !errors.Is(err, errcode.ErrPasswordUnchanged) {
		t.Fatalf("expected password unchanged error, got %v", err)
	}
}

type mockTokenService struct {
	issueFn func(userID int64, username, role string) (*TokenPair, error)
}

func (m *mockTokenService) IssueTokens(userID int64, username, role string) (*TokenPair, error) {
	return m.IssueTokensWithContext(context.Background(), userID, username, role)
}

func (m *mockTokenService) IssueTokensWithContext(_ context.Context, userID int64, username, role string) (*TokenPair, error) {
	if m.issueFn == nil {
		return nil, errors.New("unexpected call")
	}
	return m.issueFn(userID, username, role)
}

func (m *mockTokenService) RefreshAccessToken(ctx context.Context, refreshToken string) (*dtoRefreshPayload, error) {
	return nil, nil
}

func (m *mockTokenService) RevokeToken(ctx context.Context, jti string, ttl time.Duration) error {
	return nil
}

func (m *mockTokenService) ClearRefreshSession(ctx context.Context, userID int64, refreshJTI string) error {
	return nil
}

func (m *mockTokenService) IsRevoked(ctx context.Context, jti string) (bool, error) {
	return false, nil
}

func (m *mockTokenService) ParseToken(tokenString string) (*jwtpkg.Claims, error) {
	return nil, nil
}

func (m *mockTokenService) IssueWSTicket(ctx context.Context, user authctx.CurrentUser) (*WSTicket, error) {
	return nil, nil
}

func (m *mockTokenService) ConsumeWSTicket(ctx context.Context, ticket string) (*authctx.CurrentUser, error) {
	return nil, nil
}

func TestServiceRegisterSuccess(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{
		createFn: func(ctx context.Context, user *model.User) error {
			if user.Username != "alice_1" {
				t.Fatalf("unexpected username: %s", user.Username)
			}
			if user.PasswordHash == "" || user.PasswordHash == "Password123" {
				t.Fatalf("password hash was not generated")
			}
			user.ID = 101
			return nil
		},
	}
	tokenService := &mockTokenService{
		issueFn: func(userID int64, username, role string) (*TokenPair, error) {
			if userID != 101 {
				t.Fatalf("unexpected user id: %d", userID)
			}
			return &TokenPair{
				AccessToken:     "access-token",
				RefreshToken:    "refresh-token",
				AccessTokenTTL:  15 * time.Minute,
				RefreshTokenTTL: 7 * 24 * time.Hour,
			}, nil
		},
	}

	service := NewService(repo, tokenService, config.RateLimitPolicyConfig{}, zap.NewNop())
	resp, tokens, err := service.Register(context.Background(), &dto.RegisterReq{
		Username:  "alice_1",
		Password:  "Password123",
		Email:     "alice@example.com",
		ClassName: "Class A",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if resp.User.ID != 101 {
		t.Fatalf("unexpected user id: %d", resp.User.ID)
	}
	if tokens.RefreshToken != "refresh-token" {
		t.Fatalf("unexpected refresh token: %s", tokens.RefreshToken)
	}
}

func TestServiceRegisterRoleNotFound(t *testing.T) {
	t.Parallel()

	service := NewService(&mockRepository{
		createFn: func(ctx context.Context, user *model.User) error {
			return ErrRoleNotFound
		},
	}, &mockTokenService{
		issueFn: func(userID int64, username, role string) (*TokenPair, error) {
			return nil, errors.New("should not be called")
		},
	}, config.RateLimitPolicyConfig{}, zap.NewNop())

	_, _, err := service.Register(context.Background(), &dto.RegisterReq{
		Username: "alice_1",
		Password: "Password123",
	})
	appErr, ok := err.(*errcode.AppError)
	if !ok || appErr.Code != errcode.ErrInternal.Code {
		t.Fatalf("expected internal error, got %v", err)
	}
}

func TestServiceLoginInvalidPassword(t *testing.T) {
	t.Parallel()

	user := &model.User{
		ID:       1,
		Username: "alice_1",
		Role:     model.RoleStudent,
		Status:   model.UserStatusActive,
	}
	if err := user.SetPassword("Password123"); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	service := NewService(&mockRepository{
		findByUsernameFn: func(ctx context.Context, username string) (*model.User, error) {
			return user, nil
		},
		updateLoginStateFn: func(ctx context.Context, userID int64, failedAttempts int, lastFailedAt, lockedUntil *time.Time, status string) error {
			if userID != user.ID {
				t.Fatalf("unexpected user id: %d", userID)
			}
			user.FailedLoginAttempts = failedAttempts
			user.LastFailedLoginAt = lastFailedAt
			user.LockedUntil = lockedUntil
			user.Status = status
			return nil
		},
	}, &mockTokenService{
		issueFn: func(userID int64, username, role string) (*TokenPair, error) {
			return nil, errors.New("should not be called")
		},
	}, config.RateLimitPolicyConfig{Limit: 3, Window: time.Minute, LockDuration: 15 * time.Minute}, zap.NewNop())

	_, _, err := service.Login(context.Background(), &dto.LoginReq{
		Username: "alice_1",
		Password: "wrong-password",
	})
	if !errors.Is(err, errcode.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestServiceLoginLocksAccountAfterExceededAttempts(t *testing.T) {
	t.Parallel()

	user := &model.User{
		ID:                  2,
		Username:            "alice_2",
		Role:                model.RoleStudent,
		Status:              model.UserStatusActive,
		FailedLoginAttempts: 2,
	}
	if err := user.SetPassword("Password123"); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}
	lastFailedAt := time.Now().Add(-10 * time.Second)
	user.LastFailedLoginAt = &lastFailedAt

	service := NewService(&mockRepository{
		findByUsernameFn: func(ctx context.Context, username string) (*model.User, error) {
			return user, nil
		},
		updateLoginStateFn: func(ctx context.Context, userID int64, failedAttempts int, lastFailedAt, lockedUntil *time.Time, status string) error {
			user.FailedLoginAttempts = failedAttempts
			user.LastFailedLoginAt = lastFailedAt
			user.LockedUntil = lockedUntil
			user.Status = status
			return nil
		},
	}, &mockTokenService{
		issueFn: func(userID int64, username, role string) (*TokenPair, error) {
			t.Fatal("IssueTokens() should not be called")
			return nil, nil
		},
	}, config.RateLimitPolicyConfig{Limit: 3, Window: time.Minute, LockDuration: 15 * time.Minute}, zap.NewNop())

	_, _, err := service.Login(context.Background(), &dto.LoginReq{
		Username: "alice_2",
		Password: "wrong-password",
	})
	if !errors.Is(err, errcode.ErrLoginTooFrequent) {
		t.Fatalf("expected ErrLoginTooFrequent, got %v", err)
	}
	if user.Status != model.UserStatusLocked || user.LockedUntil == nil {
		t.Fatalf("expected locked user state, got %+v", user)
	}
}

func TestServiceLoginUnlocksExpiredAccountAndSucceeds(t *testing.T) {
	t.Parallel()

	user := &model.User{
		ID:                  3,
		Username:            "alice_3",
		Role:                model.RoleStudent,
		Status:              model.UserStatusLocked,
		FailedLoginAttempts: 3,
	}
	if err := user.SetPassword("Password123"); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}
	lockedUntil := time.Now().Add(-time.Minute)
	user.LockedUntil = &lockedUntil

	service := NewService(&mockRepository{
		findByUsernameFn: func(ctx context.Context, username string) (*model.User, error) {
			return user, nil
		},
		updateLoginStateFn: func(ctx context.Context, userID int64, failedAttempts int, lastFailedAt, lockedUntil *time.Time, status string) error {
			user.FailedLoginAttempts = failedAttempts
			user.LastFailedLoginAt = lastFailedAt
			user.LockedUntil = lockedUntil
			user.Status = status
			return nil
		},
	}, &mockTokenService{
		issueFn: func(userID int64, username, role string) (*TokenPair, error) {
			return &TokenPair{AccessToken: "access", RefreshToken: "refresh", AccessTokenTTL: time.Minute}, nil
		},
	}, config.RateLimitPolicyConfig{Limit: 3, Window: time.Minute, LockDuration: 15 * time.Minute}, zap.NewNop())

	resp, _, err := service.Login(context.Background(), &dto.LoginReq{
		Username: "alice_3",
		Password: "Password123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if resp == nil || resp.AccessToken == "" {
		t.Fatalf("expected login response, got %+v", resp)
	}
	if user.Status != model.UserStatusActive || user.FailedLoginAttempts != 0 || user.LockedUntil != nil {
		t.Fatalf("expected reset login tracking, got %+v", user)
	}
}

func TestBuildAuthUserIncludesName(t *testing.T) {
	t.Parallel()

	user := &model.User{
		ID:        1,
		Username:  "alice_1",
		Name:      "Alice Zhang",
		Role:      model.RoleStudent,
		ClassName: "Class A",
	}

	profile := buildAuthUser(user)
	if profile.Name == nil || *profile.Name != "Alice Zhang" {
		t.Fatalf("expected profile name, got %+v", profile)
	}
	if profile.ClassName == nil || *profile.ClassName != "Class A" {
		t.Fatalf("expected class name, got %+v", profile)
	}
}
