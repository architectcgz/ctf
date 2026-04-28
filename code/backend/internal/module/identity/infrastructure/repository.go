package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

const (
	uniqueUsernameConstraint = "uk_users_username"
	uniqueEmailConstraint    = "uk_users_email"
	uniqueStudentNoIndex     = "uk_users_student_no"
	uniqueTeacherNoIndex     = "uk_users_teacher_no"
)

type Repository struct {
	db *gorm.DB
}

var _ identitycontracts.UserRepository = (*Repository)(nil)

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, filter identitycontracts.UserListFilter) ([]*model.User, int64, error) {
	query := r.dbWithContext(ctx).Model(&model.User{}).Where("deleted_at IS NULL")
	if filter.Keyword != "" {
		keyword := "%" + strings.TrimSpace(filter.Keyword) + "%"
		query = query.Where(
			"(username LIKE ? OR name LIKE ? OR email LIKE ? OR class_name LIKE ? OR student_no LIKE ? OR teacher_no LIKE ?)",
			keyword,
			keyword,
			keyword,
			keyword,
			keyword,
			keyword,
		)
	}
	if filter.StudentNo != "" {
		query = query.Where("student_no = ?", strings.TrimSpace(filter.StudentNo))
	}
	if filter.TeacherNo != "" {
		query = query.Where("teacher_no = ?", strings.TrimSpace(filter.TeacherNo))
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
	if err := r.dbWithContext(ctx).Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, identitycontracts.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := r.dbWithContext(ctx).Where("username = ? AND deleted_at IS NULL", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, identitycontracts.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by username: %w", err)
	}
	return &user, nil
}

func (r *Repository) Create(ctx context.Context, user *model.User) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.User{}).
			Where("id = ? AND deleted_at IS NULL", user.ID).
			Updates(map[string]any{
				"password_hash": user.PasswordHash,
				"name":          user.Name,
				"email":         user.Email,
				"student_no":    user.StudentNo,
				"teacher_no":    user.TeacherNo,
				"role":          user.Role,
				"class_name":    user.ClassName,
				"status":        user.Status,
				"updated_at":    time.Now(),
			})
		if result.Error != nil {
			return mapUserWriteError(result.Error)
		}
		if result.RowsAffected == 0 {
			return identitycontracts.ErrUserNotFound
		}
		if err := syncUserRole(tx, user.ID, user.Role); err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) Delete(ctx context.Context, userID int64) error {
	result := r.dbWithContext(ctx).Where("id = ? AND deleted_at IS NULL", userID).Delete(&model.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return identitycontracts.ErrUserNotFound
	}
	return nil
}

func (r *Repository) UpdatePassword(ctx context.Context, userID int64, newHash string) error {
	result := r.dbWithContext(ctx).Model(&model.User{}).Where("id = ? AND deleted_at IS NULL", userID).Update("password_hash", newHash)
	if result.Error != nil {
		return fmt.Errorf("update password: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return identitycontracts.ErrUserNotFound
	}
	return nil
}

func (r *Repository) UpdateLoginState(ctx context.Context, userID int64, failedAttempts int, lastFailedAt, lockedUntil *time.Time, status string) error {
	updates := map[string]any{
		"failed_login_attempts": failedAttempts,
		"last_failed_login_at":  lastFailedAt,
		"locked_until":          lockedUntil,
		"status":                status,
	}
	result := r.dbWithContext(ctx).Model(&model.User{}).Where("id = ? AND deleted_at IS NULL", userID).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("update login state: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return identitycontracts.ErrUserNotFound
	}
	return nil
}

func (r *Repository) UpdateProfile(ctx context.Context, user *model.User) error {
	updates := map[string]any{
		"name":                  user.Name,
		"email":                 user.Email,
		"student_no":            user.StudentNo,
		"teacher_no":            user.TeacherNo,
		"class_name":            user.ClassName,
		"status":                user.Status,
		"failed_login_attempts": user.FailedLoginAttempts,
		"last_failed_login_at":  user.LastFailedLoginAt,
		"locked_until":          user.LockedUntil,
		"updated_at":            time.Now(),
	}
	result := r.dbWithContext(ctx).Model(&model.User{}).Where("id = ? AND deleted_at IS NULL", user.ID).Updates(updates)
	if result.Error != nil {
		return mapUserWriteError(fmt.Errorf("update profile: %w", result.Error))
	}
	if result.RowsAffected == 0 {
		return identitycontracts.ErrUserNotFound
	}
	return nil
}

func (r *Repository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func syncUserRole(tx *gorm.DB, userID int64, roleCode string) error {
	var role model.Role
	if err := tx.Where("code = ?", roleCode).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return identitycontracts.ErrRoleNotFound
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
		return identitycontracts.ErrUsernameExists
	case strings.Contains(message, uniqueEmailConstraint):
		return identitycontracts.ErrEmailExists
	case strings.Contains(message, uniqueStudentNoIndex):
		return identitycontracts.ErrStudentNoExists
	case strings.Contains(message, uniqueTeacherNoIndex):
		return identitycontracts.ErrTeacherNoExists
	default:
		return err
	}
}
