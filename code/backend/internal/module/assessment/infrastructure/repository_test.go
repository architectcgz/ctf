package infrastructure

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	assessmententity "ctf-platform/internal/module/assessment/entity"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/internal/shared/taxonomy"
	teachingadvice "ctf-platform/internal/teaching/advice"
)

type assessmentChallengeRepoRow struct {
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

func (assessmentChallengeRepoRow) TableName() string {
	return "challenges"
}

func setupAssessmentRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&identitycontracts.User{},
		&assessmentChallengeRepoRow{},
		&contestcontracts.Submission{},
		&assessmententity.SkillProfile{},
		&challengecontracts.AWDChallenge{},
		&contestcontracts.AWDAttackLog{},
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

	if err := db.Create(&identitycontracts.User{
		ID:        7,
		Username:  "alice",
		Role:      identitycontracts.RoleStudent,
		ClassName: "Class A",
		Status:    identitycontracts.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&assessmententity.SkillProfile{
		UserID:    7,
		Dimension: taxonomy.DimensionWeb,
		Score:     0.2,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed web profile: %v", err)
	}

	awdChallenges := []challengecontracts.AWDChallenge{
		{ID: 701, Name: "web-awd-easy-a", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyEasy, Status: challengecontracts.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 702, Name: "web-awd-easy-b", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyEasy, Status: challengecontracts.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 703, Name: "web-awd-medium-a", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyMedium, Status: challengecontracts.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 704, Name: "web-awd-medium-b", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyMedium, Status: challengecontracts.AWDChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
	}
	for _, challenge := range awdChallenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed awd challenge %s: %v", challenge.Name, err)
		}
	}

	logs := []contestcontracts.AWDAttackLog{
		{ID: 1, RoundID: 81, AttackerTeamID: 91, VictimTeamID: 101, ServiceID: 111, AWDChallengeID: 701, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 80, SubmittedByUserID: ptrAssessmentRepoInt64(7), CreatedAt: now},
		{ID: 2, RoundID: 81, AttackerTeamID: 91, VictimTeamID: 102, ServiceID: 112, AWDChallengeID: 702, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 80, SubmittedByUserID: ptrAssessmentRepoInt64(7), CreatedAt: now.Add(1 * time.Minute)},
		{ID: 3, RoundID: 82, AttackerTeamID: 91, VictimTeamID: 103, ServiceID: 113, AWDChallengeID: 703, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 90, SubmittedByUserID: ptrAssessmentRepoInt64(7), CreatedAt: now.Add(2 * time.Minute)},
		{ID: 4, RoundID: 82, AttackerTeamID: 91, VictimTeamID: 104, ServiceID: 114, AWDChallengeID: 704, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 90, SubmittedByUserID: ptrAssessmentRepoInt64(7), CreatedAt: now.Add(3 * time.Minute)},
		{ID: 5, RoundID: 83, AttackerTeamID: 91, VictimTeamID: 105, ServiceID: 115, AWDChallengeID: 704, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 0, SubmittedByUserID: ptrAssessmentRepoInt64(7), CreatedAt: now.Add(4 * time.Minute)},
		{ID: 6, RoundID: 83, AttackerTeamID: 91, VictimTeamID: 106, ServiceID: 116, AWDChallengeID: 704, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceManual, IsSuccess: true, ScoreGained: 100, SubmittedByUserID: ptrAssessmentRepoInt64(7), CreatedAt: now.Add(5 * time.Minute)},
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

	web := findAssessmentSnapshotDimension(t, snapshot, taxonomy.DimensionWeb)
	if web.ProfileScore != 1 {
		t.Fatalf("expected web profile score lifted by awd coverage, got %+v", web)
	}
	if web.SuccessCount != 4 {
		t.Fatalf("expected web success count to include awd success coverage, got %+v", web)
	}
	if web.AttemptCount != 5 {
		t.Fatalf("expected web attempt count to include awd submission attempts, got %+v", web)
	}
	if web.EvidenceCount != 5 {
		t.Fatalf("expected web evidence count to include awd submission attempts, got %+v", web)
	}
	if web.SolvedDifficultyCounts[taxonomy.DifficultyEasy] != 2 || web.SolvedDifficultyCounts[taxonomy.DifficultyMedium] != 2 {
		t.Fatalf("expected awd difficulty coverage merged into snapshot, got %+v", web.SolvedDifficultyCounts)
	}
}

func TestRepositoryGetStudentTeachingFactSnapshotUsesUnifiedContestAndAWDSubmissionSemantics(t *testing.T) {
	db := setupAssessmentRepoTestDB(t)
	repo := NewRepository(db)
	now := time.Now().UTC()
	contestID := int64(55)

	if err := db.Create(&identitycontracts.User{
		ID:        8,
		Username:  "bob",
		Role:      identitycontracts.RoleStudent,
		ClassName: "Class A",
		Status:    identitycontracts.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	if err := db.Create(&assessmententity.SkillProfile{
		UserID:    8,
		Dimension: taxonomy.DimensionWeb,
		Score:     0.15,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed web profile: %v", err)
	}

	challenges := []assessmentChallengeRepoRow{
		{ID: 801, Title: "web-practice", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyEasy, Points: 100, Status: challengecontracts.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 802, Title: "web-contest", Category: taxonomy.DimensionWeb, Difficulty: taxonomy.DifficultyMedium, Points: 120, Status: challengecontracts.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
	}
	for _, challenge := range challenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed challenge %s: %v", challenge.Title, err)
		}
	}

	submissions := []contestcontracts.Submission{
		{ID: 1, UserID: 8, ChallengeID: 801, IsCorrect: false, SubmittedAt: now, UpdatedAt: now},
		{ID: 2, UserID: 8, ChallengeID: 801, ContestID: &contestID, IsCorrect: true, ReviewStatus: contestcontracts.SubmissionReviewStatusApproved, SubmittedAt: now.Add(1 * time.Minute), UpdatedAt: now.Add(1 * time.Minute)},
		{ID: 3, UserID: 8, ChallengeID: 802, ContestID: &contestID, IsCorrect: true, SubmittedAt: now.Add(2 * time.Minute), UpdatedAt: now.Add(2 * time.Minute)},
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
		{ID: 11, RoundID: 81, AttackerTeamID: 91, VictimTeamID: 101, ServiceID: 111, AWDChallengeID: 901, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: false, ScoreGained: 0, SubmittedByUserID: ptrAssessmentRepoInt64(8), CreatedAt: now.Add(3 * time.Minute)},
		{ID: 12, RoundID: 81, AttackerTeamID: 91, VictimTeamID: 102, ServiceID: 112, AWDChallengeID: 901, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 80, SubmittedByUserID: ptrAssessmentRepoInt64(8), CreatedAt: now.Add(4 * time.Minute)},
		{ID: 13, RoundID: 82, AttackerTeamID: 91, VictimTeamID: 103, ServiceID: 113, AWDChallengeID: 902, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceSubmission, IsSuccess: true, ScoreGained: 90, SubmittedByUserID: ptrAssessmentRepoInt64(8), CreatedAt: now.Add(5 * time.Minute)},
		{ID: 14, RoundID: 82, AttackerTeamID: 91, VictimTeamID: 104, ServiceID: 114, AWDChallengeID: 902, AttackType: contestcontracts.AWDAttackTypeFlagCapture, Source: contestcontracts.AWDAttackSourceManual, IsSuccess: true, ScoreGained: 100, SubmittedByUserID: ptrAssessmentRepoInt64(8), CreatedAt: now.Add(6 * time.Minute)},
	}
	for _, log := range logs {
		if err := db.Create(&log).Error; err != nil {
			t.Fatalf("seed awd attack log %d: %v", log.ID, err)
		}
	}

	snapshot, err := repo.GetStudentTeachingFactSnapshot(context.Background(), 8)
	if err != nil {
		t.Fatalf("GetStudentTeachingFactSnapshot() error = %v", err)
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
		t.Fatalf("expected awd attempt/success counters to reflect student submission logs, got %+v", snapshot)
	}

	web := findAssessmentSnapshotDimension(t, snapshot, taxonomy.DimensionWeb)
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

func TestRepositoryGetDimensionScoresCachesPublishedDimensionTotals(t *testing.T) {
	db := setupAssessmentRepoTestDB(t)
	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })

	repo := NewRepository(
		db,
		WithPublishedDimensionTotalCache(
			NewDimensionTotalCacheStore(redisClient),
			time.Minute,
		),
	)
	now := time.Now().UTC()

	if err := db.Create(&assessmentChallengeRepoRow{
		ID:         901,
		Title:      "web-initial",
		Category:   taxonomy.DimensionWeb,
		Difficulty: taxonomy.DifficultyEasy,
		Points:     100,
		Status:     challengecontracts.ChallengeStatusPublished,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}
	if err := db.Create(&contestcontracts.Submission{
		UserID:      7,
		ChallengeID: 901,
		IsCorrect:   true,
		SubmittedAt: now,
	}).Error; err != nil {
		t.Fatalf("seed submission: %v", err)
	}

	first, err := repo.GetDimensionScores(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetDimensionScores() first error = %v", err)
	}
	if len(first) != 1 || first[0].TotalScore != 100 || first[0].UserScore != 100 {
		t.Fatalf("unexpected initial dimension scores: %+v", first)
	}

	if err := db.Model(&assessmentChallengeRepoRow{}).Where("id = ?", 901).Update("points", 250).Error; err != nil {
		t.Fatalf("update challenge points: %v", err)
	}

	cached, err := repo.GetDimensionScores(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetDimensionScores() cached error = %v", err)
	}
	if len(cached) != 1 || cached[0].TotalScore != 100 || cached[0].UserScore != 250 {
		t.Fatalf("expected cached total and fresh user score, got %+v", cached)
	}

	cacheStore := NewDimensionTotalCacheStore(redisClient)
	if err := cacheStore.DeletePublishedDimensionTotals(context.Background()); err != nil {
		t.Fatalf("DeletePublishedDimensionTotals() error = %v", err)
	}

	refreshed, err := repo.GetDimensionScores(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetDimensionScores() refreshed error = %v", err)
	}
	if len(refreshed) != 1 || refreshed[0].TotalScore != 250 || refreshed[0].UserScore != 250 {
		t.Fatalf("expected refreshed totals after cache clear, got %+v", refreshed)
	}
}

func ptrAssessmentRepoInt64(value int64) *int64 {
	return &value
}
