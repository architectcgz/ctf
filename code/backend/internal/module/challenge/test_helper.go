package challenge

import (
	"ctf-platform/internal/model"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	db.AutoMigrate(&model.Challenge{}, &model.Image{}, &model.Instance{})
	return db
}

func setupTagTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Challenge{}, &model.Tag{}, &model.ChallengeTag{}); err != nil {
		t.Fatalf("failed to migrate tag tables: %v", err)
	}
	return db
}
