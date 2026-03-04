package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/model"
)

func TestRepositoryCreateReturnsRoleNotFound(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	user := &model.User{
		Username:     "alice_1",
		PasswordHash: "hashed",
		Role:         model.RoleStudent,
		Status:       model.UserStatusActive,
	}

	err := repo.Create(context.Background(), user)
	if !errors.Is(err, ErrRoleNotFound) {
		t.Fatalf("expected ErrRoleNotFound, got %v", err)
	}
}

func TestRepositoryCreatePersistsUserAndUserRole(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	db := repo.(*repository).db
	seedRole(t, db, model.RoleStudent)

	user := &model.User{
		Username:     "alice_1",
		PasswordHash: "hashed",
		Role:         model.RoleStudent,
		Status:       model.UserStatusActive,
	}

	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	var persistedUser model.User
	if err := db.First(&persistedUser, user.ID).Error; err != nil {
		t.Fatalf("query user: %v", err)
	}

	var persistedUserRole model.UserRole
	if err := db.Where("user_id = ?", user.ID).First(&persistedUserRole).Error; err != nil {
		t.Fatalf("query user role: %v", err)
	}
	if persistedUserRole.UserID != user.ID {
		t.Fatalf("unexpected user role binding: %+v", persistedUserRole)
	}
}

func TestRepositoryFindByUsernameReturnsUserNotFound(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	_, err := repo.FindByUsername(context.Background(), "missing_user")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func newTestRepository(t *testing.T) Repository {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	schemaStatements := []string{
		`CREATE TABLE roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			email TEXT,
			role TEXT NOT NULL DEFAULT 'student',
			class_name TEXT,
			status TEXT NOT NULL DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		);`,
		`CREATE UNIQUE INDEX uk_users_username ON users(username);`,
		`CREATE UNIQUE INDEX uk_users_email ON users(email);`,
		`CREATE TABLE user_roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			role_id INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
	}
	for _, statement := range schemaStatements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("exec schema statement failed: %v", err)
		}
	}

	return NewRepository(db)
}

func seedRole(t *testing.T, db *gorm.DB, code string) {
	t.Helper()

	role := &model.Role{
		Code: code,
		Name: code,
	}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
}
