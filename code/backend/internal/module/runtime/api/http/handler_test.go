package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
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

func (stubRuntimeService) IssueAWDTargetProxyTicket(context.Context, authctx.CurrentUser, int64, int64, int64) (string, error) {
	return "", nil
}

func (stubRuntimeService) ResolveProxyTicket(context.Context, string) (*runtimeports.ProxyTicketClaims, error) {
	return nil, nil
}

func (stubRuntimeService) ResolveAWDTargetAccessURL(context.Context, *runtimeports.ProxyTicketClaims, int64, int64, int64) (string, error) {
	return "", nil
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

type stubAWDProxyRuntimeService struct {
	stubRuntimeService
	issuedTicket string
	targetURL    string
	claims       *runtimeports.ProxyTicketClaims
}

func (s stubAWDProxyRuntimeService) IssueAWDTargetProxyTicket(context.Context, authctx.CurrentUser, int64, int64, int64) (string, error) {
	return s.issuedTicket, nil
}

func (s stubAWDProxyRuntimeService) ResolveProxyTicket(context.Context, string) (*runtimeports.ProxyTicketClaims, error) {
	return s.claims, nil
}

func (s stubAWDProxyRuntimeService) ResolveAWDTargetAccessURL(context.Context, *runtimeports.ProxyTicketClaims, int64, int64, int64) (string, error) {
	return s.targetURL, nil
}

type failingTrafficRecorder struct {
	calls int
}

func (r *failingTrafficRecorder) RecordRuntimeProxyTrafficEvent(context.Context, int64, int64, string, string, int) error {
	r.calls++
	return errors.New("persist failed")
}

func (r *failingTrafficRecorder) RecordAWDProxyTrafficEvent(context.Context, model.AWDProxyTrafficEventInput) error {
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

func TestAccessAWDTargetReturnsTargetProxyURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(
		stubAWDProxyRuntimeService{issuedTicket: "ticket-awd"},
		nil,
		CookieConfig{},
		nil,
	)

	router := gin.New()
	router.POST("/api/v1/contests/:id/awd/services/:sid/targets/:team_id/access", func(c *gin.Context) {
		c.Set("current_user", authctx.CurrentUser{UserID: 1001, Username: "alice", Role: model.RoleStudent})
		c.Set("id", int64(7))
		c.Set("sid", int64(7009))
		c.Set("team_id", int64(14))
		handler.AccessAWDTarget(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/contests/7/awd/services/7009/targets/14/access", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Body.String(), "/api/v1/contests/7/awd/services/7009/targets/14/proxy/?ticket=ticket-awd") {
		t.Fatalf("expected awd proxy url in response, got %s", resp.Body.String())
	}
}

func TestProxyAWDTargetForwardsAndRecordsExplicitTrafficScope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/flag" {
			t.Fatalf("unexpected target path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("flag page"))
	}))
	defer target.Close()

	contestID := int64(7)
	attackerTeamID := int64(13)
	victimTeamID := int64(14)
	serviceID := int64(7009)
	challengeID := int64(9)
	recorder := &recordingProxyTrafficRecorder{}
	handler := NewHandler(
		stubAWDProxyRuntimeService{
			targetURL: target.URL,
			claims: &runtimeports.ProxyTicketClaims{
				UserID:            1001,
				Username:          "alice",
				Role:              model.RoleStudent,
				InstanceID:        42,
				ContestID:         &contestID,
				Purpose:           runtimeports.ProxyTicketPurposeAWDAttack,
				ShareScope:        model.InstanceSharingPerTeam,
				AWDAttackerTeamID: &attackerTeamID,
				AWDVictimTeamID:   &victimTeamID,
				AWDServiceID:      &serviceID,
				AWDChallengeID:    &challengeID,
			},
		},
		nil,
		CookieConfig{},
		recorder,
	)

	router := gin.New()
	router.GET("/api/v1/contests/:id/awd/services/:sid/targets/:team_id/proxy/*proxyPath", func(c *gin.Context) {
		c.Set("id", contestID)
		c.Set("sid", serviceID)
		c.Set("team_id", victimTeamID)
		handler.ProxyAWDTarget(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/contests/7/awd/services/7009/targets/14/proxy/api/flag", nil)
	req.AddCookie(&http.Cookie{Name: proxyAccessCookieName, Value: "ticket-awd"})
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected proxied status 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	if resp.Body.String() != "flag page" {
		t.Fatalf("unexpected proxy body: %s", resp.Body.String())
	}
	if recorder.awdEvent == nil {
		t.Fatal("expected explicit awd traffic event")
	}
	if recorder.awdEvent.AttackerTeamID != attackerTeamID || recorder.awdEvent.VictimTeamID != victimTeamID || recorder.awdEvent.ServiceID != serviceID {
		t.Fatalf("unexpected awd event: %+v", recorder.awdEvent)
	}
}
