package fullrouteraccess

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	AccessPublic    = "public"
	AccessProtected = "protected"
	AccessTeacher   = "teacher"
	AccessAdmin     = "admin"
)

type Route struct {
	Method string
	Path   string
}

type RequestFunc func(method, target string, payload any, headers map[string]string) *httptest.ResponseRecorder

type AccessMatrixDriver struct {
	Routes         []Route
	Classify       func(route Route) string
	Materialize    func(route Route) string
	Request        RequestFunc
	StudentHeaders func() map[string]string
	TeacherHeaders func() map[string]string
}

type AuthorizedSmokeDriver struct {
	Classify         func(route Route) string
	Request          RequestFunc
	AcceptableStatus func(route Route, status int) bool
	AfterResponse    func(t *testing.T, route Route, resp *httptest.ResponseRecorder, headers map[string]string)
}

type ExpectedInstance struct {
	ID               int64
	ChallengeID      int64
	ChallengeTitle   string
	Category         string
	Difficulty       string
	FlagType         string
	RemainingExtends int
}

type InstanceListItem struct {
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

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func VerifyAccessControlMatrix(t *testing.T, driver AccessMatrixDriver) {
	t.Helper()

	for _, route := range driver.Routes {
		access := driver.Classify(route)
		if access == AccessPublic {
			continue
		}

		target := driver.Materialize(route)
		resp := driver.Request(route.Method, target, nil, nil)
		if resp.Code != http.StatusUnauthorized {
			t.Errorf("expected unauthorized for %s %s, got %d body=%s", route.Method, route.Path, resp.Code, resp.Body.String())
			continue
		}

		if access == AccessTeacher || access == AccessAdmin {
			resp = driver.Request(route.Method, target, nil, driver.StudentHeaders())
			if resp.Code != http.StatusForbidden {
				t.Errorf("expected forbidden for student on %s %s, got %d body=%s", route.Method, route.Path, resp.Code, resp.Body.String())
			}
		}

		if access == AccessAdmin {
			resp = driver.Request(route.Method, target, nil, driver.TeacherHeaders())
			if resp.Code != http.StatusForbidden {
				t.Errorf("expected forbidden for teacher on %s %s, got %d body=%s", route.Method, route.Path, resp.Code, resp.Body.String())
			}
		}
	}
}

func VerifyAuthorizedSmokeRoute(
	t *testing.T,
	route Route,
	target string,
	payload any,
	headers map[string]string,
	driver AuthorizedSmokeDriver,
) {
	t.Helper()

	resp := driver.Request(route.Method, target, payload, headers)
	if !driver.AcceptableStatus(route, resp.Code) {
		t.Errorf("expected non-5xx for %s %s, got %d body=%s", route.Method, route.Path, resp.Code, resp.Body.String())
		return
	}

	if driver.AfterResponse != nil {
		driver.AfterResponse(t, route, resp, headers)
	}

	access := driver.Classify(route)
	if access != AccessPublic && resp.Code == http.StatusUnauthorized {
		t.Errorf("expected authorized request for %s %s, got %d body=%s", route.Method, route.Path, resp.Code, resp.Body.String())
	}
}

func VerifyListInstancesMatchesContract(t *testing.T, request RequestFunc, headers map[string]string, expected ExpectedInstance) {
	t.Helper()

	resp := request(http.MethodGet, "/api/v1/instances", nil, headers)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var items []InstanceListItem
	decodeEnvelopeData(t, resp, &items)

	if len(items) != 1 {
		t.Fatalf("expected 1 instance, got %+v", items)
	}
	item := items[0]
	if item.ID != expected.ID {
		t.Fatalf("expected instance id %d, got %d", expected.ID, item.ID)
	}
	if item.ChallengeID != expected.ChallengeID {
		t.Fatalf("expected challenge id %d, got %d", expected.ChallengeID, item.ChallengeID)
	}
	if item.ChallengeTitle != expected.ChallengeTitle {
		t.Fatalf("expected challenge title %q, got %q", expected.ChallengeTitle, item.ChallengeTitle)
	}
	if item.Category != expected.Category {
		t.Fatalf("expected category %q, got %q", expected.Category, item.Category)
	}
	if item.Difficulty != expected.Difficulty {
		t.Fatalf("expected difficulty %q, got %q", expected.Difficulty, item.Difficulty)
	}
	if item.FlagType != expected.FlagType {
		t.Fatalf("expected flag type %q, got %q", expected.FlagType, item.FlagType)
	}
	if item.RemainingExtends != expected.RemainingExtends {
		t.Fatalf("expected remaining_extends %d, got %d", expected.RemainingExtends, item.RemainingExtends)
	}
}

func decodeEnvelopeData(t *testing.T, resp *httptest.ResponseRecorder, target any) {
	t.Helper()

	var body envelope
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response envelope: %v body=%s", err, resp.Body.String())
	}
	if len(body.Data) == 0 || string(body.Data) == "null" {
		t.Fatalf("expected response data, got empty body=%s", resp.Body.String())
	}
	if err := json.Unmarshal(body.Data, target); err != nil {
		t.Fatalf("decode response data: %v body=%s", err, resp.Body.String())
	}
}
