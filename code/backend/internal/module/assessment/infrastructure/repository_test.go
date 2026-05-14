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

func setupAssessmentRepoTestDB(t *testing.T) *gorm.DB {
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
		t.Fatalf("migrate assessment repo tables: %v", err)
	}
	return db
}

func findAssessmentSnapshotDimension(t *testing.T, snapshot *teachingadvice.StudentFactSnapshot, dimension string) teachingadvice.DimensionFact {
	t.Helper()

	if snapshot == nil {
		t.Fatal("snapshot is nil")
	}
	for _, item := range snapshot.Dimensions {
		if item.Dimension == dimension {
			return item
		}
	}
	t.Fatalf("dimension %s not found in snapshot %+v", dimension, snapshot.Dimensions)
	return teachingadvice.DimensionFact{}
}

func TestRepositoryGetStudentTeachingFactSnapshotBackfillsAWDSuccessDimensionFacts(t *testing.T) {
	db := setupAssessmentRepoTestDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC()

	if err := db.Create(&model.User{
		ID:        7,
		Username:  "alice",
		Role:      model.RoleStudent,
		ClassName: "Class A",
		Status:    model.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&model.SkillProfile{
		UserID:    7,
		Dimension: model.DimensionWeb,
		Score:     0.2,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed web profile: %v", err)
	}

	awdChallenges := []model.AWDChallenge{
		{ID: 701, Name: "web-awd-easy-a", Category: model.DimensionWeb, Difficulty: model.ChallengeDifficultyEasy, Status: model.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 702, Name: "web-awd-easy-b", Category: model.DimensionWeb, Difficulty: model.ChallengeDifficultyEasy, Status: model.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 703, Name: "web-awd-medium-a", Category: model.DimensionWeb, Difficulty: model.ChallengeDifficultyMedium, Status: model.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 704, Name: "web-awd-medium-b", Category: model.DimensionWeb, Difficulty: model.ChallengeDifficultyMedium, Status: model.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
	}
	for _, challenge := range awdChallenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed awd challenge %s: %v", challenge.Name, err)
		}
	}

	logs := []model.AWDAttackLog{
		{ID: 1, RoundID: 81, AttackerTeamID: 91, VictimTeamID: 101, ServiceID: 111, AWDChallengeID: 701, AttackType: model.AWDAttackTypeFlagCapture, Source: model.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 80, SubmittedByUserID: ptrAssessmentRepoInt64(7), CreatedAt: now},
		{ID: 2, RoundID: 81, AttackerTeamID: 91, VictimTeamID: 102, ServiceID: 112, AWDChallengeID: 702, AttackType: model.AWDAttackTypeFlagCapture, Source: model.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 80, SubmittedByUserID: ptrAssessmentRepoInt64(7), CreatedAt: now.Add(1 * time.Minute)},
		{ID: 3, RoundID: 82, AttackerTeamID: 91, VictimTeamID: 103, ServiceID: 113, AWDChallengeID: 703, AttackType: model.AWDAttackTypeFlagCapture, Source: model.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 90, SubmittedByUserID: ptrAssessmentRepoInt64(7), CreatedAt: now.Add(2 * time.Minute)},
		{ID: 4, RoundID: 82, AttackerTeamID: 91, VictimTeamID: 104, ServiceID: 114, AWDChallengeID: 704, AttackType: model.AWDAttackTypeFlagCapture, Source: model.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 90, SubmittedByUserID: ptrAssessmentRepoInt64(7), CreatedAt: now.Add(3 * time.Minute)},
		{ID: 5, RoundID: 83, AttackerTeamID: 91, VictimTeamID: 105, ServiceID: 115, AWDChallengeID: 704, AttackType: model.AWDAttackTypeFlagCapture, Source: model.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 0, SubmittedByUserID: ptrAssessmentRepoInt64(7), CreatedAt: now.Add(4 * time.Minute)},
		{ID: 6, RoundID: 83, AttackerTeamID: 91, VictimTeamID: 106, ServiceID: 116, AWDChallengeID: 704, AttackType: model.AWDAttackTypeFlagCapture, Source: model.AWDAttackSourceManual, IsSuccess: true, ScoreGained: 100, SubmittedByUserID: ptrAssessmentRepoInt64(7), CreatedAt: now.Add(5 * time.Minute)},
	}
	for _, log := range logs {
		if err := db.Create(&log).Error; err != nil {
			t.Fatalf("seed awd attack log %d: %v", log.ID, err)
		}
	}

	snapshot, err := repo.GetStudentTeachingFactSnapshot(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetStudentTeachingFactSnapshot() error = %v", err)
	}
	if snapshot.AWDSuccessCount != 4 {
		t.Fatalf("expected awd success count to ignore zero-score/manual logs, got %+v", snapshot)
	}

	web := findAssessmentSnapshotDimension(t, snapshot, model.DimensionWeb)
	if web.ProfileScore != 1 {
		t.Fatalf("expected web profile score lifted by awd coverage, got %+v", web)
	}
	if web.SuccessCount != 4 {
		t.Fatalf("expected web success count to include awd success coverage, got %+v", web)
	}
	if web.EvidenceCount != 4 {
		t.Fatalf("expected web evidence count to include awd success coverage, got %+v", web)
	}
	if web.SolvedDifficultyCounts[model.ChallengeDifficultyEasy] != 2 || web.SolvedDifficultyCounts[model.ChallengeDifficultyMedium] != 2 {
		t.Fatalf("expected awd difficulty coverage merged into snapshot, got %+v", web.SolvedDifficultyCounts)
	}
}

func ptrAssessmentRepoInt64(value int64) *int64 {
	return &value
}
