package system

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func setupAuditTestDB(t *testing.T) *gorm.DB {
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

func newAuditTestService(db *gorm.DB) *AuditService {
	return NewAuditService(NewAuditRepository(db), config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     50,
	}, zap.NewNop())
}

func TestAuditServiceRecordPersistsEntry(t *testing.T) {
	db := setupAuditTestDB(t)
	service := newAuditTestService(db)

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

func TestAuditServiceListAuditLogsNormalizesPaginationAndDetails(t *testing.T) {
	db := setupAuditTestDB(t)
	service := newAuditTestService(db)

	now := time.Now().UTC().Truncate(time.Second)
	users := []model.User{
		{ID: 1, Username: "admin", Role: model.RoleAdmin, Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now},
		{ID: 2, Username: "alice", Role: model.RoleStudent, Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now},
	}
	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}

	actorID := int64(1)
	targetID := int64(11)
	entries := []model.AuditLog{
		{
			ID:           1,
			UserID:       &actorID,
			Action:       model.AuditActionLogin,
			ResourceType: "auth",
			Detail:       `{"username":"admin","result":"success"}`,
			IPAddress:    "10.0.0.1",
			CreatedAt:    now.Add(-2 * time.Minute),
		},
		{
			ID:           2,
			Action:       model.AuditActionSubmit,
			ResourceType: "challenge",
			ResourceID:   &targetID,
			Detail:       `{"username":"ghost","result":"accepted"}`,
			IPAddress:    "",
			CreatedAt:    now.Add(-time.Minute),
		},
	}
	for _, entry := range entries {
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("seed audit log: %v", err)
		}
	}

	items, total, page, pageSize, err := service.ListAuditLogs(context.Background(), &dto.AuditLogQuery{
		StartTime: now.Add(-10 * time.Minute).Format(time.RFC3339),
		EndTime:   now.Add(time.Minute).Format(time.RFC3339),
		Page:      0,
		PageSize:  999,
	})
	if err != nil {
		t.Fatalf("ListAuditLogs() error = %v", err)
	}
	if total != 2 || page != 1 || pageSize != 50 {
		t.Fatalf("unexpected paging result total=%d page=%d pageSize=%d", total, page, pageSize)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %+v", items)
	}
	if items[0].ActorUsername != "ghost" {
		t.Fatalf("expected fallback username from detail, got %+v", items[0])
	}
	if items[0].IP != nil {
		t.Fatalf("expected empty ip to stay nil, got %+v", items[0])
	}
	if items[1].ActorUsername != "admin" {
		t.Fatalf("expected joined username for known actor, got %+v", items[1])
	}
	if items[1].Detail["result"] != "success" {
		t.Fatalf("expected parsed detail, got %+v", items[1].Detail)
	}
}

func TestAuditServiceListAuditLogsRejectsInvalidTimeRange(t *testing.T) {
	service := newAuditTestService(setupAuditTestDB(t))

	_, _, _, _, err := service.ListAuditLogs(context.Background(), &dto.AuditLogQuery{
		StartTime: "bad-time",
	})
	if err == nil {
		t.Fatal("expected invalid start_time error")
	}
	appErr, ok := err.(*errcode.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != errcode.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params code, got %+v", appErr)
	}
}
