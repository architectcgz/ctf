package commands_test

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/model"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsinfra "ctf-platform/internal/module/ops/infrastructure"
)

func setupAuditCommandTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.AuditLog{}); err != nil {
		t.Fatalf("migrate audit tables: %v", err)
	}
	return db
}

func TestAuditServiceRecordPersistsEntry(t *testing.T) {
	db := setupAuditCommandTestDB(t)
	service := opscmd.NewAuditService(opsinfra.NewAuditRepository(db), zap.NewNop())

	userID := int64(7)
	resourceID := int64(23)
	userAgent := "unit-test"
	err := service.Record(context.Background(), auditlog.Entry{
		UserID:       &userID,
		Action:       model.AuditActionSubmit,
		ResourceType: "challenge",
		ResourceID:   &resourceID,
		Detail: map[string]any{
			"status":   "ok",
			"username": "alice",
		},
		IPAddress: "127.0.0.1",
		UserAgent: &userAgent,
	})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	var saved model.AuditLog
	if err := db.First(&saved).Error; err != nil {
		t.Fatalf("query saved audit log: %v", err)
	}
	if saved.Action != model.AuditActionSubmit || saved.ResourceType != "challenge" {
		t.Fatalf("unexpected saved log: %+v", saved)
	}
	if saved.UserID == nil || *saved.UserID != userID {
		t.Fatalf("expected saved user id, got %+v", saved)
	}
	if saved.Detail == "" || saved.Detail == "{}" {
		t.Fatalf("expected serialized detail, got %+v", saved)
	}
	if saved.UserAgent == nil || *saved.UserAgent != userAgent {
		t.Fatalf("expected user agent persisted, got %+v", saved)
	}
}
