package infrastructure

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	teachingadvice "ctf-platform/internal/teaching/advice"
)

func setupTeachingQueryRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.Submission{},
		&model.SkillProfile{},
		&model.AWDChallenge{},
		&model.AWDAttackLog{},
	); err != nil {
		t.Fatalf("migrate teaching query repo tables: %v", err)
	}
	return db
}

func findSnapshotDimension(t *testing.T, snapshot teachingadvice.StudentFactSnapshot, dimension string) teachingadvice.DimensionFact {
	t.Helper()

	for _, item := range snapshot.Dimensions {
		if item.Dimension == dimension {
			return item
		}
	}
	t.Fatalf("dimension %s not found in snapshot %+v", dimension, snapshot.Dimensions)
	return teachingadvice.DimensionFact{}
}

func TestRepositoryListClassTeachingFactSnapshotsBackfillsAWDSuccessDimensionFacts(t *testing.T) {
	db := setupTeachingQueryRepoTestDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC()

	students := []model.User{
		{ID: 1, Username: "alice", Role: model.RoleStudent, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now},
		{ID: 2, Username: "bob", Role: model.RoleStudent, ClassName: "Class B", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now},
	}
	for _, user := range students {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("seed user %s: %v", user.Username, err)
		}
	}
	if err := db.Create(&model.SkillProfile{
		UserID:    1,
		Dimension: model.DimensionPwn,
		Score:     0.18,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed pwn profile: %v", err)
	}

	awdChallenges := []model.AWDChallenge{
		{ID: 1001, Name: "pwn-awd-easy-a", Category: model.DimensionPwn, Difficulty: model.ChallengeDifficultyEasy, Status: model.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 1002, Name: "pwn-awd-easy-b", Category: model.DimensionPwn, Difficulty: model.ChallengeDifficultyEasy, Status: model.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 1003, Name: "pwn-awd-medium-a", Category: model.DimensionPwn, Difficulty: model.ChallengeDifficultyMedium, Status: model.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 1004, Name: "pwn-awd-medium-b", Category: model.DimensionPwn, Difficulty: model.ChallengeDifficultyMedium, Status: model.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
	}
	for _, challenge := range awdChallenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed awd challenge %s: %v", challenge.Name, err)
		}
	}

	logs := []model.AWDAttackLog{
		{ID: 1, RoundID: 11, AttackerTeamID: 21, VictimTeamID: 31, ServiceID: 41, AWDChallengeID: 1001, AttackType: model.AWDAttackTypeFlagCapture, Source: model.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 80, SubmittedByUserID: ptrTeachingQueryInt64(1), CreatedAt: now},
		{ID: 2, RoundID: 11, AttackerTeamID: 21, VictimTeamID: 32, ServiceID: 42, AWDChallengeID: 1002, AttackType: model.AWDAttackTypeFlagCapture, Source: model.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 80, SubmittedByUserID: ptrTeachingQueryInt64(1), CreatedAt: now.Add(1 * time.Minute)},
		{ID: 3, RoundID: 12, AttackerTeamID: 21, VictimTeamID: 33, ServiceID: 43, AWDChallengeID: 1003, AttackType: model.AWDAttackTypeFlagCapture, Source: model.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 90, SubmittedByUserID: ptrTeachingQueryInt64(1), CreatedAt: now.Add(2 * time.Minute)},
		{ID: 4, RoundID: 12, AttackerTeamID: 21, VictimTeamID: 34, ServiceID: 44, AWDChallengeID: 1004, AttackType: model.AWDAttackTypeFlagCapture, Source: model.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 90, SubmittedByUserID: ptrTeachingQueryInt64(1), CreatedAt: now.Add(3 * time.Minute)},
		{ID: 5, RoundID: 13, AttackerTeamID: 21, VictimTeamID: 35, ServiceID: 45, AWDChallengeID: 1004, AttackType: model.AWDAttackTypeFlagCapture, Source: model.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 0, SubmittedByUserID: ptrTeachingQueryInt64(1), CreatedAt: now.Add(4 * time.Minute)},
		{ID: 6, RoundID: 13, AttackerTeamID: 21, VictimTeamID: 36, ServiceID: 46, AWDChallengeID: 1004, AttackType: model.AWDAttackTypeFlagCapture, Source: model.AWDAttackSourceManual, IsSuccess: true, ScoreGained: 100, SubmittedByUserID: ptrTeachingQueryInt64(1), CreatedAt: now.Add(5 * time.Minute)},
	}
	for _, log := range logs {
		if err := db.Create(&log).Error; err != nil {
			t.Fatalf("seed awd attack log %d: %v", log.ID, err)
		}
	}

	snapshots, err := repo.ListClassTeachingFactSnapshots(context.Background(), "Class A", now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("ListClassTeachingFactSnapshots() error = %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("expected only Class A student snapshots, got %+v", snapshots)
	}

	snapshot := snapshots[0]
	if snapshot.UserID != 1 {
		t.Fatalf("expected alice snapshot, got %+v", snapshot)
	}
	if snapshot.AWDSuccessCount != 4 {
		t.Fatalf("expected awd success count to ignore zero-score/manual logs, got %+v", snapshot)
	}

	pwn := findSnapshotDimension(t, snapshot, model.DimensionPwn)
	if pwn.ProfileScore != 1 {
		t.Fatalf("expected pwn profile score lifted by awd coverage, got %+v", pwn)
	}
	if pwn.SuccessCount != 4 {
		t.Fatalf("expected pwn success count to include awd success coverage, got %+v", pwn)
	}
	if pwn.EvidenceCount != 4 {
		t.Fatalf("expected pwn evidence count to include awd success coverage, got %+v", pwn)
	}
	if pwn.SolvedDifficultyCounts[model.ChallengeDifficultyEasy] != 2 || pwn.SolvedDifficultyCounts[model.ChallengeDifficultyMedium] != 2 {
		t.Fatalf("expected awd difficulty coverage merged into snapshot, got %+v", pwn.SolvedDifficultyCounts)
	}
}

func ptrTeachingQueryInt64(value int64) *int64 {
	return &value
}
