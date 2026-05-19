package testsupport

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
	practiceentity "ctf-platform/internal/module/practice/entity"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
)

func SetupPracticeTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&practiceImageRecord{}); err != nil {
		t.Fatalf("migrate image: %v", err)
	}
	return db
}

func SetupScoreServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&practiceentity.Challenge{}, &contestentity.Submission{}, &identitycontracts.User{}, &practiceentity.UserScore{}); err != nil {
		t.Fatalf("migrate score tables: %v", err)
	}
	return db
}
