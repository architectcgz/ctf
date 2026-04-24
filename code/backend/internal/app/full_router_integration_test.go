package app

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	flagcrypto "ctf-platform/pkg/crypto"
)

type fullRouterTestEnv struct {
	router *gin.Engine
	db     *gorm.DB
	cache  *redislib.Client

	admin        *model.User
	teacher      *model.User
	student      *model.User
	peerStudent  *model.User
	otherTeacher *model.User
	otherStudent *model.User
	studentPwd   string
	teacherPwd   string
	adminPwd     string
	className    string
	reportDir    string
	image        *model.Image
	challenge    *model.Challenge
	template     *model.EnvironmentTemplate
	contest      *model.Contest
	awdContest   *model.Contest
	registration *model.ContestRegistration
	announcement *model.ContestAnnouncement
	team         *model.Team
	awdRound     *model.AWDRound
	instance     *model.Instance
	notification *model.Notification
	report       *model.Report
}

var (
	fullRouterSchemaTemplateOnce sync.Once
	fullRouterSchemaTemplatePath string
	fullRouterSchemaTemplateErr  error
)

var fullRouterTestSchemaModels = []any{
	&model.Role{},
	&model.User{},
	&model.UserRole{},
	&model.Image{},
	&model.Challenge{},
	&model.AWDServiceTemplate{},
	&model.ChallengePublishCheckJob{},
	&model.Tag{},
	&model.ChallengeTag{},
	&model.ChallengeHint{},
	&model.ChallengeWriteup{},
	&model.SubmissionWriteup{},
	&model.EnvironmentTemplate{},
	&model.ChallengeTopology{},
	&model.Submission{},
	&model.Instance{},
	&model.PortAllocation{},
	&model.UserScore{},
	&model.AuditLog{},
	&model.NotificationBatch{},
	&model.Notification{},
	&model.SkillProfile{},
	&model.Contest{},
	&model.ContestChallenge{},
	&model.ContestAWDService{},
	&model.ContestRegistration{},
	&model.ContestAnnouncement{},
	&model.Team{},
	&model.TeamMember{},
	&model.AWDRound{},
	&model.AWDTeamService{},
	&model.AWDAttackLog{},
	&model.AWDTrafficEvent{},
	&model.Report{},
}

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
				var report dto.ReportExportData
				decodeFullRouterData(t, resp, &report)
				waitForReportStatus(t, env, report.ReportID, headers, model.ReportStatusReady, 5*time.Second)
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

func TestFullRouter_TeacherCanOnlyManageOwnChallenges(t *testing.T) {
	env := newFullRouterTestEnv(t)

	adminHeaders := sessionHeaders(loginForSession(t, env.router, env.admin.Username, env.adminPwd))
	teacherHeaders := sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))

	createPayload := func(title string) map[string]any {
		return map[string]any{
			"title":       title,
			"description": "ownership test challenge",
			"category":    model.DimensionWeb,
			"difficulty":  model.ChallengeDifficultyEasy,
			"points":      100,
			"image_id":    env.image.ID,
		}
	}

	adminCreateResp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/challenges", createPayload("admin-owned"), adminHeaders)
	assertFullRouterStatus(t, adminCreateResp, http.StatusOK)
	var adminChallenge dto.ChallengeResp
	decodeFullRouterData(t, adminCreateResp, &adminChallenge)

	teacherCreateResp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/challenges", createPayload("teacher-owned"), teacherHeaders)
	assertFullRouterStatus(t, teacherCreateResp, http.StatusOK)
	var teacherChallenge dto.ChallengeResp
	decodeFullRouterData(t, teacherCreateResp, &teacherChallenge)

	listResp := performFullRouterRequest(t, env.router, http.MethodGet, "/api/v1/authoring/challenges?page=1&page_size=50", nil, teacherHeaders)
	assertFullRouterStatus(t, listResp, http.StatusOK)
	var listResult struct {
		List []dto.ChallengeResp `json:"list"`
	}
	decodeFullRouterData(t, listResp, &listResult)

	foundTeacherOwned := false
	foundAdminOwned := false
	for _, item := range listResult.List {
		if item.ID == teacherChallenge.ID {
			foundTeacherOwned = true
		}
		if item.ID == adminChallenge.ID {
			foundAdminOwned = true
		}
	}
	if !foundTeacherOwned {
		t.Fatalf("teacher should see own challenge %d in list, got %+v", teacherChallenge.ID, listResult.List)
	}
	if foundAdminOwned {
		t.Fatalf("teacher should not see admin challenge %d in list, got %+v", adminChallenge.ID, listResult.List)
	}

	for _, tc := range []struct {
		name    string
		method  string
		path    string
		payload any
	}{
		{name: "get detail", method: http.MethodGet, path: fmt.Sprintf("/api/v1/authoring/challenges/%d", adminChallenge.ID)},
		{name: "update challenge", method: http.MethodPut, path: fmt.Sprintf("/api/v1/authoring/challenges/%d", adminChallenge.ID), payload: map[string]any{"title": "forbidden-update"}},
		{name: "configure flag", method: http.MethodPut, path: fmt.Sprintf("/api/v1/authoring/challenges/%d/flag", adminChallenge.ID), payload: map[string]any{
			"flag_type":   model.FlagTypeStatic,
			"flag":        "flag{ownership-check}",
			"flag_prefix": "flag",
		}},
		{name: "upsert writeup", method: http.MethodPut, path: fmt.Sprintf("/api/v1/authoring/challenges/%d/writeup", adminChallenge.ID), payload: map[string]any{
			"title":      "forbidden writeup",
			"content":    "forbidden content",
			"visibility": model.WriteupVisibilityPublic,
		}},
		{name: "save topology", method: http.MethodPut, path: fmt.Sprintf("/api/v1/authoring/challenges/%d/topology", adminChallenge.ID), payload: map[string]any{
			"entry_node_key": "web",
			"nodes": []map[string]any{
				{
					"key":          "web",
					"name":         "web",
					"image_id":     env.image.ID,
					"service_port": 80,
					"inject_flag":  true,
					"tier":         model.TopologyTierPublic,
				},
			},
		}},
		{name: "self check", method: http.MethodPost, path: fmt.Sprintf("/api/v1/authoring/challenges/%d/self-check", adminChallenge.ID)},
	} {
		resp := performFullRouterRequest(t, env.router, tc.method, tc.path, tc.payload, teacherHeaders)
		if resp.Code != http.StatusForbidden {
			t.Fatalf("%s should be forbidden, got status=%d body=%s", tc.name, resp.Code, resp.Body.String())
		}
	}

	ownDetailResp := performFullRouterRequest(t, env.router, http.MethodGet, fmt.Sprintf("/api/v1/authoring/challenges/%d", teacherChallenge.ID), nil, teacherHeaders)
	assertFullRouterStatus(t, ownDetailResp, http.StatusOK)
}

func TestFullRouter_CreateChallengeStoresCreator(t *testing.T) {
	env := newFullRouterTestEnv(t)

	teacherHeaders := sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))
	resp := performFullRouterRequest(t, env.router, http.MethodPost, "/api/v1/authoring/challenges", map[string]any{
		"title":       "creator-marker",
		"description": "creator marker challenge",
		"category":    model.DimensionWeb,
		"difficulty":  model.ChallengeDifficultyEasy,
		"points":      100,
		"image_id":    env.image.ID,
	}, teacherHeaders)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var challengeData map[string]any
	decodeFullRouterData(t, resp, &challengeData)
	if _, ok := challengeData["created_by"]; !ok {
		t.Fatalf("expected challenge response to include created_by, got %+v", challengeData)
	}

	challengeIDFloat, ok := challengeData["id"].(float64)
	if !ok {
		t.Fatalf("expected numeric challenge id, got %+v", challengeData["id"])
	}

	var createdBy sql.NullInt64
	if err := env.db.Raw("SELECT created_by FROM challenges WHERE id = ?", int64(challengeIDFloat)).Scan(&createdBy).Error; err != nil {
		t.Fatalf("query challenge created_by: %v", err)
	}
	if !createdBy.Valid || createdBy.Int64 != env.teacher.ID {
		t.Fatalf("unexpected created_by=%+v, want %d", createdBy, env.teacher.ID)
	}
}

func TestFullRouter_ChallengeSelfCheckRunsPrecheckAndRuntime(t *testing.T) {
	env := newFullRouterTestEnv(t)

	teacherHeaders := sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))
	resp := performFullRouterRequest(
		t,
		env.router,
		http.MethodPost,
		fmt.Sprintf("/api/v1/authoring/challenges/%d/self-check", env.challenge.ID),
		nil,
		teacherHeaders,
	)
	assertFullRouterStatus(t, resp, http.StatusOK)

	var result dto.ChallengeSelfCheckResp
	decodeFullRouterData(t, resp, &result)
	if result.ChallengeID != env.challenge.ID {
		t.Fatalf("expected challenge_id=%d, got %d", env.challenge.ID, result.ChallengeID)
	}
	if !result.Precheck.Passed {
		t.Fatalf("expected precheck passed, got %+v", result.Precheck)
	}
	if !result.Runtime.Passed {
		t.Fatalf("expected runtime passed, got %+v", result.Runtime)
	}
	if result.Runtime.AccessURL == "" {
		t.Fatalf("expected runtime access url, got empty")
	}
	if len(result.Runtime.Steps) == 0 {
		t.Fatalf("expected runtime steps, got empty")
	}
}

func TestFullRouter_AdminChallengePublishRequestLifecycle(t *testing.T) {
	env := newFullRouterTestEnv(t)
	if err := env.db.Model(&model.Challenge{}).
		Where("id = ?", env.challenge.ID).
		Update("status", model.ChallengeStatusDraft).Error; err != nil {
		t.Fatalf("set challenge draft: %v", err)
	}
	env.challenge.Status = model.ChallengeStatusDraft

	teacherHeaders := sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))
	createResp := performFullRouterRequest(
		t,
		env.router,
		http.MethodPost,
		fmt.Sprintf("/api/v1/authoring/challenges/%d/publish-requests", env.challenge.ID),
		nil,
		teacherHeaders,
	)
	assertFullRouterStatus(t, createResp, http.StatusAccepted)

	var created map[string]any
	decodeFullRouterData(t, createResp, &created)
	if created["challenge_id"] != float64(env.challenge.ID) {
		t.Fatalf("unexpected created publish request payload: %+v", created)
	}
	if created["status"] != "queued" {
		t.Fatalf("expected queued publish request, got %+v", created)
	}
	if created["active"] != true {
		t.Fatalf("expected active publish request, got %+v", created)
	}

	latestResp := performFullRouterRequest(
		t,
		env.router,
		http.MethodGet,
		fmt.Sprintf("/api/v1/authoring/challenges/%d/publish-requests/latest", env.challenge.ID),
		nil,
		teacherHeaders,
	)
	assertFullRouterStatus(t, latestResp, http.StatusOK)

	var latest map[string]any
	decodeFullRouterData(t, latestResp, &latest)
	if latest["id"] != created["id"] {
		t.Fatalf("expected latest publish request id %v, got %+v", created["id"], latest)
	}
	if latest["status"] != "queued" {
		t.Fatalf("expected latest queued publish request, got %+v", latest)
	}
}

func TestRouterBuildUsesCompositionModules(t *testing.T) {
	cfg, db, cache := newAppTestDependencies(t)

	var calls []string

	originalBuildRuntimeModule := buildRuntimeModule
	originalBuildOpsModule := buildOpsModule
	originalBuildIdentityModule := buildIdentityModule
	originalBuildAuthModule := buildAuthModule
	originalBuildChallengeModule := buildChallengeModule
	originalBuildAssessmentModule := buildAssessmentModule
	originalBuildTeachingReadmodelModule := buildTeachingReadmodelModule
	originalBuildContestModule := buildContestModule
	originalBuildPracticeModule := buildPracticeModule
	originalBuildPracticeReadmodelModule := buildPracticeReadmodelModule
	defer func() {
		buildRuntimeModule = originalBuildRuntimeModule
		buildOpsModule = originalBuildOpsModule
		buildIdentityModule = originalBuildIdentityModule
		buildAuthModule = originalBuildAuthModule
		buildChallengeModule = originalBuildChallengeModule
		buildAssessmentModule = originalBuildAssessmentModule
		buildTeachingReadmodelModule = originalBuildTeachingReadmodelModule
		buildContestModule = originalBuildContestModule
		buildPracticeModule = originalBuildPracticeModule
		buildPracticeReadmodelModule = originalBuildPracticeReadmodelModule
	}()

	buildRuntimeModule = func(root *composition.Root) *composition.RuntimeModule {
		if root == nil {
			t.Fatal("expected root for runtime module builder")
		}
		calls = append(calls, "runtime")
		return originalBuildRuntimeModule(root)
	}
	buildOpsModule = func(root *composition.Root, runtime *composition.RuntimeModule) *composition.OpsModule {
		if root == nil || runtime == nil {
			t.Fatal("expected root and runtime for ops module builder")
		}
		calls = append(calls, "ops")
		return originalBuildOpsModule(root, runtime)
	}
	buildIdentityModule = func(root *composition.Root) (*composition.IdentityModule, error) {
		if root == nil {
			t.Fatal("expected root for identity module builder")
		}
		calls = append(calls, "identity")
		return originalBuildIdentityModule(root)
	}
	buildAuthModule = func(root *composition.Root, ops *composition.OpsModule, identity *composition.IdentityModule) (*composition.AuthModule, error) {
		if root == nil || ops == nil || identity == nil {
			t.Fatal("expected root, ops and identity for auth module builder")
		}
		calls = append(calls, "auth")
		return originalBuildAuthModule(root, ops, identity)
	}
	buildChallengeModule = func(root *composition.Root, runtime *composition.RuntimeModule, ops *composition.OpsModule) (*composition.ChallengeModule, error) {
		if root == nil || runtime == nil || ops == nil {
			t.Fatal("expected root, runtime and ops for challenge module builder")
		}
		calls = append(calls, "challenge")
		return originalBuildChallengeModule(root, runtime, ops)
	}
	buildAssessmentModule = func(root *composition.Root, challenge *composition.ChallengeModule) *composition.AssessmentModule {
		if root == nil || challenge == nil {
			t.Fatal("expected root and challenge for assessment module builder")
		}
		calls = append(calls, "assessment")
		return originalBuildAssessmentModule(root, challenge)
	}
	buildTeachingReadmodelModule = func(root *composition.Root, assessment *composition.AssessmentModule) *composition.TeachingReadmodelModule {
		if root == nil || assessment == nil {
			t.Fatal("expected root and assessment for teaching readmodel module builder")
		}
		calls = append(calls, "teaching_readmodel")
		return originalBuildTeachingReadmodelModule(root, assessment)
	}
	buildContestModule = func(root *composition.Root, challenge *composition.ChallengeModule, runtime *composition.RuntimeModule) *composition.ContestModule {
		if root == nil || challenge == nil || runtime == nil {
			t.Fatal("expected root, challenge and runtime for contest module builder")
		}
		calls = append(calls, "contest")
		return originalBuildContestModule(root, challenge, runtime)
	}
	buildPracticeModule = func(root *composition.Root, challenge *composition.ChallengeModule, runtime *composition.RuntimeModule, assessment *composition.AssessmentModule) *composition.PracticeModule {
		if root == nil || challenge == nil || runtime == nil || assessment == nil {
			t.Fatal("expected root, challenge, runtime, and assessment for practice module builder")
		}
		calls = append(calls, "practice")
		return originalBuildPracticeModule(root, challenge, runtime, assessment)
	}
	buildPracticeReadmodelModule = func(root *composition.Root) *composition.PracticeReadmodelModule {
		if root == nil {
			t.Fatal("expected root for practice readmodel module builder")
		}
		calls = append(calls, "practice_readmodel")
		return originalBuildPracticeReadmodelModule(root)
	}

	router, err := NewRouter(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
	if router == nil {
		t.Fatal("expected router")
	}

	expectedCalls := []string{"runtime", "ops", "identity", "auth", "challenge", "assessment", "teaching_readmodel", "contest", "practice", "practice_readmodel"}
	if len(calls) != len(expectedCalls) {
		t.Fatalf("expected %d module builder calls, got %d (%v)", len(expectedCalls), len(calls), calls)
	}
	for i, expected := range expectedCalls {
		if calls[i] != expected {
			t.Fatalf("expected builder call %d to be %q, got %q (%v)", i, expected, calls[i], calls)
		}
	}
}

func isAcceptableSmokeStatus(method, path string, status int) bool {
	if status < http.StatusInternalServerError {
		return true
	}
	if method == http.MethodGet && (path == "/api/v1/auth/cas/login" || path == "/api/v1/auth/cas/callback") && status == http.StatusServiceUnavailable {
		return true
	}
	return false
}

type routeAccessLevel string

const (
	routeAccessPublic    routeAccessLevel = "public"
	routeAccessProtected routeAccessLevel = "protected"
	routeAccessTeacher   routeAccessLevel = "teacher"
	routeAccessAdmin     routeAccessLevel = "admin"
)

func filteredRouterRoutes(routes gin.RoutesInfo) gin.RoutesInfo {
	filtered := make(gin.RoutesInfo, 0, len(routes))
	for _, route := range routes {
		if strings.HasPrefix(route.Path, "/ws/") {
			continue
		}
		if route.Path == "/favicon.ico" {
			continue
		}
		filtered = append(filtered, route)
	}
	return filtered
}

func classifyRouteAccess(method, path string) routeAccessLevel {
	if isPublicRoute(method, path) {
		return routeAccessPublic
	}
	if strings.HasPrefix(path, "/api/v1/admin") {
		if isTeacherAuthoringAdminRoute(path) {
			return routeAccessTeacher
		}
		return routeAccessAdmin
	}
	if strings.HasPrefix(path, "/api/v1/teacher") {
		return routeAccessTeacher
	}
	if path == "/api/v1/users/:id/skill-profile" || path == "/api/v1/reports/class" {
		return routeAccessTeacher
	}
	return routeAccessProtected
}

func isTeacherAuthoringAdminRoute(path string) bool {
	return strings.HasPrefix(path, "/api/v1/authoring/challenges") ||
		strings.HasPrefix(path, "/api/v1/authoring/images") ||
		strings.HasPrefix(path, "/api/v1/authoring/environment-templates")
}

func isPublicRoute(method, path string) bool {
	switch path {
	case "/health", "/health/db", "/health/redis",
		"/api/v1/health", "/api/v1/health/db", "/api/v1/health/redis",
		"/api/v1/auth/register", "/api/v1/auth/login",
		"/api/v1/auth/cas/status", "/api/v1/auth/cas/login", "/api/v1/auth/cas/callback",
		"/ws/notifications",
		"/ws/contests/:id/announcements", "/ws/contests/:id/scoreboard",
		"/api/v1/contests", "/api/v1/contests/:id", "/api/v1/contests/:id/scoreboard", "/api/v1/contests/:id/announcements",
		"/api/v1/instances/:id/proxy", "/api/v1/instances/:id/proxy/*proxyPath":
		return true
	}
	return false
}

func authorizedHeadersForRoute(t *testing.T, env *fullRouterTestEnv, method, path string) map[string]string {
	t.Helper()

	switch classifyRouteAccess(method, path) {
	case routeAccessAdmin:
		return sessionHeaders(loginForSession(t, env.router, env.admin.Username, env.adminPwd))
	case routeAccessTeacher:
		return sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))
	case routeAccessProtected:
		return sessionHeaders(loginForSession(t, env.router, env.student.Username, env.studentPwd))
	default:
		return nil
	}
}

func routePayload(method, path string) any {
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		if strings.HasPrefix(path, "/api/v1/auth/register") {
			return map[string]any{
				"username": "matrix_user",
				"password": "Password123",
			}
		}
		if strings.HasPrefix(path, "/api/v1/auth/login") {
			return map[string]any{
				"username": "matrix_user",
				"password": "Password123",
			}
		}
		return map[string]any{}
	}
	return nil
}

func materializeRoutePath(path string, env *fullRouterTestEnv) string {
	target := path

	switch {
	case strings.Contains(path, "/api/v1/authoring/images/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.image.ID, 10))
	case strings.Contains(path, "/api/v1/authoring/challenges/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.challenge.ID, 10))
	case strings.Contains(path, "/api/v1/authoring/environment-templates/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.template.ID, 10))
	case strings.Contains(path, "/api/v1/admin/users/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.student.ID, 10))
	case strings.Contains(path, "/api/v1/admin/contests/:id/awd/rounds/:rid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.awdContest.ID, 10))
		target = strings.ReplaceAll(target, ":rid", strconv.FormatInt(env.awdRound.ID, 10))
	case strings.Contains(path, "/api/v1/admin/contests/:id/registrations/:rid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
		target = strings.ReplaceAll(target, ":rid", strconv.FormatInt(env.registration.ID, 10))
	case strings.Contains(path, "/api/v1/admin/contests/:id/announcements/:aid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
		target = strings.ReplaceAll(target, ":aid", strconv.FormatInt(env.announcement.ID, 10))
	case strings.Contains(path, "/api/v1/admin/contests/:id/challenges/:cid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
		target = strings.ReplaceAll(target, ":cid", strconv.FormatInt(env.challenge.ID, 10))
	case strings.Contains(path, "/api/v1/admin/contests/:id/scoreboard/live"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.awdContest.ID, 10))
	case strings.Contains(path, "/api/v1/admin/contests/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
	case strings.Contains(path, "/api/v1/teacher/instances/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.instance.ID, 10))
	case strings.Contains(path, "/api/v1/teacher/students/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.student.ID, 10))
	case strings.Contains(path, "/api/v1/teacher/classes/:name"):
		target = strings.ReplaceAll(target, ":name", env.className)
	case strings.Contains(path, "/api/v1/notifications/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.notification.ID, 10))
	case strings.Contains(path, "/api/v1/reports/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.report.ID, 10))
	case strings.Contains(path, "/api/v1/users/:id/skill-profile"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.student.ID, 10))
	case strings.Contains(path, "/api/v1/contests/:id/awd/challenges/:cid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.awdContest.ID, 10))
		target = strings.ReplaceAll(target, ":cid", strconv.FormatInt(env.challenge.ID, 10))
	case strings.Contains(path, "/api/v1/contests/:id/teams/:tid/members/:uid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
		target = strings.ReplaceAll(target, ":tid", strconv.FormatInt(env.team.ID, 10))
		target = strings.ReplaceAll(target, ":uid", strconv.FormatInt(env.student.ID, 10))
	case strings.Contains(path, "/api/v1/contests/:id/teams/:tid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
		target = strings.ReplaceAll(target, ":tid", strconv.FormatInt(env.team.ID, 10))
	case strings.Contains(path, "/api/v1/contests/:id/challenges/:cid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
		target = strings.ReplaceAll(target, ":cid", strconv.FormatInt(env.challenge.ID, 10))
	case strings.Contains(path, "/api/v1/contests/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
	case strings.Contains(path, "/api/v1/challenges/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.challenge.ID, 10))
	case strings.Contains(path, "/api/v1/instances/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.instance.ID, 10))
	}

	target = strings.ReplaceAll(target, ":level", "1")
	target = strings.ReplaceAll(target, "*proxyPath", "sample")
	return target
}

func newFullRouterTestEnv(t *testing.T) *fullRouterTestEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	cache := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() { _ = cache.Close() })

	db := openFullRouterTestDB(t)
	t.Cleanup(func() {
		if sqlDB, sqlErr := db.DB(); sqlErr == nil {
			_ = sqlDB.Close()
		}
	})

	cfg := newFullRouterTestConfig(t)
	router, err := NewRouter(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("new router: %v", err)
	}

	env := &fullRouterTestEnv{
		router:     router,
		db:         db,
		cache:      cache,
		adminPwd:   "Password123",
		teacherPwd: "Password123",
		studentPwd: "Password123",
		className:  "ClassA",
		reportDir:  cfg.Report.StorageDir,
	}

	seedFullRouterData(t, env)
	return env
}

func openFullRouterTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	templatePath, err := ensureFullRouterSchemaTemplate()
	if err != nil {
		t.Fatalf("prepare full router schema template: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "full-router.sqlite")
	if err := copySQLiteTemplate(templatePath, dbPath); err != nil {
		t.Fatalf("clone sqlite schema template: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func ensureFullRouterSchemaTemplate() (string, error) {
	fullRouterSchemaTemplateOnce.Do(func() {
		dir, err := os.MkdirTemp("", "full-router-schema-*")
		if err != nil {
			fullRouterSchemaTemplateErr = fmt.Errorf("create schema temp dir: %w", err)
			return
		}

		fullRouterSchemaTemplatePath = filepath.Join(dir, "schema.sqlite")
		fullRouterSchemaTemplateErr = buildFullRouterSchemaTemplate(fullRouterSchemaTemplatePath)
	})

	if fullRouterSchemaTemplateErr != nil {
		return "", fullRouterSchemaTemplateErr
	}
	return fullRouterSchemaTemplatePath, nil
}

func buildFullRouterSchemaTemplate(path string) error {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("open sqlite template: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sqlite template handle: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	if err := db.AutoMigrate(fullRouterTestSchemaModels...); err != nil {
		return fmt.Errorf("auto migrate sqlite template: %w", err)
	}

	return nil
}

func copySQLiteTemplate(srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open sqlite template: %w", err)
	}
	defer func() { _ = src.Close() }()

	dst, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create sqlite copy: %w", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		return fmt.Errorf("copy sqlite template: %w", err)
	}
	if err := dst.Sync(); err != nil {
		_ = dst.Close()
		return fmt.Errorf("sync sqlite copy: %w", err)
	}
	if err := dst.Close(); err != nil {
		return fmt.Errorf("close sqlite copy: %w", err)
	}
	return nil
}

func newFullRouterTestConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg := newPracticeFlowTestConfig(t)
	cfg.RateLimit.Global.Enabled = false
	cfg.RateLimit.Login.Enabled = false
	cfg.Score = config.ScoreConfig{
		CacheTTL:        time.Minute,
		LockTimeout:     2 * time.Second,
		MaxRankingLimit: 100,
	}
	cfg.Recommendation = config.RecommendationConfig{
		WeakThreshold: 0.4,
		CacheTTL:      time.Hour,
		DefaultLimit:  6,
		MaxLimit:      20,
	}
	cfg.Report = config.ReportConfig{
		StorageDir:      filepath.Join(t.TempDir(), "reports"),
		DefaultFormat:   model.ReportFormatPDF,
		PersonalTimeout: 10 * time.Second,
		ClassTimeout:    10 * time.Second,
		FileTTL:         24 * time.Hour,
		MaxWorkers:      1,
	}
	cfg.Dashboard = config.DashboardConfig{
		CacheTTL:       time.Minute,
		AlertThreshold: 80,
		RedisKeyPrefix: "test:dashboard",
	}
	cfg.Contest = config.ContestConfig{
		StatusUpdateInterval:  time.Minute,
		StatusUpdateBatchSize: 100,
		BaseScore:             1000,
		MinScore:              100,
		Decay:                 0.9,
		FirstBloodBonus:       0.1,
		AWD: config.ContestAWDConfig{
			SchedulerInterval:  30 * time.Second,
			SchedulerBatchSize: 100,
			RoundInterval:      5 * time.Minute,
			RoundLockTTL:       30 * time.Second,
			PreviousRoundGrace: 15 * time.Second,
			CheckerTimeout:     2 * time.Second,
			CheckerHealthPath:  "/health",
		},
	}
	return cfg
}

func seedFullRouterData(t *testing.T, env *fullRouterTestEnv) {
	t.Helper()

	seedRoles(t, env.db)

	env.admin = createFullRouterUser(t, env.db, "admin_matrix", env.adminPwd, model.RoleAdmin, "")
	env.teacher = createFullRouterUser(t, env.db, "teacher_matrix", env.teacherPwd, model.RoleTeacher, env.className)
	env.student = createFullRouterUser(t, env.db, "student_matrix", env.studentPwd, model.RoleStudent, env.className)
	env.peerStudent = createFullRouterUser(t, env.db, "student_peer", "Password123", model.RoleStudent, env.className)
	env.otherTeacher = createFullRouterUser(t, env.db, "teacher_other", "Password123", model.RoleTeacher, "ClassB")
	env.otherStudent = createFullRouterUser(t, env.db, "student_other", "Password123", model.RoleStudent, "ClassB")

	env.image = createFlowImage(t, env.db)

	salt, err := flagcrypto.GenerateSalt()
	if err != nil {
		t.Fatalf("generate flag salt: %v", err)
	}
	env.challenge = &model.Challenge{
		Title:         "Matrix Web Challenge",
		Description:   "challenge for full router integration tests",
		Category:      model.DimensionWeb,
		Difficulty:    model.ChallengeDifficultyEasy,
		Points:        100,
		ImageID:       env.image.ID,
		Status:        model.ChallengeStatusPublished,
		FlagType:      model.FlagTypeStatic,
		FlagSalt:      salt,
		FlagHash:      flagcrypto.HashStaticFlag("flag{matrix}", salt),
		FlagPrefix:    "flag",
		AttachmentURL: "https://example.com/files/matrix.zip",
		CreatedBy:     &env.teacher.ID,
	}
	if err := env.db.Create(env.challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	hint := &model.ChallengeHint{
		ChallengeID: env.challenge.ID,
		Level:       1,
		Title:       "入口提示",
		Content:     "先查看登录表单。",
	}
	if err := env.db.Create(hint).Error; err != nil {
		t.Fatalf("create hint: %v", err)
	}

	writeup := &model.ChallengeWriteup{
		ChallengeID: env.challenge.ID,
		Title:       "题解",
		Content:     "writeup content",
		Visibility:  model.WriteupVisibilityPublic,
		CreatedBy:   &env.admin.ID,
	}
	if err := env.db.Create(writeup).Error; err != nil {
		t.Fatalf("create writeup: %v", err)
	}

	spec, err := model.EncodeTopologySpec(model.TopologySpec{
		Networks: []model.TopologyNetwork{{Key: model.TopologyDefaultNetworkKey, Name: "default"}},
		Nodes: []model.TopologyNode{{
			Key:         "web",
			Name:        "Web Node",
			ImageID:     env.image.ID,
			ServicePort: 80,
			InjectFlag:  true,
			Tier:        model.TopologyTierPublic,
			NetworkKeys: []string{model.TopologyDefaultNetworkKey},
		}},
	})
	if err != nil {
		t.Fatalf("encode topology: %v", err)
	}

	env.template = &model.EnvironmentTemplate{
		Name:         "Matrix Template",
		Description:  "template for integration tests",
		EntryNodeKey: "web",
		Spec:         spec,
	}
	if err := env.db.Create(env.template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}

	now := time.Now()
	env.contest = &model.Contest{
		Title:       "Matrix Jeopardy",
		Description: "contest",
		Mode:        model.ContestModeJeopardy,
		StartTime:   now.Add(-time.Hour),
		EndTime:     now.Add(time.Hour),
		Status:      model.ContestStatusRunning,
	}
	if err := env.db.Create(env.contest).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}

	env.awdContest = &model.Contest{
		Title:       "Matrix AWD",
		Description: "awd contest",
		Mode:        model.ContestModeAWD,
		StartTime:   now.Add(-time.Hour),
		EndTime:     now.Add(time.Hour),
		Status:      model.ContestStatusRunning,
	}
	if err := env.db.Create(env.awdContest).Error; err != nil {
		t.Fatalf("create awd contest: %v", err)
	}

	contestChallenge := &model.ContestChallenge{
		ContestID:   env.contest.ID,
		ChallengeID: env.challenge.ID,
		Points:      100,
		Order:       1,
		IsVisible:   true,
	}
	if err := env.db.Create(contestChallenge).Error; err != nil {
		t.Fatalf("create contest challenge: %v", err)
	}
	awdContestChallenge := &model.ContestChallenge{
		ContestID:   env.awdContest.ID,
		ChallengeID: env.challenge.ID,
		Points:      100,
		Order:       1,
		IsVisible:   true,
	}
	if err := env.db.Create(awdContestChallenge).Error; err != nil {
		t.Fatalf("create awd contest challenge: %v", err)
	}

	env.registration = &model.ContestRegistration{
		ContestID: env.contest.ID,
		UserID:    env.student.ID,
		Status:    model.ContestRegistrationStatusApproved,
	}
	if err := env.db.Create(env.registration).Error; err != nil {
		t.Fatalf("create registration: %v", err)
	}
	awdRegistration := &model.ContestRegistration{
		ContestID: env.awdContest.ID,
		UserID:    env.student.ID,
		Status:    model.ContestRegistrationStatusApproved,
	}
	if err := env.db.Create(awdRegistration).Error; err != nil {
		t.Fatalf("create awd registration: %v", err)
	}

	env.announcement = &model.ContestAnnouncement{
		ContestID: env.contest.ID,
		Title:     "公告",
		Content:   "contest starts",
		CreatedBy: &env.admin.ID,
	}
	if err := env.db.Create(env.announcement).Error; err != nil {
		t.Fatalf("create announcement: %v", err)
	}

	env.team = &model.Team{
		ContestID:  env.contest.ID,
		Name:       "Matrix Team",
		CaptainID:  env.student.ID,
		InviteCode: "MATRIX123",
		MaxMembers: 4,
	}
	if err := env.db.Create(env.team).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := env.db.Create(&model.TeamMember{
		ContestID: env.contest.ID,
		TeamID:    env.team.ID,
		UserID:    env.student.ID,
		JoinedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
	env.registration.TeamID = &env.team.ID
	if err := env.db.Save(env.registration).Error; err != nil {
		t.Fatalf("update registration team: %v", err)
	}

	env.awdRound = &model.AWDRound{
		ContestID:    env.awdContest.ID,
		RoundNumber:  1,
		Status:       model.AWDRoundStatusRunning,
		StartedAt:    &now,
		AttackScore:  50,
		DefenseScore: 50,
	}
	if err := env.db.Create(env.awdRound).Error; err != nil {
		t.Fatalf("create awd round: %v", err)
	}
	if err := env.db.Create(&model.AWDTeamService{
		RoundID:       env.awdRound.ID,
		TeamID:        env.team.ID,
		ChallengeID:   env.challenge.ID,
		ServiceStatus: model.AWDServiceStatusUp,
		CheckResult:   `{"status":"ok"}`,
	}).Error; err != nil {
		t.Fatalf("create awd team service: %v", err)
	}
	if err := env.db.Create(&model.AWDAttackLog{
		RoundID:        env.awdRound.ID,
		AttackerTeamID: env.team.ID,
		VictimTeamID:   env.team.ID,
		ChallengeID:    env.challenge.ID,
		AttackType:     model.AWDAttackTypeFlagCapture,
		Source:         model.AWDAttackSourceManual,
		IsSuccess:      false,
	}).Error; err != nil {
		t.Fatalf("create awd attack log: %v", err)
	}

	runtimeDetails, err := model.EncodeInstanceRuntimeDetails(model.InstanceRuntimeDetails{
		Containers: []model.InstanceRuntimeContainer{{
			NodeKey:      "web",
			ContainerID:  "ctf-instance",
			ServicePort:  80,
			IsEntryPoint: true,
			HostPort:     30001,
		}},
	})
	if err != nil {
		t.Fatalf("encode runtime details: %v", err)
	}
	env.instance = &model.Instance{
		UserID:         env.student.ID,
		ChallengeID:    env.challenge.ID,
		ContainerID:    "ctf-instance",
		NetworkID:      "ctf-network",
		RuntimeDetails: runtimeDetails,
		Status:         model.InstanceStatusRunning,
		AccessURL:      "http://127.0.0.1:30001",
		Nonce:          "matrix-nonce",
		ExpiresAt:      now.Add(2 * time.Hour),
		MaxExtends:     2,
	}
	if err := env.db.Create(env.instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}

	if err := env.db.Create(&model.Submission{
		UserID:      env.student.ID,
		ChallengeID: env.challenge.ID,
		IsCorrect:   true,
		SubmittedAt: now.Add(-10 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("create submission: %v", err)
	}
	if err := env.db.Create(&model.UserScore{
		UserID:     env.student.ID,
		TotalScore: 100,
	}).Error; err != nil {
		t.Fatalf("create user score: %v", err)
	}
	if err := env.db.Create(&model.SkillProfile{
		UserID:    env.student.ID,
		Dimension: model.DimensionWeb,
		Score:     0.3,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create skill profile: %v", err)
	}

	env.notification = &model.Notification{
		UserID:    env.student.ID,
		Type:      model.NotificationTypeSystem,
		Title:     "通知",
		Content:   "hello",
		IsRead:    false,
		CreatedAt: now,
	}
	if err := env.db.Create(env.notification).Error; err != nil {
		t.Fatalf("create notification: %v", err)
	}

	if err := os.MkdirAll(env.reportDir, 0o755); err != nil {
		t.Fatalf("mkdir report dir: %v", err)
	}
	reportPath := filepath.Join(env.reportDir, "personal-report.pdf")
	if err := os.WriteFile(reportPath, []byte("matrix report"), 0o644); err != nil {
		t.Fatalf("write report file: %v", err)
	}
	expiresAt := now.Add(24 * time.Hour)
	completedAt := now
	env.report = &model.Report{
		Type:        model.ReportTypePersonal,
		Format:      model.ReportFormatPDF,
		UserID:      &env.student.ID,
		Status:      model.ReportStatusReady,
		FilePath:    reportPath,
		ExpiresAt:   &expiresAt,
		CompletedAt: &completedAt,
	}
	if err := env.db.Create(env.report).Error; err != nil {
		t.Fatalf("create report: %v", err)
	}
}

func seedRoles(t *testing.T, db *gorm.DB) {
	t.Helper()

	roles := []*model.Role{
		{Code: model.RoleAdmin, Name: "管理员"},
		{Code: model.RoleTeacher, Name: "教师"},
		{Code: model.RoleStudent, Name: "学生"},
	}
	for _, role := range roles {
		if err := db.Create(role).Error; err != nil {
			t.Fatalf("create role %s: %v", role.Code, err)
		}
	}
}

func createFullRouterUser(t *testing.T, db *gorm.DB, username, password, role, className string) *model.User {
	t.Helper()

	user := &model.User{
		Username:  username,
		Email:     fmt.Sprintf("%s@example.com", username),
		Role:      role,
		Status:    model.UserStatusActive,
		ClassName: className,
		Name:      username,
	}
	setTestPassword(t, user, password)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}
	return user
}

func performFullRouterRequest(
	t *testing.T,
	router http.Handler,
	method string,
	target string,
	payload any,
	headers map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, target, &body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}
