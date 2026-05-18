package testsupport

import (
	"testing"

	"ctf-platform/internal/model"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
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
		&challengeentity.ChallengePackageRevision{},
		&model.AWDChallenge{},
		&challengeentity.ChallengePublishCheckJob{},
		&model.Image{},
		&challengeentity.ImageBuildJob{},
		&model.Instance{},
		&contestcontracts.Submission{},
		&challengeentity.ChallengeHint{},
		&model.ChallengeWriteup{},
		&challengeentity.SubmissionWriteup{},
		&model.ChallengeTopology{},
		&model.EnvironmentTemplate{},
		&model.Contest{},
		&model.ContestAWDService{},
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
