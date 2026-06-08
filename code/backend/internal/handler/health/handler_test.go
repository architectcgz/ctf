package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	servicehealth "ctf-platform/internal/service/health"
)

type stubHealthService struct {
	live  *servicehealth.Status
	ready *servicehealth.Status
}

func (s stubHealthService) Check(context.Context) *servicehealth.Status {
	return servicehealth.NewStatus(servicehealth.HealthStatus{Status: "ok"}, true)
}

func (s stubHealthService) CheckLive(context.Context) *servicehealth.Status {
	return s.live
}

func (s stubHealthService) CheckReady(context.Context) *servicehealth.Status {
	return s.ready
}

func (s stubHealthService) CheckDB(context.Context) error {
	return nil
}

func (s stubHealthService) CheckRedis(context.Context) error {
	return nil
}

func TestGetLiveReturnsProcessLiveness(t *testing.T) {
	t.Parallel()

	handler := NewHandler(stubHealthService{
		live: servicehealth.NewStatus(servicehealth.HealthStatus{Status: "ok"}, true),
	})

	resp := performHealthRequest(http.MethodGet, "/live", handler.GetLive)
	if resp.Code != http.StatusOK {
		t.Fatalf("GetLive() status = %d, want 200", resp.Code)
	}
}

func TestGetReadyReturnsUnavailableWhenNotReady(t *testing.T) {
	t.Parallel()

	handler := NewHandler(stubHealthService{
		ready: servicehealth.NewStatus(servicehealth.HealthStatus{Status: "not_ready"}, false),
	})

	resp := performHealthRequest(http.MethodGet, "/ready", handler.GetReady)
	if resp.Code != http.StatusServiceUnavailable {
		t.Fatalf("GetReady() status = %d, want 503", resp.Code)
	}
}

func performHealthRequest(method, target string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Handle(method, target, handler)

	req := httptest.NewRequest(method, target, nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)
	return resp
}
