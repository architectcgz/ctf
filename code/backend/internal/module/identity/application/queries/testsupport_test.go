package queries

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
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
