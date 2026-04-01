package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/auditlog"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type recordingAuditRecorder struct {
	entries []auditlog.Entry
}

func (r *recordingAuditRecorder) Record(_ context.Context, entry auditlog.Entry) error {
	r.entries = append(r.entries, entry)
	return nil
}

type recordingProxyTrafficRecorder struct {
	instanceID  int64
	userID      int64
	method      string
	requestPath string
	statusCode  int
	callCount   int
}

func (r *recordingProxyTrafficRecorder) RecordRuntimeProxyTrafficEvent(_ context.Context, instanceID, userID int64, method, requestPath string, statusCode int) error {
	r.instanceID = instanceID
	r.userID = userID
	r.method = method
	r.requestPath = requestPath
	r.statusCode = statusCode
	r.callCount++
	return nil
}

func TestRecordProxyAuditAlsoRecordsAWDTrafficEvent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/instances/100/proxy/admin/login", nil)
	ctx.Set("request_id", "req-traffic-1")

	auditRecorder := &recordingAuditRecorder{}
	trafficRecorder := &recordingProxyTrafficRecorder{}
	h := NewHandler(stubRuntimeService{}, auditRecorder, CookieConfig{}, trafficRecorder)

	h.recordProxyAudit(
		ctx,
		&runtimeports.ProxyTicketClaims{UserID: 2001, Username: "attacker", InstanceID: 100},
		100,
		"attacker",
		"req-traffic-1",
		"/admin/login",
		http.StatusInternalServerError,
		"",
		false,
		false,
	)

	if len(auditRecorder.entries) != 1 {
		t.Fatalf("expected audit recorder called once, got %d", len(auditRecorder.entries))
	}
	if trafficRecorder.callCount != 1 {
		t.Fatalf("expected traffic recorder called once, got %d", trafficRecorder.callCount)
	}
	if trafficRecorder.instanceID != 100 || trafficRecorder.userID != 2001 {
		t.Fatalf("unexpected traffic recorder ids: %+v", trafficRecorder)
	}
	if trafficRecorder.method != http.MethodPost || trafficRecorder.requestPath != "/admin/login" {
		t.Fatalf("unexpected traffic recorder request: %+v", trafficRecorder)
	}
	if trafficRecorder.statusCode != http.StatusInternalServerError {
		t.Fatalf("unexpected traffic recorder status: %+v", trafficRecorder)
	}
}
