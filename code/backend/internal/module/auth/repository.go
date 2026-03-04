package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

const (
	uniqueUsernameConstraint = "uk_users_username"
	uniqueEmailConstraint    = "uk_users_email"
)

type Repository interface {
	Create(ctx context.Context, user *model.User) error
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByID(ctx context.Context, userID int64) (*model.User, error)
	UpdatePassword(ctx context.Context, userID int64, newHash string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return mapCreateError(err)
		}

		role := &model.Role{}
		if err := tx.Where("code = ?", user.Role).First(role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRoleNotFound
			}
			return fmt.Errorf("find role: %w", err)
		}

		userRole := &model.UserRole{
			UserID: user.ID,
			RoleID: role.ID,
		}
		if err := tx.Create(userRole).Error; err != nil {
			return fmt.Errorf("create user role: %w", err)
		}

		return nil
	})
}

func (r *repository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	user := &model.User{}
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by username: %w", err)
	}
	return user, nil
}

func (r *repository) FindByID(ctx context.Context, userID int64) (*model.User, error) {
	user := &model.User{}
	if err := r.db.WithContext(ctx).First(user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

func (r *repository) UpdatePassword(ctx context.Context, userID int64, newHash string) error {
	result := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("password_hash", newHash)
	if result.Error != nil {
		return fmt.Errorf("update password: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func mapCreateError(err error) error {
	message := err.Error()
	switch {
	case containsConstraint(message, uniqueUsernameConstraint):
		return ErrUsernameExists
	case containsConstraint(message, uniqueEmailConstraint):
		return ErrEmailExists
	default:
		return fmt.Errorf("create user: %w", err)
	}
}

func containsConstraint(message, constraint string) bool {
	return message != "" && constraint != "" && strings.Contains(message, constraint)
}
