package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/config"
	assessmenthttp "ctf-platform/internal/module/assessment/api/http"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentinfra "ctf-platform/internal/module/assessment/infrastructure"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

type routerChallengeLookupContextStub struct {
	findByIDWithContextFn func(ctx context.Context, id int64) (*challengecontracts.ContestChallenge, error)
}

func (s *routerChallengeLookupContextStub) FindByID(ctx context.Context, id int64) (*challengecontracts.ContestChallenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func TestNewRouterRegistersStudentChallengeRoutes(t *testing.T) {
	cfg, db, cache := newAppTestDependencies(t)

	router, err := NewRouter(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	assertHasRoute(t, router, "GET", "/api/v1/challenges")
	assertHasRoute(t, router, "GET", "/api/v1/challenges/:id")
	assertHasRoute(t, router, "GET", "/.well-known/oauth-protected-resource")
	assertHasRoute(t, router, "GET", "/.well-known/oauth-authorization-server")
	assertHasRoute(t, router, "POST", "/api/v1/oauth/register")
	assertHasRoute(t, router, "GET", "/api/v1/oauth/authorize")
	assertHasRoute(t, router, "POST", "/api/v1/oauth/authorize")
	assertRouteMissing(t, router, "POST", "/api/v1/auth/mcp-token")
	assertHasRoute(t, router, "POST", "/api/v1/contests/:id/challenges/:cid/instances")
	assertHasRoute(t, router, "POST", "/api/v1/contests/:id/awd/services/:sid/instances")
	assertHasRoute(t, router, "POST", "/api/v1/contests/:id/awd/services/:sid/instances/restart")
	assertHasRoute(t, router, "GET", "/api/v1/teacher/instances")
	assertHasRoute(t, router, "DELETE", "/api/v1/teacher/instances/:id")
	assertHasRoute(t, router, "GET", "/api/v1/users/me/progress")
	assertHasRoute(t, router, "GET", "/api/v1/users/me/timeline")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/users/me/progress", "internal/module/practice/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/users/me/timeline", "internal/module/practice/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/audit-logs", "internal/module/ops")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/dashboard", "internal/module/ops")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/cheat-detection", "internal/module/ops")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/admin/notifications", "internal/module/ops")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/users/me/skill-profile", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/users/me/recommendations", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/users/:id/skill-profile", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/reports/personal", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/reports/:id", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/reports/:id/download", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/reports/class", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/students/:id/skill-profile", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/reports/class", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/awd/reviews", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/awd/reviews/:id", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/awd/reviews/:id/export/archive", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/awd/reviews/:id/export/report", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/overview", "internal/module/teaching_analysis/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/students/:id/evidence", "internal/module/teaching_analysis/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/students/:id/attack-sessions", "internal/module/teaching_analysis/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/challenges/:id/writeup-submissions", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/challenges/:id/writeup-submissions/me", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/authoring/challenges/:id/writeup/recommend", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "DELETE", "/api/v1/authoring/challenges/:id/writeup/recommend", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/community-writeups/:id/recommend", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "DELETE", "/api/v1/teacher/community-writeups/:id/recommend", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/community-writeups/:id/hide", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/community-writeups/:id/restore", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/writeup-submissions", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/writeup-submissions/:id", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/challenges", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/challenges/:id", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/authoring/challenges", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/authoring/challenge-imports", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/authoring/challenge-imports", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/authoring/challenge-imports/:id", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/authoring/challenge-imports/:id/commit", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/authoring/challenges/:id/self-check", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/authoring/challenges/:id/publish-requests", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/authoring/challenges/:id/publish-requests/latest", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "PUT", "/api/v1/authoring/challenges/:id/flag", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/authoring/images", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/admin/contests", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/contests/:id", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/contests/:id/awd/services", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/contests/:id/awd/instances", "internal/module/practice/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/admin/contests/:id/awd/instances", "internal/module/practice/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/admin/contests/:id/awd/instances/prewarm", "internal/module/practice/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/admin/contests/:id/awd/services", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "PUT", "/api/v1/admin/contests/:id/awd/services/:sid", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "DELETE", "/api/v1/admin/contests/:id/awd/services/:sid", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/admin/contests/:id/awd/checker-preview", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/contests/:id/awd/readiness", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/contests/:id/awd/rounds/:rid/traffic/summary", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/contests/:id/awd/rounds/:rid/traffic/events", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/contests/:id/scoreboard", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/contests/:id/awd/workspace", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/contests/:id/challenges/:cid/submissions", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/contests/:id/teams", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/challenges/:id/instances", "internal/module/practice/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/contests/:id/challenges/:cid/instances", "internal/module/practice/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/contests/:id/awd/services/:sid/instances", "internal/module/practice/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/challenges/:id/submit", "internal/module/practice/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/challenges/:id/submissions/mine", "internal/module/practice/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/notifications", "internal/module/ops")
	assertRouteHandlerContains(t, router, "PUT", "/api/v1/notifications/:id/read", "internal/module/ops")
	assertRouteHandlerContains(t, router, "GET", "/ws/notifications", "internal/module/ops")
	assertRouteHandlerContains(t, router, "GET", "/ws/contests/:id/announcements", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "GET", "/ws/contests/:id/scoreboard", "internal/module/contest/api/http")
}

func TestChallengeOwnerGuardPropagatesRequestContextToLookup(t *testing.T) {
	t.Parallel()

	type ctxKey string

	const expectedCtxValue = "ctx-owner-guard"
	var called bool

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/authoring/challenges/:id",
		func(c *gin.Context) {
			authctx.SetCurrentUser(c, authctx.CurrentUser{UserID: 42, Role: identitycontracts.RoleTeacher})
			c.Next()
		},
		challengeOwnerGuard(&routerChallengeLookupContextStub{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.ContestChallenge, error) {
				called = true
				if got := ctx.Value(ctxKey("owner-guard")); got != expectedCtxValue {
					t.Fatalf("expected ctx value %v, got %v", expectedCtxValue, got)
				}
				if id != 11 {
					t.Fatalf("unexpected challenge id: %d", id)
				}
				createdBy := int64(42)
				return &challengecontracts.ContestChallenge{ID: id, CreatedBy: &createdBy}, nil
			},
		}),
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/authoring/challenges/11", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxKey("owner-guard"), expectedCtxValue))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected challenge lookup to be called")
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestNewRouterUsesInstanceHandlersForInstanceRoutes(t *testing.T) {
	cfg, db, cache := newAppTestDependencies(t)

	router, err := NewRouter(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	assertRouteHandlerContains(t, router, "GET", "/api/v1/instances", "internal/module/instance/api/http")
	assertRouteHandlerContains(t, router, "DELETE", "/api/v1/instances/:id", "internal/module/instance/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/instances/:id/extend", "internal/module/instance/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/instances/:id/access", "internal/module/instance/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/instances/:id/proxy", "internal/module/instance/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/instances/:id/proxy/*proxyPath", "internal/module/instance/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/instances/:id/proxy/*proxyPath", "internal/module/instance/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/contests/:id/awd/services/:sid/targets/:team_id/access", "internal/module/instance/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/contests/:id/awd/services/:sid/defense/ssh", "internal/module/instance/api/http")
	assertRouteMissing(t, router, "GET", "/api/v1/contests/:id/awd/services/:sid/defense/files")
	assertRouteMissing(t, router, "GET", "/api/v1/contests/:id/awd/services/:sid/defense/directories")
	assertRouteMissing(t, router, "PUT", "/api/v1/contests/:id/awd/services/:sid/defense/files")
	assertRouteMissing(t, router, "POST", "/api/v1/contests/:id/awd/services/:sid/defense/commands")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/contests/:id/awd/services/:sid/targets/:team_id/proxy", "internal/module/instance/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/contests/:id/awd/services/:sid/targets/:team_id/proxy/*proxyPath", "internal/module/instance/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/contests/:id/awd/services/:sid/targets/:team_id/proxy/*proxyPath", "internal/module/instance/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/instances", "internal/module/instance/api/http")
	assertRouteHandlerContains(t, router, "DELETE", "/api/v1/teacher/instances/:id", "internal/module/instance/api/http")
}

func TestNewRouterFailsWhenRemoteRuntimeAgentDialFails(t *testing.T) {
	cfg, db, cache := newAppTestDependencies(t)
	cfg.App.Env = "dev"
	caFile, certFile, keyFile := writeRouterRemoteAgentClientTLSFiles(t)
	cfg.RuntimeAgent = config.RuntimeAgentConfig{
		Enabled:     true,
		Endpoint:    "127.0.0.1:1",
		DialTimeout: 50 * time.Millisecond,
		ServerName:  "runtime-agent.local",
		CAFile:      caFile,
		CertFile:    certFile,
		KeyFile:     keyFile,
	}

	router, err := NewRouter(cfg, zap.NewNop(), db, cache)
	if err == nil {
		t.Fatal("expected NewRouter() to fail when runtime agent dial fails")
	}
	if router != nil {
		t.Fatalf("expected router to be nil on runtime agent dial failure, got %+v", router)
	}
}

func TestTeacherAWDReviewArchiveReqUsesPlannedQueryParams(t *testing.T) {
	t.Parallel()

	reqType := reflect.TypeOf(assessmenthttp.GetTeacherAWDReviewArchiveReq{})

	roundField, ok := reqType.FieldByName("RoundNumber")
	if !ok {
		t.Fatalf("RoundNumber field missing")
	}
	if tag := roundField.Tag.Get("form"); tag != "round" {
		t.Fatalf("RoundNumber form tag = %q, want %q", tag, "round")
	}
	if tag := roundField.Tag.Get("binding"); tag != "omitempty,min=1" {
		t.Fatalf("RoundNumber binding tag = %q, want %q", tag, "omitempty,min=1")
	}

	teamField, ok := reqType.FieldByName("TeamID")
	if !ok {
		t.Fatalf("TeamID field missing")
	}
	if tag := teamField.Tag.Get("form"); tag != "team_id" {
		t.Fatalf("TeamID form tag = %q, want %q", tag, "team_id")
	}
}

func TestTeacherAWDReviewContestQueryUsesPlannedQueryParams(t *testing.T) {
	t.Parallel()

	reqType := reflect.TypeOf(assessmenthttp.TeacherAWDReviewContestQuery{})

	statusField, ok := reqType.FieldByName("Status")
	if !ok || statusField.Tag.Get("form") != "status" {
		t.Fatalf("Status form tag = %q, want %q", statusField.Tag.Get("form"), "status")
	}
	if tag := statusField.Tag.Get("binding"); tag != "omitempty,max=32" {
		t.Fatalf("Status binding tag = %q, want %q", tag, "omitempty,max=32")
	}

	keywordField, ok := reqType.FieldByName("Keyword")
	if !ok || keywordField.Tag.Get("form") != "keyword" {
		t.Fatalf("Keyword form tag = %q, want %q", keywordField.Tag.Get("form"), "keyword")
	}
	if tag := keywordField.Tag.Get("binding"); tag != "omitempty,max=128" {
		t.Fatalf("Keyword binding tag = %q, want %q", tag, "omitempty,max=128")
	}

	pageField, ok := reqType.FieldByName("Page")
	if !ok || pageField.Tag.Get("form") != "page" {
		t.Fatalf("Page form tag = %q, want %q", pageField.Tag.Get("form"), "page")
	}
	if tag := pageField.Tag.Get("binding"); tag != "omitempty,min=1" {
		t.Fatalf("Page binding tag = %q, want %q", tag, "omitempty,min=1")
	}

	sizeField, ok := reqType.FieldByName("Size")
	if !ok || sizeField.Tag.Get("form") != "page_size" {
		t.Fatalf("Size form tag = %q, want %q", sizeField.Tag.Get("form"), "page_size")
	}
	if tag := sizeField.Tag.Get("binding"); tag != "omitempty,min=1,max=100" {
		t.Fatalf("Size binding tag = %q, want %q", tag, "omitempty,min=1,max=100")
	}
}

func TestTeacherAWDReviewServiceInvalidRoundUsesRoundMessage(t *testing.T) {
	t.Parallel()

	db := contesttestsupport.SetupAWDTestDB(t)
	now := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	service := assessmentqry.NewTeacherAWDReviewService(
		assessmentinfra.NewTeacherAWDReviewRepository(db),
		config.PaginationConfig{DefaultPageSize: 20, MaxPageSize: 100},
	)

	contesttestsupport.CreateAWDContestFixture(t, db, 901, now)
	contesttestsupport.CreateAWDRoundFixtureWithWindow(t, db, 90101, 901, 1, 60, 40, now.Add(-40*time.Minute), now.Add(-20*time.Minute))

	_, err := service.GetContestArchive(context.Background(), 1, 901, assessmentqry.GetTeacherAWDReviewArchiveInput{
		RoundNumber: func(v int) *int { return &v }(2),
	})
	if err == nil {
		t.Fatalf("expected invalid round error")
	}

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error, got %T", err)
	}
	if appErr.Message != "round 无效" {
		t.Fatalf("app error message = %q, want %q", appErr.Message, "round 无效")
	}
}
