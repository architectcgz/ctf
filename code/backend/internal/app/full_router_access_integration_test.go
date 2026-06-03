package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	assessmentcmd "ctf-platform/internal/module/assessment/application/commands"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	fullrouteraccess "ctf-platform/tests/system/http/fullrouteraccess"
)

func TestFullRouter_AccessControlMatrix(t *testing.T) {
	env := newFullRouterTestEnv(t)
	fullrouteraccess.VerifyAccessControlMatrix(t, fullrouteraccess.AccessMatrixDriver{
		Routes: fullRouterAccessRoutes(env),
		Classify: func(route fullrouteraccess.Route) string {
			return string(classifyRouteAccess(route.Method, route.Path))
		},
		Materialize: func(route fullrouteraccess.Route) string {
			return materializeRoutePath(route.Path, env)
		},
		Request: func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
			return performFullRouterRequest(t, env.router, method, target, payload, headers)
		},
		StudentHeaders: func() map[string]string {
			return sessionHeaders(loginForSession(t, env.router, env.student.Username, env.studentPwd))
		},
		TeacherHeaders: func() map[string]string {
			return sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))
		},
	})
}

func TestFullRouter_AuthorizedSmokeMatrix(t *testing.T) {
	baseEnv := newFullRouterTestEnv(t)
	routes := fullRouterAccessRoutes(baseEnv)
	driver := fullrouteraccess.AuthorizedSmokeDriver{
		Classify: func(route fullrouteraccess.Route) string {
			return string(classifyRouteAccess(route.Method, route.Path))
		},
		Request: func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
			panic("request should be bound per subtest")
		},
		AcceptableStatus: func(route fullrouteraccess.Route, status int) bool {
			return isAcceptableSmokeStatus(route.Method, route.Path, status)
		},
	}

	for _, route := range routes {
		route := route
		t.Run(route.Method+" "+route.Path, func(t *testing.T) {
			env := newFullRouterTestEnv(t)
			driver.Request = func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder {
				return performFullRouterRequest(t, env.router, method, target, payload, headers)
			}
			driver.AfterResponse = func(t *testing.T, route fullrouteraccess.Route, resp *httptest.ResponseRecorder, headers map[string]string) {
				if route.Method == http.MethodPost && route.Path == "/api/v1/reports/class" && resp.Code == http.StatusOK {
					var report assessmentcmd.ReportExportData
					decodeFullRouterData(t, resp, &report)
					waitForReportStatus(t, env, report.ReportID, headers, assessmententity.ReportStatusReady, 5*time.Second)
				}
			}

			fullrouteraccess.VerifyAuthorizedSmokeRoute(
				t,
				route,
				materializeRoutePath(route.Path, env),
				routePayload(route.Method, route.Path),
				authorizedHeadersForRoute(t, env, route.Method, route.Path),
				driver,
			)
		})
	}
}

func TestFullRouter_ListInstancesMatchesContract(t *testing.T) {
	env := newFullRouterTestEnv(t)

	headers := sessionHeaders(loginForSession(t, env.router, env.student.Username, env.studentPwd))
	fullrouteraccess.VerifyListInstancesMatchesContract(
		t,
		func(method, target string, payload any, reqHeaders map[string]string) *httptest.ResponseRecorder {
			return performFullRouterRequest(t, env.router, method, target, payload, reqHeaders)
		},
		headers,
		fullrouteraccess.ExpectedInstance{
			ID:               env.instance.ID,
			ChallengeID:      env.challenge.ID,
			ChallengeTitle:   env.challenge.Title,
			Category:         env.challenge.Category,
			Difficulty:       env.challenge.Difficulty,
			FlagType:         env.challenge.FlagType,
			RemainingExtends: env.instance.MaxExtends - env.instance.ExtendCount,
		},
	)
}

func fullRouterAccessRoutes(env *fullRouterTestEnv) []fullrouteraccess.Route {
	routes := filteredRouterRoutes(env.router.Routes())
	items := make([]fullrouteraccess.Route, 0, len(routes))
	for _, route := range routes {
		items = append(items, fullrouteraccess.Route{
			Method: route.Method,
			Path:   route.Path,
		})
	}
	return items
}
