package commands

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	authports "ctf-platform/internal/module/auth/ports"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/pkg/errcode"
)

func TestCASServiceAuthenticateAutoProvisionSuccess(t *testing.T) {
	t.Parallel()

	repo := &mockRepository{
		findByUsernameFn: func(context.Context, string) (*model.User, error) {
			return nil, identitycontracts.ErrUserNotFound
		},
		createFn: func(_ context.Context, user *model.User) error {
			if user.Username != "cas_user_1" {
				t.Fatalf("unexpected username: %s", user.Username)
			}
			if user.Role != model.RoleStudent {
				t.Fatalf("unexpected role: %s", user.Role)
			}
			if user.Name != "CAS User" || user.Email != "cas_user_1@example.edu" || user.ClassName != "CTF-1" {
				t.Fatalf("unexpected provisioned profile: %+v", user)
			}
			if user.StudentNo != "20260001" {
				t.Fatalf("unexpected student no: %s", user.StudentNo)
			}
			if user.PasswordHash == "" {
				t.Fatalf("expected password hash to be generated")
			}
			user.ID = 101
			return nil
		},
	}
	tokenService := &mockTokenService{
		issueFn: func(userID int64, username, role string) (*authcontracts.Session, error) {
			if userID != 101 || username != "cas_user_1" || role != model.RoleStudent {
				t.Fatalf("unexpected session issue params: %d %s %s", userID, username, role)
			}
			return &authcontracts.Session{
				ID:        "cas-session-1",
				UserID:    userID,
				Username:  username,
				Role:      role,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}, nil
		},
	}

	validator := &mockCASTicketValidator{
		validateTicketFn: func(_ context.Context, validateURL string) (*authports.CASPrincipal, error) {
			expected := "https://cas.example.edu/cas/serviceValidate?service=https%3A%2F%2Fctf.example.edu%2Fapi%2Fv1%2Fauth%2Fcas%2Fcallback&ticket=ST-1"
			if validateURL != expected {
				t.Fatalf("unexpected validate url: %s", validateURL)
			}
			return &authports.CASPrincipal{
				Username:  "cas_user_1",
				Name:      "CAS User",
				Email:     "cas_user_1@example.edu",
				ClassName: "CTF-1",
				StudentNo: "20260001",
			}, nil
		},
	}

	service := NewCASService(config.CASConfig{
		Enabled:       true,
		BaseURL:       "https://cas.example.edu/cas",
		ServiceURL:    "https://ctf.example.edu/api/v1/auth/cas/callback",
		AutoProvision: true,
	}, repo, tokenService, zap.NewNop(), validator)

	resp, tokens, err := service.Authenticate(context.Background(), "ST-1")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if resp.User.Username != "cas_user_1" || resp.User.Role != model.RoleStudent {
		t.Fatalf("unexpected login response user: %+v", resp.User)
	}
	if tokens.ID != "cas-session-1" {
		t.Fatalf("unexpected session id: %s", tokens.ID)
	}
}

func TestCASServiceAuthenticateExistingUserSyncsProfileAndUnlocksExpired(t *testing.T) {
	t.Parallel()

	expired := time.Now().Add(-time.Minute)
	user := &model.User{
		ID:                  202,
		Username:            "cas_user_2",
		Name:                "Old Name",
		Role:                model.RoleStudent,
		Status:              model.UserStatusLocked,
		ClassName:           "Old Class",
		FailedLoginAttempts: 3,
		LastFailedLoginAt:   &expired,
		LockedUntil:         &expired,
	}

	repo := &mockRepository{
		findByUsernameFn: func(context.Context, string) (*model.User, error) {
			return user, nil
		},
		updateProfileFn: func(_ context.Context, updated *model.User) error {
			if updated.Name != "Updated Name" || updated.Email != "cas_user_2@example.edu" {
				t.Fatalf("unexpected updated profile: %+v", updated)
			}
			if updated.ClassName != "CTF-2" || updated.StudentNo != "20260002" {
				t.Fatalf("unexpected updated attributes: %+v", updated)
			}
			if updated.Status != model.UserStatusActive || updated.FailedLoginAttempts != 0 {
				t.Fatalf("expected login tracking reset, got %+v", updated)
			}
			if updated.LastFailedLoginAt != nil || updated.LockedUntil != nil {
				t.Fatalf("expected lock fields cleared, got %+v", updated)
			}
			return nil
		},
	}
	tokenService := &mockTokenService{
		issueFn: func(userID int64, username, role string) (*authcontracts.Session, error) {
			return &authcontracts.Session{
				ID:        "cas-session-2",
				UserID:    userID,
				Username:  username,
				Role:      role,
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}, nil
		},
	}

	validator := &mockCASTicketValidator{
		validateTicketFn: func(context.Context, string) (*authports.CASPrincipal, error) {
			return &authports.CASPrincipal{
				Username:  "cas_user_2",
				Name:      "Updated Name",
				Email:     "cas_user_2@example.edu",
				ClassName: "CTF-2",
				StudentNo: "20260002",
			}, nil
		},
	}

	service := NewCASService(config.CASConfig{
		Enabled:    true,
		BaseURL:    "https://cas.example.edu/cas",
		ServiceURL: "https://ctf.example.edu/api/v1/auth/cas/callback",
	}, repo, tokenService, zap.NewNop(), validator)

	resp, _, err := service.Authenticate(context.Background(), "ST-2")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if resp.User.Username != "cas_user_2" {
		t.Fatalf("unexpected username: %+v", resp.User)
	}
}

func TestCASServiceAuthenticateRejectsUserWhenAutoProvisionDisabled(t *testing.T) {
	t.Parallel()

	service := NewCASService(config.CASConfig{
		Enabled:       true,
		BaseURL:       "https://cas.example.edu/cas",
		ServiceURL:    "https://ctf.example.edu/api/v1/auth/cas/callback",
		AutoProvision: false,
	}, &mockRepository{
		findByUsernameFn: func(context.Context, string) (*model.User, error) {
			return nil, identitycontracts.ErrUserNotFound
		},
	}, &mockTokenService{}, zap.NewNop(), &mockCASTicketValidator{
		validateTicketFn: func(context.Context, string) (*authports.CASPrincipal, error) {
			return &authports.CASPrincipal{Username: "cas_user_3"}, nil
		},
	})

	_, _, err := service.Authenticate(context.Background(), "ST-3")
	if err != errcode.ErrCASUserNotProvisioned {
		t.Fatalf("expected ErrCASUserNotProvisioned, got %v", err)
	}
}

func TestCASServiceAuthenticateRejectsInvalidTicket(t *testing.T) {
	t.Parallel()

	service := NewCASService(config.CASConfig{
		Enabled:    true,
		BaseURL:    "https://cas.example.edu/cas",
		ServiceURL: "https://ctf.example.edu/api/v1/auth/cas/callback",
	}, &mockRepository{}, &mockTokenService{}, zap.NewNop(), &mockCASTicketValidator{
		validateTicketFn: func(context.Context, string) (*authports.CASPrincipal, error) {
			return nil, authports.ErrCASTicketInvalid
		},
	})

	_, _, err := service.Authenticate(context.Background(), "ST-invalid")
	if err != errcode.ErrCASTicketInvalid {
		t.Fatalf("expected ErrCASTicketInvalid, got %v", err)
	}
}

type mockCASTicketValidator struct {
	validateTicketFn func(ctx context.Context, validateURL string) (*authports.CASPrincipal, error)
}

func (m *mockCASTicketValidator) ValidateTicket(ctx context.Context, validateURL string) (*authports.CASPrincipal, error) {
	if m != nil && m.validateTicketFn != nil {
		return m.validateTicketFn(ctx, validateURL)
	}
	return nil, nil
}
