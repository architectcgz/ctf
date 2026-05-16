package infrastructure

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
)

func TestRefreshActiveAWDInstanceExpiryByContest(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Instance{}); err != nil {
		t.Fatalf("auto migrate instances: %v", err)
	}

	now := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	contestID := int64(41)
	serviceID := int64(81)
	otherContestID := int64(42)
	activeAt := now
	newExpiresAt := now.Add(40 * time.Minute)

	rows := []model.Instance{
		{ID: 1, UserID: 1, ContestID: &contestID, ServiceID: &serviceID, ChallengeID: 101, Status: model.InstanceStatusPending, ExpiresAt: now.Add(10 * time.Minute), CreatedAt: now, UpdatedAt: now},
		{ID: 2, UserID: 2, ContestID: &contestID, ServiceID: &serviceID, ChallengeID: 102, Status: model.InstanceStatusRunning, ExpiresAt: now.Add(5 * time.Minute), CreatedAt: now, UpdatedAt: now},
		{ID: 3, UserID: 3, ContestID: &contestID, ServiceID: &serviceID, ChallengeID: 103, Status: model.InstanceStatusRunning, ExpiresAt: now.Add(-time.Minute), CreatedAt: now, UpdatedAt: now},
		{ID: 4, UserID: 4, ContestID: &contestID, ChallengeID: 104, Status: model.InstanceStatusRunning, ExpiresAt: now.Add(5 * time.Minute), CreatedAt: now, UpdatedAt: now},
		{ID: 5, UserID: 5, ContestID: &otherContestID, ServiceID: &serviceID, ChallengeID: 105, Status: model.InstanceStatusRunning, ExpiresAt: now.Add(5 * time.Minute), CreatedAt: now, UpdatedAt: now},
	}
	for _, row := range rows {
		instance := row
		if err := db.Create(&instance).Error; err != nil {
			t.Fatalf("create instance %d: %v", instance.ID, err)
		}
	}

	repo := NewRepository(db)
	if err := repo.RefreshActiveAWDInstanceExpiryByContest(context.Background(), contestID, activeAt, newExpiresAt); err != nil {
		t.Fatalf("RefreshActiveAWDInstanceExpiryByContest() error = %v", err)
	}

	assertInstanceExpiry(t, db, 1, newExpiresAt)
	assertInstanceExpiry(t, db, 2, newExpiresAt)
	assertInstanceExpiry(t, db, 3, now.Add(-time.Minute))
	assertInstanceExpiry(t, db, 4, now.Add(5*time.Minute))
	assertInstanceExpiry(t, db, 5, now.Add(5*time.Minute))
}

func assertInstanceExpiry(t *testing.T, db *gorm.DB, instanceID int64, want time.Time) {
	t.Helper()

	var instance model.Instance
	if err := db.Where("id = ?", instanceID).First(&instance).Error; err != nil {
		t.Fatalf("load instance %d: %v", instanceID, err)
	}
	if !instance.ExpiresAt.Equal(want) {
		t.Fatalf("instance %d expiry = %s, want %s", instanceID, instance.ExpiresAt, want)
	}
}
