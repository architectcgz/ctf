package infrastructure

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

func TestGetUserTimelineUsesDestroyedAtForInstanceDestroyEvent(t *testing.T) {
	t.Parallel()

	db := newPracticeTimelineRepositoryTestDB(t)

	repo := NewRepository(db)
	now := time.Date(2026, 4, 23, 11, 0, 0, 0, time.UTC)
	destroyedAt := now.Add(-15 * time.Minute)
	updatedAt := now.Add(-3 * time.Minute)

	challenge := model.Challenge{
		ID:         41,
		Title:      "web-baby",
		Category:   "web",
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		Status:     model.ChallengeStatusPublished,
		CreatedAt:  now.Add(-2 * time.Hour),
		UpdatedAt:  now.Add(-2 * time.Hour),
	}
	if err := db.Create(&challenge).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}

	instance := model.Instance{
		ID:          101,
		UserID:      7,
		ChallengeID: challenge.ID,
		ContainerID: "inst-stopped",
		Status:      model.InstanceStatusStopped,
		CreatedAt:   now.Add(-40 * time.Minute),
		UpdatedAt:   updatedAt,
		ExpiresAt:   now.Add(20 * time.Minute),
	}
	if err := db.Create(&instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if err := db.Exec("UPDATE instances SET destroyed_at = ? WHERE id = ?", destroyedAt, instance.ID).Error; err != nil {
		t.Fatalf("seed destroyed_at: %v", err)
	}

	events, err := repo.GetUserTimeline(context.Background(), instance.UserID, 20, 0)
	if err != nil {
		t.Fatalf("GetUserTimeline() error = %v", err)
	}

	for _, event := range events {
		if event.Type != "instance_destroy" {
			continue
		}
		if !event.Timestamp.Equal(destroyedAt) {
			t.Fatalf("instance_destroy timestamp = %v, want %v", event.Timestamp, destroyedAt)
		}
		return
	}

	t.Fatalf("expected instance_destroy event, got %+v", events)
}

func TestRepositoryDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("repository.go")
	if err != nil {
		t.Fatalf("read repository.go: %v", err)
	}
	if strings.Contains(string(source), "context.Background()") {
		t.Fatal("repository should not create context.Background()")
	}
}

func newPracticeTimelineRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Challenge{}, &model.Instance{}, &model.Submission{}, &model.AuditLog{}); err != nil {
		t.Fatalf("migrate sqlite: %v", err)
	}
	return db
}
