package app

import (
	"net/http"
	"testing"
	"time"

	assessmentcmd "ctf-platform/internal/module/assessment/application/commands"
	assessmententity "ctf-platform/internal/module/assessment/entity"
)

func TestFullRouter_AccessControlMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)

	for _, route := range filteredRouterRoutes(env.router.Routes()) {
		access := classifyRouteAccess(route.Method, route.Path)
		if access == routeAccessPublic {
			continue
		}

		target := materializeRoutePath(route.Path, env)
		resp := performFullRouterRequest(t, env.router, route.Method, target, nil, nil)
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("expected unauthorized for %s %s, got %d body=%s", route.Method, route.Path, resp.Code, resp.Body.String())
			continue
		}

		if access == routeAccessTeacher || access == routeAccessAdmin {
			studentHeaders := sessionHeaders(loginForSession(t, env.router, env.student.Username, env.studentPwd))
			resp = performFullRouterRequest(t, env.router, route.Method, target, nil, studentHeaders)
			if resp.Code != http.StatusForbidden {
				t.Errorf("expected forbidden for student on %s %s, got %d body=%s", route.Method, route.Path, resp.Code, resp.Body.String())
			}
		}

		if access == routeAccessAdmin {
			teacherHeaders := sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))
			resp = performFullRouterRequest(t, env.router, route.Method, target, nil, teacherHeaders)
			if resp.Code != http.StatusForbidden {
				t.Errorf("expected forbidden for teacher on %s %s, got %d body=%s", route.Method, route.Path, resp.Code, resp.Body.String())
			}
		}
	}
}

func TestFullRouter_AuthorizedSmokeMatrix(t *testing.T) {
	baseEnv := newFullRouterTestEnv(t)

	for _, route := range filteredRouterRoutes(baseEnv.router.Routes()) {
		route := route
		t.Run(route.Method+" "+route.Path, func(t *testing.T) {
			env := newFullRouterTestEnv(t)
			target := materializeRoutePath(route.Path, env)
			headers := authorizedHeadersForRoute(t, env, route.Method, route.Path)
			payload := routePayload(route.Method, route.Path)

			resp := performFullRouterRequest(t, env.router, route.Method, target, payload, headers)
			if !isAcceptableSmokeStatus(route.Method, route.Path, resp.Code) {
				t.Errorf("expected non-5xx for %s %s, got %d body=%s", route.Method, route.Path, resp.Code, resp.Body.String())
				return
			}

			if route.Method == http.MethodPost && route.Path == "/api/v1/reports/class" && resp.Code == http.StatusOK {
				var report assessmentcmd.ReportExportData
				decodeFullRouterData(t, resp, &report)
				waitForReportStatus(t, env, report.ReportID, headers, assessmententity.ReportStatusReady, 5*time.Second)
			}

			access := classifyRouteAccess(route.Method, route.Path)
			if access != routeAccessPublic && resp.Code == http.StatusUnauthorized {
				t.Errorf("expected authorized request for %s %s, got %d body=%s", route.Method, route.Path, resp.Code, resp.Body.String())
			}
		})
	}
}

func TestFullRouter_ListInstancesMatchesContract(t *testing.T) {
	env := newFullRouterTestEnv(t)

	headers := sessionHeaders(loginForSession(t, env.router, env.student.Username, env.studentPwd))
	resp := performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/instances", nil, headers)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var items []struct {
		ID               int64     `json:"id"`
		ChallengeID      int64     `json:"challenge_id"`
		ChallengeTitle   string    `json:"challenge_title"`
		Category         string    `json:"category"`
		Difficulty       string    `json:"difficulty"`
		FlagType         string    `json:"flag_type"`
		Status           string    `json:"status"`
		AccessURL        string    `json:"access_url"`
		ExpiresAt        time.Time `json:"expires_at"`
		RemainingTime    int64     `json:"remaining_time"`
		RemainingExtends int       `json:"remaining_extends"`
	}
	decodeFullRouterData(t, resp, &items)

	if len(items) != 1 {
		t.Fatalf("expected 1 instance, got %+v", items)
	}
	item := items[0]
	if item.ID != env.instance.ID {
		t.Fatalf("expected instance id %d, got %d", env.instance.ID, item.ID)
	}
	if item.ChallengeID != env.challenge.ID {
		t.Fatalf("expected challenge id %d, got %d", env.challenge.ID, item.ChallengeID)
	}
	if item.ChallengeTitle != env.challenge.Title {
		t.Fatalf("expected challenge title %q, got %q", env.challenge.Title, item.ChallengeTitle)
	}
	if item.Category != env.challenge.Category {
		t.Fatalf("expected category %q, got %q", env.challenge.Category, item.Category)
	}
	if item.Difficulty != env.challenge.Difficulty {
		t.Fatalf("expected difficulty %q, got %q", env.challenge.Difficulty, item.Difficulty)
	}
	if item.FlagType != env.challenge.FlagType {
		t.Fatalf("expected flag type %q, got %q", env.challenge.FlagType, item.FlagType)
	}
	if item.RemainingExtends != env.instance.MaxExtends-env.instance.ExtendCount {
		t.Fatalf("expected remaining_extends %d, got %d", env.instance.MaxExtends-env.instance.ExtendCount, item.RemainingExtends)
	}
}
