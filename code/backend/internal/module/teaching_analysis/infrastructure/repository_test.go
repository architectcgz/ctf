package infrastructure

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	assessmententity "ctf-platform/internal/module/assessment/entity"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/internal/shared/taxonomy"
	teachingadvice "ctf-platform/internal/teaching/advice"
)

type teachingAnalysisChallengeRepoRow struct {
	ID         int64          `gorm:"column:id;primaryKey"`
	Title      string         `gorm:"column:title"`
	Category   string         `gorm:"column:category"`
	Difficulty string         `gorm:"column:difficulty"`
	Points     int            `gorm:"column:points"`
	Status     string         `gorm:"column:status"`
	CreatedAt  time.Time      `gorm:"column:created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (teachingAnalysisChallengeRepoRow) TableName() string {
	return "challenges"
}

func setupTeachingAnalysisRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&identitycontracts.User{},
		&teachingAnalysisChallengeRepoRow{},
		&contestcontracts.Submission{},
		&assessmententity.SkillProfile{},
		&challengecontracts.AWDChallenge{},
		&contestcontracts.AWDAttackLog{},
	); err != nil {
		t.Fatalf("migrate teaching analysis repo tables: %v", err)
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

func TestClassInsightRepositoryListClassTeachingFactSnapshotsBackfillsAWDSuccessDimensionFacts(t *testing.T) {
	db := setupTeachingAnalysisRepoTestDB(t)
	repo := NewClassInsightRepository(db)
	now := time.Now().UTC()

	students := []identitycontracts.User{
		{ID: 1, Username: "alice", Role: identitycontracts.RoleStudent, ClassName: "Class A", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now},
		{ID: 2, Username: "bob", Role: identitycontracts.RoleStudent, ClassName: "Class B", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now},
	}
	for _, user := range students {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("seed user %s: %v", user.Username, err)
		}
	}
	if err := db.Create(&assessmententity.SkillProfile{
		UserID:    1,
		Dimension: taxonomy.DimensionPwn,
		Score:     0.18,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed pwn profile: %v", err)
	}

	awdChallenges := []challengecontracts.AWDChallenge{
		{ID: 1001, Name: "pwn-awd-easy-a", Category: taxonomy.DimensionPwn, Difficulty: taxonomy.DifficultyEasy, Status: challengecontracts.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 1002, Name: "pwn-awd-easy-b", Category: taxonomy.DimensionPwn, Difficulty: taxonomy.DifficultyEasy, Status: challengecontracts.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 1003, Name: "pwn-awd-medium-a", Category: taxonomy.DimensionPwn, Difficulty: taxonomy.DifficultyMedium, Status: challengecontracts.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 1004, Name: "pwn-awd-medium-b", Category: taxonomy.DimensionPwn, Difficulty: taxonomy.DifficultyMedium, Status: challengecontracts.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
	}
	for _, challenge := range awdChallenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed awd challenge %s: %v", challenge.Name, err)
		}
	}

	logs := []contestcontracts.AWDAttackLog{
		{ID: 1, RoundID: 11, AttackerTeamID: 21, VictimTeamID: 31, ServiceID: 41, AWDChallengeID: 1001, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 80, SubmittedByUserID: ptrTeachingAnalysisInt64(1), CreatedAt: now},
		{ID: 2, RoundID: 11, AttackerTeamID: 21, VictimTeamID: 32, ServiceID: 42, AWDChallengeID: 1002, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 80, SubmittedByUserID: ptrTeachingAnalysisInt64(1), CreatedAt: now.Add(1 * time.Minute)},
		{ID: 3, RoundID: 12, AttackerTeamID: 21, VictimTeamID: 33, ServiceID: 43, AWDChallengeID: 1003, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 90, SubmittedByUserID: ptrTeachingAnalysisInt64(1), CreatedAt: now.Add(2 * time.Minute)},
		{ID: 4, RoundID: 12, AttackerTeamID: 21, VictimTeamID: 34, ServiceID: 44, AWDChallengeID: 1004, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 90, SubmittedByUserID: ptrTeachingAnalysisInt64(1), CreatedAt: now.Add(3 * time.Minute)},
		{ID: 5, RoundID: 13, AttackerTeamID: 21, VictimTeamID: 35, ServiceID: 45, AWDChallengeID: 1004, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 0, SubmittedByUserID: ptrTeachingAnalysisInt64(1), CreatedAt: now.Add(4 * time.Minute)},
		{ID: 6, RoundID: 13, AttackerTeamID: 21, VictimTeamID: 36, ServiceID: 46, AWDChallengeID: 1004, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceManual, IsSuccess: true, ScoreGained: 100, SubmittedByUserID: ptrTeachingAnalysisInt64(1), CreatedAt: now.Add(5 * time.Minute)},
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

	pwn := findSnapshotDimension(t, snapshot, taxonomy.DimensionPwn)
	if pwn.ProfileScore != 1 {
		t.Fatalf("expected pwn profile score lifted by awd coverage, got %+v", pwn)
	}
	if pwn.SuccessCount != 4 {
		t.Fatalf("expected pwn success count to include awd success coverage, got %+v", pwn)
	}
	if pwn.AttemptCount != 5 {
		t.Fatalf("expected pwn attempt count to include awd submission attempts, got %+v", pwn)
	}
	if pwn.EvidenceCount != 5 {
		t.Fatalf("expected pwn evidence count to include awd submission attempts, got %+v", pwn)
	}
	if pwn.SolvedDifficultyCounts[taxonomy.DifficultyEasy] != 2 || pwn.SolvedDifficultyCounts[taxonomy.DifficultyMedium] != 2 {
		t.Fatalf("expected awd difficulty coverage merged into snapshot, got %+v", pwn.SolvedDifficultyCounts)
	}
}

func TestRepositoryListClassTeachingFactSnapshotsUsesUnifiedContestAndAWDSubmissionSemantics(t *testing.T) {
	db := setupTeachingAnalysisRepoTestDB(t)
	repo := NewClassInsightRepository(db)
	now := time.Now().UTC()
	contestID := int64(66)

	students := []identitycontracts.User{
		{ID: 11, Username: "alice", Role: identitycontracts.RoleStudent, ClassName: "Class A", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now},
		{ID: 12, Username: "bob", Role: identitycontracts.RoleStudent, ClassName: "Class B", Status: identitycontracts.UserStatusActive, CreatedAt: now, UpdatedAt: now},
	}
	for _, user := range students {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("seed user %s: %v", user.Username, err)
		}
	}
	if err := db.Create(&assessmententity.SkillProfile{
		UserID:    11,
		Dimension: taxonomy.DimensionWeb,
		Score:     0.15,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed web profile: %v", err)
	}

	challenges := []teachingAnalysisChallengeRepoRow{
		{ID: 801, Title: "web-practice", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyEasy, Points: 100, Status: challengecontracts.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 802, Title: "web-contest", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyMedium, Points: 120, Status: challengecontracts.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
	}
	for _, challenge := range challenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed challenge %s: %v", challenge.Title, err)
		}
	}

	submissions := []contestcontracts.Submission{
		{ID: 1, UserID: 11, ChallengeID: 801, IsCorrect: false, SubmittedAt: now, UpdatedAt: now},
		{ID: 2, UserID: 11, ChallengeID: 801, ContestID: &contestID, IsCorrect: true, ReviewStatus: contestcontracts.SubmissionReviewStatusApproved, SubmittedAt: now.Add(1 * time.Minute), UpdatedAt: now.Add(1 * time.Minute)},
		{ID: 3, UserID: 11, ChallengeID: 802, ContestID: &contestID, IsCorrect: true, SubmittedAt: now.Add(2 * time.Minute), UpdatedAt: now.Add(2 * time.Minute)},
	}
	for _, submission := range submissions {
		if err := db.Create(&submission).Error; err != nil {
			t.Fatalf("seed submission %d: %v", submission.ID, err)
		}
	}

	awdChallenges := []challengecontracts.AWDChallenge{
		{ID: 901, Name: "web-awd-easy", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyEasy, Status: challengecontracts.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 902, Name: "web-awd-medium", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyMedium, Status: challengecontracts.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
	}
	for _, challenge := range awdChallenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed awd challenge %s: %v", challenge.Name, err)
		}
	}

	logs := []contestcontracts.AWDAttackLog{
		{ID: 11, RoundID: 81, AttackerTeamID: 91, VictimTeamID: 101, ServiceID: 111, AWDChallengeID: 901, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: false, ScoreGained: 0, SubmittedByUserID: ptrTeachingAnalysisInt64(11), CreatedAt: now.Add(3 * time.Minute)},
		{ID: 12, RoundID: 81, AttackerTeamID: 91, VictimTeamID: 102, ServiceID: 112, AWDChallengeID: 901, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 80, SubmittedByUserID: ptrTeachingAnalysisInt64(11), CreatedAt: now.Add(4 * time.Minute)},
		{ID: 13, RoundID: 82, AttackerTeamID: 91, VictimTeamID: 103, ServiceID: 113, AWDChallengeID: 902, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 90, SubmittedByUserID: ptrTeachingAnalysisInt64(11), CreatedAt: now.Add(5 * time.Minute)},
		{ID: 14, RoundID: 82, AttackerTeamID: 91, VictimTeamID: 104, ServiceID: 114, AWDChallengeID: 902, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceManual, IsSuccess: true, ScoreGained: 100, SubmittedByUserID: ptrTeachingAnalysisInt64(11), CreatedAt: now.Add(6 * time.Minute)},
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
		t.Fatalf("expected only Class A snapshot, got %+v", snapshots)
	}

	snapshot := snapshots[0]
	if snapshot.UserID != 11 {
		t.Fatalf("expected Class A student snapshot, got %+v", snapshot)
	}
	if snapshot.CorrectSubmissionCount != 4 {
		t.Fatalf("expected explicit success count to include contest and awd successes, got %+v", snapshot)
	}
	if snapshot.WrongSubmissionCount != 2 {
		t.Fatalf("expected explicit failure count to include challenge and awd failures, got %+v", snapshot)
	}
	if snapshot.ChallengeSuccessCount != 2 {
		t.Fatalf("expected challenge success count to include contest submissions, got %+v", snapshot)
	}
	if snapshot.SubmissionSuccessCount != 4 || snapshot.SubmissionFailureCount != 2 {
		t.Fatalf("expected explicit submission counters, got %+v", snapshot)
	}
	if snapshot.AWDAttemptCount != 3 || snapshot.AWDSuccessCount != 2 {
		t.Fatalf("expected awd attempt/success counters, got %+v", snapshot)
	}

	web := findSnapshotDimension(t, snapshot, taxonomy.DimensionWeb)
	if web.ProfileScore != 1 {
		t.Fatalf("expected web profile score lifted by awd coverage, got %+v", web)
	}
	if web.AttemptCount != 6 {
		t.Fatalf("expected attempts to include contest submissions and awd attempts, got %+v", web)
	}
	if web.SuccessCount != 4 {
		t.Fatalf("expected success count to include contest submissions and awd successes, got %+v", web)
	}
	if web.EvidenceCount != 7 {
		t.Fatalf("expected evidence count to include contest approved review, got %+v", web)
	}
	if web.SolvedDifficultyCounts[taxonomy.DifficultyEasy] != 2 || web.SolvedDifficultyCounts[taxonomy.DifficultyMedium] != 2 {
		t.Fatalf("expected solved difficulty counts to merge contest solves and awd coverage, got %+v", web.SolvedDifficultyCounts)
	}
}

func ptrTeachingAnalysisInt64(value int64) *int64 {
	return &value
}
