package commands

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
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

func newAdminServiceForTest(db *gorm.DB) *AdminService {
	return NewAdminService(identityinfra.NewRepository(db), zap.NewNop())
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
