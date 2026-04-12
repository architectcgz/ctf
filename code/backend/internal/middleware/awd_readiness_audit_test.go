package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/model"
)

type recordingAWDReadinessAuditRecorder struct {
	entries []auditlog.Entry
}

func (r *recordingAWDReadinessAuditRecorder) Record(_ context.Context, entry auditlog.Entry) error {
	r.entries = append(r.entries, entry)
	return nil
}

func TestAWDReadinessAuditRecordsSucceededForcedOverride(t *testing.T) {
	entry := runAWDReadinessAuditRequest(t, http.MethodPost, "/admin/contests/42/awd/rounds", func(c *gin.Context) {
		authctx.SetCurrentUser(c, authctx.CurrentUser{UserID: 1001, Username: "admin"})
		c.Set(RequestIDKey, "req-awd-audit-succeeded")
		SetAWDReadinessAuditPayload(c, &AWDReadinessAuditPayload{
			GateAction:            "create_round",
			ForceOverride:         true,
			GateAllowed:           true,
			OverrideReason:        "teacher drill",
			BlockingCount:         2,
			GlobalBlockingReasons: []string{"no_challenges"},
			BlockingItems: []AWDReadinessAuditItem{
				{
					ChallengeID:     11,
					Title:           "calc",
					CheckerType:     model.AWDCheckerType("http_standard"),
					ValidationState: "failed",
					BlockingReason:  "last_preview_failed",
				},
			},
			ExecutionOutcome: "succeeded",
		})
		c.Status(http.StatusOK)
	})

	assertAWDReadinessAuditEntry(t, entry, 42)
	detail := entry.Detail
	if detail["gate_action"] != "create_round" {
		t.Fatalf("unexpected gate action: %+v", detail)
	}
	if detail["override_reason"] != "teacher drill" {
		t.Fatalf("unexpected override reason: %+v", detail)
	}
	if detail["blocking_count"] != 2 {
		t.Fatalf("unexpected blocking count: %+v", detail)
	}
	if detail["execution_outcome"] != "succeeded" {
		t.Fatalf("unexpected execution outcome: %+v", detail)
	}
	if detail["execution_error"] != "" {
		t.Fatalf("unexpected execution error: %+v", detail)
	}
}

func TestAWDReadinessAuditRecordsFailedForcedOverride(t *testing.T) {
	entry := runAWDReadinessAuditRequest(t, http.MethodPut, "/admin/contests/42", func(c *gin.Context) {
		authctx.SetCurrentUser(c, authctx.CurrentUser{UserID: 1001, Username: "admin"})
		c.Set(RequestIDKey, "req-awd-audit-failed")
		SetAWDReadinessAuditPayload(c, &AWDReadinessAuditPayload{
			GateAction:            "start_contest",
			ForceOverride:         true,
			GateAllowed:           true,
			OverrideReason:        "teacher drill",
			BlockingCount:         1,
			GlobalBlockingReasons: []string{"missing_checker"},
			BlockingItems: []AWDReadinessAuditItem{
				{
					ChallengeID:     21,
					Title:           "pwn",
					CheckerType:     model.AWDCheckerType("legacy_probe"),
					ValidationState: "missing_checker",
					BlockingReason:  "missing_checker",
				},
			},
			ExecutionOutcome: "failed",
			ExecutionError:   "db update failed",
		})
		c.Status(http.StatusInternalServerError)
	})

	assertAWDReadinessAuditEntry(t, entry, 42)
	detail := entry.Detail
	if detail["gate_action"] != "start_contest" {
		t.Fatalf("unexpected gate action: %+v", detail)
	}
	if detail["execution_outcome"] != "failed" {
		t.Fatalf("unexpected execution outcome: %+v", detail)
	}
	if detail["execution_error"] != "db update failed" {
		t.Fatalf("unexpected execution error: %+v", detail)
	}
}

func runAWDReadinessAuditRequest(t *testing.T, method, path string, handler gin.HandlerFunc) auditlog.Entry {
	t.Helper()
	gin.SetMode(gin.TestMode)

	recorder := &recordingAWDReadinessAuditRecorder{}
	engine := gin.New()
	engine.Use(AWDReadinessAudit(recorder, zap.NewNop()))
	engine.PUT("/admin/contests/:id", handler)
	engine.POST("/admin/contests/:id/awd/rounds", handler)

	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("User-Agent", "awd-audit-test")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if len(recorder.entries) != 1 {
		t.Fatalf("expected one audit entry, got %d", len(recorder.entries))
	}
	return recorder.entries[0]
}

func assertAWDReadinessAuditEntry(t *testing.T, entry auditlog.Entry, contestID int64) {
	t.Helper()
	if entry.Action != model.AuditActionAdminOp || entry.ResourceType != "contest" {
		t.Fatalf("unexpected audit envelope: %+v", entry)
	}
	if entry.ResourceID == nil || *entry.ResourceID != contestID {
		t.Fatalf("unexpected resource id: %+v", entry.ResourceID)
	}
	detail := entry.Detail
	if detail["module"] != "awd_readiness_gate" {
		t.Fatalf("unexpected module: %+v", detail)
	}
}
