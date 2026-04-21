package testsupport

import (
	"ctf-platform/internal/model"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.Challenge{},
		&model.ChallengePackageRevision{},
		&model.AWDServiceTemplate{},
		&model.ChallengePublishCheckJob{},
		&model.Image{},
		&model.Instance{},
		&model.Submission{},
		&model.ChallengeHint{},
		&model.ChallengeWriteup{},
		&model.SubmissionWriteup{},
		&model.ChallengeTopology{},
		&model.EnvironmentTemplate{},
	); err != nil {
		t.Fatalf("failed to migrate tables: %v", err)
	}
	return db
}

func SetupTagTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Challenge{}, &model.Tag{}, &model.ChallengeTag{}); err != nil {
		t.Fatalf("failed to migrate tag tables: %v", err)
	}
	return db
}
