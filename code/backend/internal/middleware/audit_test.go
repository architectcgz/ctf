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

const skipAuditKey = "skip_audit"

type recordingAuditRecorder struct {
	entries []auditlog.Entry
}

func (r *recordingAuditRecorder) Record(_ context.Context, entry auditlog.Entry) error {
	r.entries = append(r.entries, entry)
	return nil
}

func TestAuditSkipsWhenContextMarked(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	recorder := &recordingAuditRecorder{}
	engine := gin.New()
	engine.Use(Audit(recorder, AuditOptions{
		Action:          model.AuditActionSubmit,
		ResourceType:    "challenge_submission",
		ResourceIDParam: "id",
	}, zap.NewNop()))
	engine.POST("/api/v1/challenges/:id/submit", func(c *gin.Context) {
		authctx.SetCurrentUser(c, authctx.CurrentUser{UserID: 7, Username: "student7"})
		c.Set(RequestIDKey, "req-audit-skip")
		c.Set(skipAuditKey, true)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/challenges/11/submit", nil)
	req.Header.Set("User-Agent", "audit-skip-test")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected response status: %d", resp.Code)
	}
	if len(recorder.entries) != 0 {
		t.Fatalf("expected no audit entry when skip flag is set, got %d", len(recorder.entries))
	}
}
