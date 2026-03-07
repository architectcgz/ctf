package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"ctf-platform/internal/authctx"
	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
	jwtpkg "ctf-platform/pkg/jwt"
)

type mockRepository struct {
	createFn         func(ctx context.Context, user *model.User) error
	findByUsernameFn func(ctx context.Context, username string) (*model.User, error)
	findByIDFn       func(ctx context.Context, userID int64) (*model.User, error)
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
	return nil
}

type mockTokenService struct {
	issueFn func(userID int64, username, role string) (*TokenPair, error)
}

func (m *mockTokenService) IssueTokens(userID int64, username, role string) (*TokenPair, error) {
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

	service := NewService(repo, tokenService, zap.NewNop())
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
	}, zap.NewNop())

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
	}, &mockTokenService{
		issueFn: func(userID int64, username, role string) (*TokenPair, error) {
			return nil, errors.New("should not be called")
		},
	}, zap.NewNop())

	_, _, err := service.Login(context.Background(), &dto.LoginReq{
		Username: "alice_1",
		Password: "wrong-password",
	})
	if !errors.Is(err, errcode.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}
