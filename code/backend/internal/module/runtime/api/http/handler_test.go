package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type stubRuntimeService struct{}

func (stubRuntimeService) DestroyInstance(context.Context, int64, int64) error {
	return nil
}

func (stubRuntimeService) ExtendInstance(context.Context, int64, int64) (*dto.InstanceResp, error) {
	return nil, nil
}

func (stubRuntimeService) GetAccessURL(context.Context, int64, int64) (string, error) {
	return "", nil
}

func (stubRuntimeService) GetUserInstances(context.Context, int64) ([]*dto.InstanceInfo, error) {
	return nil, nil
}

func (stubRuntimeService) ListTeacherInstances(context.Context, int64, string, *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error) {
	return nil, nil
}

func (stubRuntimeService) DestroyTeacherInstance(context.Context, int64, int64, string) error {
	return nil
}

func (stubRuntimeService) IssueProxyTicket(context.Context, authctx.CurrentUser, int64) (string, error) {
	return "", nil
}

func (stubRuntimeService) ResolveProxyTicket(context.Context, string) (*runtimeports.ProxyTicketClaims, error) {
	return nil, nil
}

func (stubRuntimeService) ProxyTicketMaxAge() int {
	return 0
}

func (stubRuntimeService) ProxyBodyPreviewSize() int {
	return 0
}

func TestHandlerContractsCompile(t *testing.T) {
	var _ runtimeService = stubRuntimeService{}
	_ = NewHandler(stubRuntimeService{}, nil, CookieConfig{}, nil)
}

type stubProxyRuntimeService struct {
	stubRuntimeService
	targetURL string
	claims    *runtimeports.ProxyTicketClaims
}

func (s stubProxyRuntimeService) GetAccessURL(context.Context, int64, int64) (string, error) {
	return s.targetURL, nil
}

func (s stubProxyRuntimeService) ResolveProxyTicket(context.Context, string) (*runtimeports.ProxyTicketClaims, error) {
	return s.claims, nil
}

type failingTrafficRecorder struct {
	calls int
}

func (r *failingTrafficRecorder) RecordRuntimeProxyTrafficEvent(context.Context, int64, int64, string, string, int) error {
	r.calls++
	return errors.New("persist failed")
}

func TestProxyInstanceTrafficRecorderFailureDoesNotAffectProxyResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("proxied"))
	}))
	defer target.Close()

	recorder := &failingTrafficRecorder{}
	handler := NewHandler(
		stubProxyRuntimeService{
			targetURL: target.URL,
			claims: &runtimeports.ProxyTicketClaims{
				UserID:     1001,
				Username:   "alice",
				InstanceID: 42,
			},
		},
		nil,
		CookieConfig{},
		recorder,
	)

	router := gin.New()
	router.GET("/api/v1/instances/:id/proxy/*proxyPath", handler.ProxyInstance)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/instances/42/proxy/ping", nil)
	req.AddCookie(&http.Cookie{Name: proxyAccessCookieName, Value: "ticket-1"})
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Fatalf("expected proxied status %d, got %d body=%s", http.StatusCreated, resp.Code, resp.Body.String())
	}
	if body := resp.Body.String(); body != "proxied" {
		t.Fatalf("unexpected proxy body: %s", body)
	}
	if recorder.calls != 1 {
		t.Fatalf("expected traffic recorder called once, got %d", recorder.calls)
	}
}
