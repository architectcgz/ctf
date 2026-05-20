package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

type mockProfileRepository struct {
	findByIDFn       func(ctx context.Context, userID int64) (*identitycontracts.User, error)
	updatePasswordFn func(ctx context.Context, userID int64, newHash string) error
}

func (m *mockProfileRepository) List(context.Context, identitycontracts.UserListFilter) ([]*identitycontracts.User, int64, error) {
	return nil, 0, nil
}

func (m *mockProfileRepository) FindByID(ctx context.Context, userID int64) (*identitycontracts.User, error) {
	if m.findByIDFn == nil {
		return nil, identitycontracts.ErrUserNotFound
	}
	return m.findByIDFn(ctx, userID)
}

func (m *mockProfileRepository) FindByUsername(context.Context, string) (*identitycontracts.User, error) {
	return nil, identitycontracts.ErrUserNotFound
}

func (m *mockProfileRepository) Create(context.Context, *identitycontracts.User) error {
	return nil
}

func (m *mockProfileRepository) Update(context.Context, *identitycontracts.User) error {
	return nil
}

func (m *mockProfileRepository) Delete(context.Context, int64) error {
	return nil
}

func (m *mockProfileRepository) UpdatePassword(ctx context.Context, userID int64, newHash string) error {
	if m.updatePasswordFn == nil {
		return nil
	}
	return m.updatePasswordFn(ctx, userID, newHash)
}

func (m *mockProfileRepository) UpdateLoginState(context.Context, int64, int, *time.Time, *time.Time, string) error {
	return nil
}

func (m *mockProfileRepository) UpdateProfile(context.Context, *identitycontracts.User) error {
	return nil
}

func TestProfileServiceChangePasswordSuccess(t *testing.T) {
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
	oldHash := user.PasswordHash

	var updatedHash string
	service := NewProfileService(&mockProfileRepository{
		findByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
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
	}, zap.NewNop())

	err := service.ChangePassword(context.Background(), user.ID, identitycontracts.ChangePasswordInput{
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

func TestProfileServiceChangePasswordOldPasswordInvalid(t *testing.T) {
	t.Parallel()

	user := &identitycontracts.User{ID: 1, Username: "alice_1"}
	if err := user.SetPassword("Password123"); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	service := NewProfileService(&mockProfileRepository{
		findByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			return user, nil
		},
		updatePasswordFn: func(ctx context.Context, userID int64, newHash string) error {
			t.Fatalf("UpdatePassword() should not be called")
			return nil
		},
	}, zap.NewNop())

	err := service.ChangePassword(context.Background(), user.ID, identitycontracts.ChangePasswordInput{
		OldPassword: "wrong-password",
		NewPassword: "Password456",
	})
	if !errors.Is(err, identitycontracts.ErrInvalidOldPassword) {
		t.Fatalf("expected old password invalid, got %v", err)
	}
}

func TestProfileServiceChangePasswordRejectsSamePassword(t *testing.T) {
	t.Parallel()

	user := &identitycontracts.User{ID: 1, Username: "alice_1"}
	if err := user.SetPassword("Password123"); err != nil {
		t.Fatalf("SetPassword() error = %v", err)
	}

	service := NewProfileService(&mockProfileRepository{
		findByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			return user, nil
		},
		updatePasswordFn: func(ctx context.Context, userID int64, newHash string) error {
			t.Fatalf("UpdatePassword() should not be called")
			return nil
		},
	}, zap.NewNop())

	err := service.ChangePassword(context.Background(), user.ID, identitycontracts.ChangePasswordInput{
		OldPassword: "Password123",
		NewPassword: "Password123",
	})
	if !errors.Is(err, identitycontracts.ErrPasswordReuse) {
		t.Fatalf("expected password unchanged error, got %v", err)
	}
}
