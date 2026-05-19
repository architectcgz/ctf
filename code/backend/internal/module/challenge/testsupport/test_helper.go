package testsupport

import (
	"testing"

	"ctf-platform/internal/model"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
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
		&identitycontracts.User{},
		&model.Challenge{},
		&challengeentity.ChallengePackageRevision{},
		&challengeentity.AWDChallenge{},
		&challengeentity.ChallengePublishCheckJob{},
		&challengeentity.Image{},
		&challengeentity.ImageBuildJob{},
		&instancecontracts.Instance{},
		&contestcontracts.Submission{},
		&challengeentity.ChallengeHint{},
		&challengeentity.ChallengeWriteup{},
		&challengeentity.SubmissionWriteup{},
		&challengeentity.ChallengeTopology{},
		&challengeentity.EnvironmentTemplate{},
		&contestcontracts.Contest{},
		&contestcontracts.ContestAWDService{},
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
	if err := db.AutoMigrate(&model.Challenge{}, &challengeentity.Tag{}, &challengeentity.ChallengeTag{}); err != nil {
		t.Fatalf("failed to migrate tag tables: %v", err)
	}
	return db
}
