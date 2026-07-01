package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"go.uber.org/zap"
)

type mockRepository struct {
	createFn           func(ctx context.Context, user *identitycontracts.User) error
	findByIDFn         func(ctx context.Context, userID int64) (*identitycontracts.User, error)
	findByUsernameFn   func(ctx context.Context, username string) (*identitycontracts.User, error)
	listFn             func(ctx context.Context, filter identitycontracts.UserListFilter) ([]*identitycontracts.User, int64, error)
	updateFn           func(ctx context.Context, user *identitycontracts.User) error
	deleteFn           func(ctx context.Context, userID int64) error
	updatePasswordFn   func(ctx context.Context, userID int64, newHash string) error
	updateLoginStateFn func(ctx context.Context, userID int64, failedAttempts int, lastFailedAt, lockedUntil *time.Time, status string) error
	updateProfileFn    func(ctx context.Context, user *identitycontracts.User) error
}

func (m *mockRepository) Create(ctx context.Context, user *identitycontracts.User) error {
	if m.createFn == nil {
		return nil
	}
	return m.createFn(ctx, user)
}

func (m *mockRepository) FindByUsername(ctx context.Context, username string) (*identitycontracts.User, error) {
	if m.findByUsernameFn == nil {
		return nil, identitycontracts.ErrUserNotFound
	}
	return m.findByUsernameFn(ctx, username)
}

func (m *mockRepository) FindByID(ctx context.Context, userID int64) (*identitycontracts.User, error) {
	if m.findByIDFn == nil {
		return nil, identitycontracts.ErrUserNotFound
	}
	return m.findByIDFn(ctx, userID)
}

func (m *mockRepository) List(ctx context.Context, filter identitycontracts.UserListFilter) ([]*identitycontracts.User, int64, error) {
	if m.listFn == nil {
		return nil, 0, nil
	}
	return m.listFn(ctx, filter)
}

func (m *mockRepository) Update(ctx context.Context, user *identitycontracts.User) error {
	if m.updateFn == nil {
		return nil
	}
	return m.updateFn(ctx, user)
}

func (m *mockRepository) Delete(ctx context.Context, userID int64) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, userID)
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

func (m *mockRepository) UpdateProfile(ctx context.Context, user *identitycontracts.User) error {
	if m.updateProfileFn == nil {
		return nil
	}
	return m.updateProfileFn(ctx, user)
}

type mockTokenService struct {
	issueFn func(userID int64, username, role string) (*authcontracts.Session, error)
}

func (m *mockTokenService) CreateSession(_ context.Context, userID int64, username, role string) (*authcontracts.Session, error) {
	if m.issueFn == nil {
		return nil, errors.New("unexpected call")
	}
	return m.issueFn(userID, username, role)
}

func (m *mockTokenService) GetSession(context.Context, string) (*authcontracts.Session, error) {
	return nil, nil
}

func (m *mockTokenService) DeleteSession(context.Context, string) error {
	return nil
}

func (m *mockTokenService) IssueWSTicket(context.Context, authctx.CurrentUser) (*authcontracts.WSTicket, error) {
	return nil, nil
}

func (m *mockTokenService) ConsumeWSTicket(context.Context, string) (*authctx.CurrentUser, error) {
	return nil, nil
}

func (m *mockTokenService) IssueMCPToken(context.Context, authctx.CurrentUser) (*authcontracts.MCPToken, error) {
	return nil, nil
}

func (m *mockTokenService) ResolveMCPToken(context.Context, string) (*authctx.CurrentUser, error) {
	return nil, nil
}

func (m *mockTokenService) RevokeAllUserSessions(context.Context, int64) error {
	return nil
}

func (m *mockTokenService) CurrentSessionVersion(context.Context, int64) (int64, error) {
	return 0, nil
}

func (m *mockTokenService) ListUserSessions(context.Context, int64) ([]authcontracts.Session, error) {
	return nil, nil
}

func TestServiceRegisterSuccess(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{
		createFn: func(ctx context.Context, user *identitycontracts.User) error {
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
		issueFn: func(userID int64, username, role string) (*authcontracts.Session, error) {
			if userID != 101 {
				t.Fatalf("unexpected user id: %d", userID)
			}
			return &authcontracts.Session{
				ID:        "session-1",
				UserID:    userID,
				Username:  username,
				Role:      role,
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			}, nil
		},
	}

	service := NewService(repo, tokenService, config.RateLimitPolicyConfig{}, zap.NewNop())
	resp, tokens, err := service.Register(context.Background(), RegisterInput{
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
	if tokens.ID != "session-1" {
		t.Fatalf("unexpected session id: %s", tokens.ID)
	}
}

func TestServiceRegisterTrimsEmail(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{
		createFn: func(ctx context.Context, user *identitycontracts.User) error {
			if user.Email != "alice@example.com" {
				t.Fatalf("expected trimmed email, got %q", user.Email)
			}
			user.ID = 102
			return nil
		},
	}
	tokenService := &mockTokenService{
		issueFn: func(userID int64, username, role string) (*authcontracts.Session, error) {
			return &authcontracts.Session{
				ID:        "session-2",
				UserID:    userID,
				Username:  username,
				Role:      role,
				ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
			}, nil
		},
	}

	service := NewService(repo, tokenService, config.RateLimitPolicyConfig{}, zap.NewNop())
	_, _, err := service.Register(context.Background(), RegisterInput{
		Username:  "alice_trim",
		Password:  "Password123",
		Email:     "  alice@example.com  ",
		ClassName: "Class A",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
}

func TestServiceRegisterRoleNotFound(t *testing.T) {
	t.Parallel()

	service := NewService(&mockRepository{
		createFn: func(context.Context, *identitycontracts.User) error {
			return identitycontracts.ErrRoleNotFound
		},
	}, &mockTokenService{
		issueFn: func(int64, string, string) (*authcontracts.Session, error) {
			return nil, errors.New("should not be called")
		},
	}, config.RateLimitPolicyConfig{}, zap.NewNop())

	_, _, err := service.Register(context.Background(), RegisterInput{
		Username: "alice_1",
		Password: "Password123",
	})
	appErr, ok := err.(*apperror.AppError)
	if !ok || appErr.Code != apperror.ErrInternal.Code {
		t.Fatalf("expected internal error, got %v", err)
	}
}

func TestServiceLoginInvalidPassword(t *testing.T) {
	t.Parallel()

	user := &identitycontracts.User{
		ID:       1,
		Username: "alice_1",
		Role:     identitycontracts.RoleStudent,
		Status:   identitycontracts.UserStatusActive,
	}
	if err := user.SetPassword("Password123"); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	service := NewService(&mockRepository{
		findByUsernameFn: func(context.Context, string) (*identitycontracts.User, error) {
			return user, nil
		},
		updateLoginStateFn: func(context.Context, int64, int, *time.Time, *time.Time, string) error {
			return nil
		},
	}, &mockTokenService{
		issueFn: func(int64, string, string) (*authcontracts.Session, error) {
			return nil, errors.New("should not be called")
		},
	}, config.RateLimitPolicyConfig{Limit: 3, Window: time.Minute, LockDuration: 15 * time.Minute}, zap.NewNop())

	_, _, err := service.Login(context.Background(), LoginInput{
		Username: "alice_1",
		Password: "wrong-password",
	})
	if !errors.Is(err, authcontracts.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestServiceLoginLocksAccountAfterExceededAttempts(t *testing.T) {
	t.Parallel()

	user := &identitycontracts.User{
		ID:                  2,
		Username:            "alice_2",
		Role:                identitycontracts.RoleStudent,
		Status:              identitycontracts.UserStatusActive,
		FailedLoginAttempts: 2,
	}
	if err := user.SetPassword("Password123"); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}
	lastFailedAt := time.Now().Add(-10 * time.Second)
	user.LastFailedLoginAt = &lastFailedAt

	service := NewService(&mockRepository{
		findByUsernameFn: func(context.Context, string) (*identitycontracts.User, error) {
			return user, nil
		},
		updateLoginStateFn: func(_ context.Context, _ int64, failedAttempts int, lastFailedAt, lockedUntil *time.Time, status string) error {
			user.FailedLoginAttempts = failedAttempts
			user.LastFailedLoginAt = lastFailedAt
			user.LockedUntil = lockedUntil
			user.Status = status
			return nil
		},
	}, &mockTokenService{
		issueFn: func(int64, string, string) (*authcontracts.Session, error) {
			t.Fatal("CreateSession() should not be called")
			return nil, nil
		},
	}, config.RateLimitPolicyConfig{Limit: 3, Window: time.Minute, LockDuration: 15 * time.Minute}, zap.NewNop())

	_, _, err := service.Login(context.Background(), LoginInput{
		Username: "alice_2",
		Password: "wrong-password",
	})
	if !errors.Is(err, authcontracts.ErrLoginTooFrequent) {
		t.Fatalf("expected ErrLoginTooFrequent, got %v", err)
	}
	if user.Status != identitycontracts.UserStatusLocked || user.LockedUntil == nil {
		t.Fatalf("expected locked user state, got %+v", user)
	}
}

func TestServiceLoginUnlocksExpiredAccountAndSucceeds(t *testing.T) {
	t.Parallel()

	user := &identitycontracts.User{
		ID:                  3,
		Username:            "alice_3",
		Role:                identitycontracts.RoleStudent,
		Status:              identitycontracts.UserStatusLocked,
		FailedLoginAttempts: 3,
	}
	if err := user.SetPassword("Password123"); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}
	lockedUntil := time.Now().Add(-time.Minute)
	user.LockedUntil = &lockedUntil

	service := NewService(&mockRepository{
		findByUsernameFn: func(context.Context, string) (*identitycontracts.User, error) {
			return user, nil
		},
		updateLoginStateFn: func(_ context.Context, _ int64, failedAttempts int, lastFailedAt, lockedUntil *time.Time, status string) error {
			user.FailedLoginAttempts = failedAttempts
			user.LastFailedLoginAt = lastFailedAt
			user.LockedUntil = lockedUntil
			user.Status = status
			return nil
		},
	}, &mockTokenService{
		issueFn: func(userID int64, username, role string) (*authcontracts.Session, error) {
			return &authcontracts.Session{
				ID:        "session-3",
				UserID:    userID,
				Username:  username,
				Role:      role,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}, nil
		},
	}, config.RateLimitPolicyConfig{Limit: 3, Window: time.Minute, LockDuration: 15 * time.Minute}, zap.NewNop())

	_, _, err := service.Login(context.Background(), LoginInput{
		Username: "alice_3",
		Password: "Password123",
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if user.Status != identitycontracts.UserStatusActive || user.FailedLoginAttempts != 0 || user.LockedUntil != nil {
		t.Fatalf("expected login tracking to be reset, got %+v", user)
	}
}
