package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	assessmentapp "ctf-platform/internal/module/assessment/application"
	assessmentinfra "ctf-platform/internal/module/assessment/infrastructure"
	"ctf-platform/pkg/errcode"
)

func setupAssessmentTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.Challenge{},
		&model.Submission{},
		&model.SkillProfile{},
	); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	return db
}

func newAssessmentTestService(db *gorm.DB, redisClient *redis.Client) *assessmentapp.Service {
	return assessmentapp.NewService(
		assessmentinfra.NewRepository(db),
		redisClient,
		config.AssessmentConfig{
			RedisKeyPrefix: "assessment:test",
			LockTTL:        time.Minute,
		},
		zap.NewNop(),
	)
}

func TestCalculateSkillProfileWithContextPersistsComputedScores(t *testing.T) {
	db := setupAssessmentTestDB(t)
	service := newAssessmentTestService(db, nil)
	now := time.Now()

	student := model.User{
		ID:        1,
		Username:  "alice",
		Role:      model.RoleStudent,
		ClassName: "Class A",
		Status:    model.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&student).Error; err != nil {
		t.Fatalf("seed student: %v", err)
	}

	challenges := []model.Challenge{
		{ID: 11, Title: "web-1", Category: model.DimensionWeb, Difficulty: model.ChallengeDifficultyEasy, Points: 100, Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 12, Title: "web-2", Category: model.DimensionWeb, Difficulty: model.ChallengeDifficultyMedium, Points: 50, Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 13, Title: "crypto-1", Category: model.DimensionCrypto, Difficulty: model.ChallengeDifficultyEasy, Points: 200, Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now},
		{ID: 14, Title: "draft-ignored", Category: model.DimensionPwn, Difficulty: model.ChallengeDifficultyEasy, Points: 300, Status: model.ChallengeStatusDraft, CreatedAt: now, UpdatedAt: now},
	}
	for _, challenge := range challenges {
		if err := db.Create(&challenge).Error; err != nil {
			t.Fatalf("seed challenge: %v", err)
		}
	}

	submissions := []model.Submission{
		{UserID: student.ID, ChallengeID: 11, IsCorrect: true, SubmittedAt: now},
		{UserID: student.ID, ChallengeID: 12, IsCorrect: false, SubmittedAt: now},
		{UserID: student.ID, ChallengeID: 13, IsCorrect: true, SubmittedAt: now},
	}
	for _, submission := range submissions {
		if err := db.Create(&submission).Error; err != nil {
			t.Fatalf("seed submission: %v", err)
		}
	}

	dimensions, err := service.CalculateSkillProfileWithContext(context.Background(), student.ID)
	if err != nil {
		t.Fatalf("CalculateSkillProfileWithContext() error = %v", err)
	}
	if len(dimensions) != 2 {
		t.Fatalf("expected 2 computed dimensions, got %+v", dimensions)
	}

	scoreByDimension := make(map[string]float64, len(dimensions))
	for _, item := range dimensions {
		scoreByDimension[item.Dimension] = item.Score
	}
	if scoreByDimension[model.DimensionWeb] != float64(100)/float64(150) {
		t.Fatalf("unexpected web score map: %+v", scoreByDimension)
	}
	if scoreByDimension[model.DimensionCrypto] != 1 {
		t.Fatalf("unexpected crypto score map: %+v", scoreByDimension)
	}

	var profiles []model.SkillProfile
	if err := db.Order("dimension ASC").Find(&profiles).Error; err != nil {
		t.Fatalf("query profiles: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 persisted profiles, got %+v", profiles)
	}
}

func TestCalculateSkillProfileWithContextReturnsExistingProfileWhenLocked(t *testing.T) {
	db := setupAssessmentTestDB(t)
	now := time.Now()

	profiles := []model.SkillProfile{
		{UserID: 2, Dimension: model.DimensionWeb, Score: 0.75, UpdatedAt: now},
		{UserID: 2, Dimension: model.DimensionCrypto, Score: 0.25, UpdatedAt: now},
	}
	for _, profile := range profiles {
		if err := db.Create(&profile).Error; err != nil {
			t.Fatalf("seed profile: %v", err)
		}
	}

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	if err := redisClient.Set(context.Background(), "assessment:test:lock:2", "1", time.Minute).Err(); err != nil {
		t.Fatalf("seed lock key: %v", err)
	}

	service := newAssessmentTestService(db, redisClient)
	dimensions, err := service.CalculateSkillProfileWithContext(context.Background(), 2)
	if err != nil {
		t.Fatalf("CalculateSkillProfileWithContext() error = %v", err)
	}
	if len(dimensions) != 2 {
		t.Fatalf("expected existing profile dimensions, got %+v", dimensions)
	}

	scoreByDimension := make(map[string]float64, len(dimensions))
	for _, item := range dimensions {
		scoreByDimension[item.Dimension] = item.Score
	}
	if scoreByDimension[model.DimensionWeb] != 0.75 || scoreByDimension[model.DimensionCrypto] != 0.25 {
		t.Fatalf("expected fallback scores from db, got %+v", scoreByDimension)
	}
}

func TestGetSkillProfileReturnsEmptyDimensionsWhenProfileMissing(t *testing.T) {
	service := newAssessmentTestService(setupAssessmentTestDB(t), nil)

	profile, err := service.GetSkillProfile(42)
	if err != nil {
		t.Fatalf("GetSkillProfile() error = %v", err)
	}
	if profile.UserID != 42 {
		t.Fatalf("expected user id 42, got %+v", profile)
	}
	if profile.UpdatedAt != "" {
		t.Fatalf("expected empty updated_at, got %+v", profile)
	}
	if len(profile.Dimensions) != len(model.AllDimensions) {
		t.Fatalf("expected all dimensions, got %+v", profile.Dimensions)
	}
	for _, item := range profile.Dimensions {
		if item.Score != 0 {
			t.Fatalf("expected zero score for empty profile, got %+v", profile.Dimensions)
		}
	}
}

func TestGetStudentSkillProfileRejectsTeacherFromOtherClass(t *testing.T) {
	db := setupAssessmentTestDB(t)
	service := newAssessmentTestService(db, nil)
	now := time.Now()

	users := []model.User{
		{ID: 10, Username: "teacher-a", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now},
		{ID: 20, Username: "student-b", Role: model.RoleStudent, ClassName: "Class B", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now},
	}
	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}

	_, err := service.GetStudentSkillProfile(context.Background(), 10, model.RoleTeacher, 20)
	if err == nil {
		t.Fatal("expected forbidden error for cross-class teacher access")
	}
	appErr, ok := err.(*errcode.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != errcode.ErrForbidden.Code {
		t.Fatalf("expected forbidden code, got %+v", appErr)
	}
}

func TestCalculateSkillProfileWithContextHonorsCancellation(t *testing.T) {
	db := setupAssessmentTestDB(t)
	service := newAssessmentTestService(db, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.CalculateSkillProfileWithContext(ctx, 1)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestGetSkillProfileWithContextHonorsCancellation(t *testing.T) {
	db := setupAssessmentTestDB(t)
	service := newAssessmentTestService(db, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := service.GetSkillProfileWithContext(ctx, 1)
	if err == nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}
