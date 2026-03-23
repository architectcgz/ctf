package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	identitymodule "ctf-platform/internal/module/identity"
	"ctf-platform/internal/module/identity/infrastructure"
	"ctf-platform/pkg/errcode"
)

func setupIdentityTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Role{}, &model.User{}, &model.UserRole{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	roles := []model.Role{
		{ID: 1, Code: model.RoleStudent, Name: "Student"},
		{ID: 2, Code: model.RoleTeacher, Name: "Teacher"},
		{ID: 3, Code: model.RoleAdmin, Name: "Admin"},
	}
	for _, role := range roles {
		if err := db.Create(&role).Error; err != nil {
			t.Fatalf("seed role: %v", err)
		}
	}

	return db
}

func newAdminServiceForTest(db *gorm.DB) identitymodule.AdminService {
	return NewAdminService(infrastructure.NewRepository(db), config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, zap.NewNop())
}

func TestAdminServiceCreateUserStoresIdentityNumbersByRole(t *testing.T) {
	db := setupIdentityTestDB(t)
	service := newAdminServiceForTest(db)

	resp, err := service.CreateUser(context.Background(), &dto.CreateAdminUserReq{
		Username:  "student-1",
		Name:      "Alice",
		Password:  "Password123",
		Role:      model.RoleStudent,
		StudentNo: "20240001",
		TeacherNo: "T-ignored",
		Status:    model.UserStatusActive,
	})
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if resp.StudentNo == nil || *resp.StudentNo != "20240001" {
		t.Fatalf("expected student no in response, got %+v", resp)
	}
	if resp.Name == nil || *resp.Name != "Alice" {
		t.Fatalf("expected name in response, got %+v", resp)
	}
	if resp.TeacherNo != nil {
		t.Fatalf("expected teacher no to be cleared for student, got %+v", resp)
	}

	var user model.User
	if err := db.First(&user, resp.ID).Error; err != nil {
		t.Fatalf("load created user: %v", err)
	}
	if user.Name != "Alice" || user.StudentNo != "20240001" || user.TeacherNo != "" {
		t.Fatalf("unexpected stored identity numbers: %+v", user)
	}
}

func TestAdminServiceCreateUserRejectsDuplicateUsername(t *testing.T) {
	db := setupIdentityTestDB(t)
	service := newAdminServiceForTest(db)

	if _, err := service.CreateUser(context.Background(), &dto.CreateAdminUserReq{
		Username: "duplicate-user",
		Password: "Password123",
		Role:     model.RoleStudent,
		Status:   model.UserStatusActive,
	}); err != nil {
		t.Fatalf("seed CreateUser() error = %v", err)
	}

	_, err := service.CreateUser(context.Background(), &dto.CreateAdminUserReq{
		Username: "duplicate-user",
		Password: "Password123",
		Role:     model.RoleStudent,
		Status:   model.UserStatusActive,
	})
	if !errors.Is(err, errcode.ErrUsernameExists) {
		t.Fatalf("expected ErrUsernameExists, got %v", err)
	}
}

func TestAdminServiceListUsersFiltersByIdentityNumber(t *testing.T) {
	db := setupIdentityTestDB(t)
	now := time.Now()
	users := []model.User{
		{
			ID:           1,
			Username:     "student-1",
			Name:         "Alice",
			PasswordHash: "hash",
			StudentNo:    "20240001",
			Role:         model.RoleStudent,
			Status:       model.UserStatusActive,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           2,
			Username:     "teacher-1",
			Name:         "Bob",
			PasswordHash: "hash",
			TeacherNo:    "T-1001",
			Role:         model.RoleTeacher,
			Status:       model.UserStatusActive,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}

	service := newAdminServiceForTest(db)
	list, total, _, _, err := service.ListUsers(context.Background(), &dto.AdminUserQuery{
		StudentNo: "20240001",
	})
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected one user, got total=%d list=%+v", total, list)
	}
	if list[0].StudentNo == nil || *list[0].StudentNo != "20240001" {
		t.Fatalf("expected student no in response, got %+v", list[0])
	}
	if list[0].Name == nil || *list[0].Name != "Alice" {
		t.Fatalf("expected name in response, got %+v", list[0])
	}
	if list[0].TeacherNo != nil {
		t.Fatalf("expected teacher no to be empty, got %+v", list[0])
	}
}
