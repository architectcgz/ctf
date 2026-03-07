package adminuser

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

const (
	uniqueUsernameConstraint = "uk_users_username"
	uniqueEmailConstraint    = "uk_users_email"
)

var ErrUserNotFound = errors.New("admin user not found")
var ErrUsernameExists = errors.New("admin username already exists")
var ErrEmailExists = errors.New("admin email already exists")
var ErrRoleNotFound = errors.New("admin role not found")

type UserListFilter struct {
	Keyword   string
	Role      string
	Status    string
	ClassName string
	Offset    int
	Limit     int
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, filter UserListFilter) ([]*model.User, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.User{}).Where("deleted_at IS NULL")
	if filter.Keyword != "" {
		keyword := "%" + strings.TrimSpace(filter.Keyword) + "%"
		query = query.Where("(username LIKE ? OR email LIKE ? OR class_name LIKE ?)", keyword, keyword, keyword)
	}
	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.ClassName != "" {
		query = query.Where("class_name = ?", strings.TrimSpace(filter.ClassName))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	users := make([]*model.User, 0)
	if err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *Repository) FindByID(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *Repository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return mapUserWriteError(err)
		}
		if err := syncUserRole(tx, user.ID, user.Role); err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).
			Where("id = ? AND deleted_at IS NULL", user.ID).
			Updates(map[string]any{
				"password_hash": user.PasswordHash,
				"email":         user.Email,
				"role":          user.Role,
				"class_name":    user.ClassName,
				"status":        user.Status,
				"updated_at":    time.Now(),
			}).Error; err != nil {
			return mapUserWriteError(err)
		}
		if err := syncUserRole(tx, user.ID, user.Role); err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) Delete(ctx context.Context, userID int64) error {
	result := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", userID).Delete(&model.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func syncUserRole(tx *gorm.DB, userID int64, roleCode string) error {
	var role model.Role
	if err := tx.Where("code = ?", roleCode).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}
		return fmt.Errorf("find role: %w", err)
	}
	if err := tx.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		return fmt.Errorf("delete user roles: %w", err)
	}
	if err := tx.Create(&model.UserRole{
		UserID: userID,
		RoleID: role.ID,
	}).Error; err != nil {
		return fmt.Errorf("create user role: %w", err)
	}
	return nil
}

func mapUserWriteError(err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, uniqueUsernameConstraint):
		return ErrUsernameExists
	case strings.Contains(message, uniqueEmailConstraint):
		return ErrEmailExists
	default:
		return err
	}
}
