package testsupport

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/model"
)

func SetupContestTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Contest{},
		&model.Challenge{},
		&model.AWDChallenge{},
		&model.ContestAWDService{},
		&model.User{},
		&model.Team{},
		&model.TeamMember{},
		&model.ContestRegistration{},
		&model.ContestAnnouncement{},
		&model.ContestChallenge{},
		&model.Submission{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uk_submissions_contest_user_challenge_correct
		ON submissions(contest_id, user_id, challenge_id)
		WHERE is_correct = 1 AND contest_id IS NOT NULL AND team_id IS NULL
	`).Error; err != nil {
		t.Fatalf("create contest submission user unique index: %v", err)
	}
	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS uk_submissions_contest_team_challenge_correct
		ON submissions(contest_id, team_id, challenge_id)
		WHERE is_correct = 1 AND contest_id IS NOT NULL AND team_id IS NOT NULL
	`).Error; err != nil {
		t.Fatalf("create contest submission team unique index: %v", err)
	}
	return db
}

func SetupAWDTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db := SetupContestTestDB(t)
	if err := db.AutoMigrate(
		&model.Instance{},
		&model.AWDRound{},
		&model.AWDTeamService{},
		&model.AWDAttackLog{},
		&model.AWDTrafficEvent{},
		&model.AWDServiceOperation{},
	); err != nil {
		t.Fatalf("auto migrate awd tables: %v", err)
	}
	return db
}
