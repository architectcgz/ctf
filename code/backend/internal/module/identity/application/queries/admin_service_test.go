package queries

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
)

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

	service := NewAdminService(identityinfra.NewRepository(db), config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, zap.NewNop())
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
