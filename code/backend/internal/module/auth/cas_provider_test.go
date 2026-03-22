package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	"ctf-platform/pkg/errcode"
)

func TestCASProviderAuthenticateAutoProvisionSuccess(t *testing.T) {
	t.Parallel()

	casServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("ticket"); got != "ST-1" {
			t.Fatalf("unexpected ticket: %s", got)
		}
		if got := r.URL.Query().Get("service"); got != "https://ctf.example.edu/api/v1/auth/cas/callback" {
			t.Fatalf("unexpected service: %s", got)
		}
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
  <cas:authenticationSuccess>
    <cas:user>cas_user_1</cas:user>
    <cas:attributes>
      <cas:displayName>CAS User</cas:displayName>
      <cas:mail>cas_user_1@example.edu</cas:mail>
      <cas:className>CTF-1</cas:className>
      <cas:studentNo>20260001</cas:studentNo>
    </cas:attributes>
  </cas:authenticationSuccess>
</cas:serviceResponse>`)
	}))
	defer casServer.Close()

	repo := &mockRepository{
		findByUsernameFn: func(ctx context.Context, username string) (*model.User, error) {
			return nil, ErrUserNotFound
		},
		createFn: func(ctx context.Context, user *model.User) error {
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
		issueFn: func(userID int64, username, role string) (*authcontracts.TokenPair, error) {
			if userID != 101 || username != "cas_user_1" || role != model.RoleStudent {
				t.Fatalf("unexpected token issue params: %d %s %s", userID, username, role)
			}
			return &authcontracts.TokenPair{
				AccessToken:     "cas-access-token",
				RefreshToken:    "cas-refresh-token",
				AccessTokenTTL:  15 * time.Minute,
				RefreshTokenTTL: 24 * time.Hour,
			}, nil
		},
	}

	provider := NewCASProvider(config.CASConfig{
		Enabled:       true,
		BaseURL:       casServer.URL,
		ServiceURL:    "https://ctf.example.edu/api/v1/auth/cas/callback",
		AutoProvision: true,
	}, repo, tokenService, zap.NewNop(), casServer.Client())

	resp, tokens, err := provider.Authenticate(context.Background(), "ST-1")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if resp.User.Username != "cas_user_1" || resp.User.Role != model.RoleStudent {
		t.Fatalf("unexpected login response user: %+v", resp.User)
	}
	if tokens.RefreshToken != "cas-refresh-token" {
		t.Fatalf("unexpected refresh token: %s", tokens.RefreshToken)
	}
}

func TestCASProviderAuthenticateExistingUserSyncsProfileAndUnlocksExpired(t *testing.T) {
	t.Parallel()

	casServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<serviceResponse>
  <authenticationSuccess>
    <user>cas_user_2</user>
    <attributes>
      <name>Updated Name</name>
      <email>cas_user_2@example.edu</email>
      <class_name>CTF-2</class_name>
      <student_no>20260002</student_no>
    </attributes>
  </authenticationSuccess>
</serviceResponse>`)
	}))
	defer casServer.Close()

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
		findByUsernameFn: func(ctx context.Context, username string) (*model.User, error) {
			return user, nil
		},
		updateCASProfileFn: func(ctx context.Context, updated *model.User) error {
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
		issueFn: func(userID int64, username, role string) (*authcontracts.TokenPair, error) {
			return &authcontracts.TokenPair{
				AccessToken:     "existing-access-token",
				RefreshToken:    "existing-refresh-token",
				AccessTokenTTL:  15 * time.Minute,
				RefreshTokenTTL: 24 * time.Hour,
			}, nil
		},
	}

	provider := NewCASProvider(config.CASConfig{
		Enabled:    true,
		BaseURL:    casServer.URL,
		ServiceURL: "https://ctf.example.edu/api/v1/auth/cas/callback",
	}, repo, tokenService, zap.NewNop(), casServer.Client())

	resp, _, err := provider.Authenticate(context.Background(), "ST-2")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if resp.User.Username != "cas_user_2" {
		t.Fatalf("unexpected username: %+v", resp.User)
	}
}

func TestCASProviderAuthenticateRejectsUserWhenAutoProvisionDisabled(t *testing.T) {
	t.Parallel()

	casServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<serviceResponse>
  <authenticationSuccess>
    <user>cas_user_3</user>
  </authenticationSuccess>
</serviceResponse>`)
	}))
	defer casServer.Close()

	provider := NewCASProvider(config.CASConfig{
		Enabled:       true,
		BaseURL:       casServer.URL,
		ServiceURL:    "https://ctf.example.edu/api/v1/auth/cas/callback",
		AutoProvision: false,
	}, &mockRepository{
		findByUsernameFn: func(ctx context.Context, username string) (*model.User, error) {
			return nil, ErrUserNotFound
		},
	}, &mockTokenService{}, zap.NewNop(), casServer.Client())

	_, _, err := provider.Authenticate(context.Background(), "ST-3")
	if err != errcode.ErrCASUserNotProvisioned {
		t.Fatalf("expected ErrCASUserNotProvisioned, got %v", err)
	}
}

func TestCASProviderAuthenticateRejectsInvalidTicket(t *testing.T) {
	t.Parallel()

	casServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<serviceResponse>
  <authenticationFailure code="INVALID_TICKET">ticket not recognized</authenticationFailure>
</serviceResponse>`)
	}))
	defer casServer.Close()

	provider := NewCASProvider(config.CASConfig{
		Enabled:    true,
		BaseURL:    casServer.URL,
		ServiceURL: "https://ctf.example.edu/api/v1/auth/cas/callback",
	}, &mockRepository{}, &mockTokenService{}, zap.NewNop(), casServer.Client())

	_, _, err := provider.Authenticate(context.Background(), "ST-invalid")
	if err != errcode.ErrCASTicketInvalid {
		t.Fatalf("expected ErrCASTicketInvalid, got %v", err)
	}
}
