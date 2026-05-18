package queries

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

func setupIdentityTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&identitycontracts.Role{}, &identitycontracts.User{}, &identitycontracts.UserRole{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}

	roles := []identitycontracts.Role{
		{ID: 1, Code: identitycontracts.RoleStudent, Name: "Student"},
		{ID: 2, Code: identitycontracts.RoleTeacher, Name: "Teacher"},
		{ID: 3, Code: identitycontracts.RoleAdmin, Name: "Admin"},
	}
	for _, role := range roles {
		if err := db.Create(&role).Error; err != nil {
			t.Fatalf("seed role: %v", err)
		}
	}

	return db
}
