package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/config"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	authruntime "ctf-platform/internal/module/auth/runtime"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestentity "ctf-platform/internal/module/contest/entity"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	opsentity "ctf-platform/internal/module/ops/entity"
	practiceentity "ctf-platform/internal/module/practice/entity"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	"ctf-platform/internal/platform/randomstring"
	flagcrypto "ctf-platform/internal/shared/flagcrypto"
	"ctf-platform/internal/shared/taxonomy"
)

type fullRouterTestEnv struct {
	router *gin.Engine
	db     *gorm.DB
	cache  *redislib.Client

	admin        *identitycontracts.User
	teacher      *identitycontracts.User
	student      *identitycontracts.User
	peerStudent  *identitycontracts.User
	otherTeacher *identitycontracts.User
	otherStudent *identitycontracts.User
	studentPwd   string
	teacherPwd   string
	adminPwd     string
	className    string
	reportDir    string
	image        *appImageRow
	challenge    *appChallengeRow
	template     *challengeentity.EnvironmentTemplate
	contest      *contestcontracts.Contest
	awdContest   *contestcontracts.Contest
	registration *contestcontracts.ContestRegistration
	announcement *contestentity.ContestAnnouncement
	team         *contestcontracts.Team
	awdRound     *contestcontracts.AWDRound
	instance     *instancecontracts.Instance
	notification *opsentity.Notification
	report       *assessmententity.Report
}

func TestRouterBuildUsesCompositionModules(t *testing.T) {
	cfg, db, cache := newAppTestDependencies(t)

	var calls []string

	originalBuildContainerRuntimeModule := buildContainerRuntimeModule
	originalBuildInstanceModule := buildInstanceModule
	originalBuildOpsModule := buildOpsModule
	originalBuildIdentityModule := buildIdentityModule
	originalBuildAuthModule := buildAuthModule
	originalBuildChallengeModule := buildChallengeModule
	originalBuildAssessmentModule := buildAssessmentModule
	originalBuildTeachingQueryModule := buildTeachingQueryModule
	originalBuildContestModule := buildContestModule
	originalBuildPracticeModule := buildPracticeModule
	defer func() {
		buildContainerRuntimeModule = originalBuildContainerRuntimeModule
		buildInstanceModule = originalBuildInstanceModule
		buildOpsModule = originalBuildOpsModule
		buildIdentityModule = originalBuildIdentityModule
		buildAuthModule = originalBuildAuthModule
		buildChallengeModule = originalBuildChallengeModule
		buildAssessmentModule = originalBuildAssessmentModule
		buildTeachingQueryModule = originalBuildTeachingQueryModule
		buildContestModule = originalBuildContestModule
		buildPracticeModule = originalBuildPracticeModule
	}()

	buildContainerRuntimeModule = func(root *composition.Root) (*composition.ContainerRuntimeModule, error) {
		if root == nil {
			t.Fatal("expected root for container runtime module builder")
		}
		calls = append(calls, "container_runtime")
		return originalBuildContainerRuntimeModule(root)
	}
	buildInstanceModule = func(root *composition.Root, runtime *composition.ContainerRuntimeModule) *composition.InstanceModule {
		if root == nil || runtime == nil {
			t.Fatal("expected root and container runtime for instance module builder")
		}
		calls = append(calls, "instance")
		return originalBuildInstanceModule(root, runtime)
	}
	buildOpsModule = func(root *composition.Root, runtime *composition.ContainerRuntimeModule) *composition.OpsModule {
		if root == nil || runtime == nil {
			t.Fatal("expected root and container runtime for ops module builder")
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
	buildAuthModule = func(root *composition.Root, ops *composition.OpsModule, identity *composition.IdentityModule, tokenService authcontracts.TokenService) (*authruntime.Module, error) {
		if root == nil || ops == nil || identity == nil || tokenService == nil {
			t.Fatal("expected root, ops, identity and token service for auth module builder")
		}
		calls = append(calls, "auth")
		return originalBuildAuthModule(root, ops, identity, tokenService)
	}
	buildChallengeModule = func(root *composition.Root, runtime *composition.ContainerRuntimeModule) (*composition.ChallengeModule, error) {
		if root == nil || runtime == nil {
			t.Fatal("expected root and container runtime for challenge module builder")
		}
		calls = append(calls, "challenge")
		return originalBuildChallengeModule(root, runtime)
	}
	buildAssessmentModule = func(root *composition.Root, challenge *composition.ChallengeModule) *composition.AssessmentModule {
		if root == nil || challenge == nil {
			t.Fatal("expected root and challenge for assessment module builder")
		}
		calls = append(calls, "assessment")
		return originalBuildAssessmentModule(root, challenge)
	}
	buildTeachingQueryModule = func(root *composition.Root, assessment *composition.AssessmentModule, identity *composition.IdentityModule) *composition.TeachingQueryModule {
		if root == nil || assessment == nil || identity == nil {
			t.Fatal("expected root, assessment and identity for teaching query module builder")
		}
		calls = append(calls, "teaching_query")
		return originalBuildTeachingQueryModule(root, assessment, identity)
	}
	buildContestModule = func(root *composition.Root, challenge *composition.ChallengeModule, runtime *composition.ContainerRuntimeModule) *composition.ContestModule {
		if root == nil || challenge == nil || runtime == nil {
			t.Fatal("expected root, challenge and container runtime for contest module builder")
		}
		calls = append(calls, "contest")
		return originalBuildContestModule(root, challenge, runtime)
	}
	buildPracticeModule = func(root *composition.Root, challenge *composition.ChallengeModule, instance *composition.InstanceModule) *composition.PracticeModule {
		if root == nil || challenge == nil || instance == nil {
			t.Fatal("expected root, challenge and instance for practice module builder")
		}
		calls = append(calls, "practice")
		return originalBuildPracticeModule(root, challenge, instance)
	}

	router, err := NewRouter(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
	if router == nil {
		t.Fatal("expected router")
	}

	expectedCalls := []string{"container_runtime", "ops", "instance", "identity", "auth", "challenge", "assessment", "teaching_query", "contest", "practice"}
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
		"/api/v1/instances/:id/proxy", "/api/v1/instances/:id/proxy/*proxyPath",
		"/api/v1/contests/:id/awd/services/:sid/targets/:team_id/proxy",
		"/api/v1/contests/:id/awd/services/:sid/targets/:team_id/proxy/*proxyPath":
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
	return openInternalAppTestSQLite(t, "full-router.sqlite")
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
		DefaultFormat:   assessmententity.ReportFormatPDF,
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

	env.admin = createFullRouterUser(t, env.db, "admin_matrix", env.adminPwd, identitycontracts.RoleAdmin, "")
	env.teacher = createFullRouterUser(t, env.db, "teacher_matrix", env.teacherPwd, identitycontracts.RoleTeacher, env.className)
	env.student = createFullRouterUser(t, env.db, "student_matrix", env.studentPwd, identitycontracts.RoleStudent, env.className)
	env.peerStudent = createFullRouterUser(t, env.db, "student_peer", "Password123", identitycontracts.RoleStudent, env.className)
	env.otherTeacher = createFullRouterUser(t, env.db, "teacher_other", "Password123", identitycontracts.RoleTeacher, "ClassB")
	env.otherStudent = createFullRouterUser(t, env.db, "student_other", "Password123", identitycontracts.RoleStudent, "ClassB")

	env.image = createFlowImage(t, env.db)

	salt, err := randomstring.Generate()
	if err != nil {
		t.Fatalf("generate flag salt: %v", err)
	}
	env.challenge = &appChallengeRow{
		Title:         "Matrix Web Challenge",
		Description:   "challenge for full router integration tests",
		Category:      taxonomy.DimensionWeb,
		Difficulty:    taxonomy.DifficultyEasy,
		Points:        100,
		ImageID:       env.image.ID,
		Status:        challengecontracts.ChallengeStatusPublished,
		FlagType:      challengecontracts.FlagTypeStatic,
		FlagSalt:      salt,
		FlagHash:      flagcrypto.HashStaticFlag("flag{matrix}", salt),
		FlagPrefix:    "flag",
		AttachmentURL: "https://example.com/files/matrix.zip",
		CreatedBy:     &env.teacher.ID,
	}
	if err := env.db.Create(env.challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	hint := &challengeentity.ChallengeHint{
		ChallengeID: env.challenge.ID,
		Level:       1,
		Title:       "入口提示",
		Content:     "先查看登录表单。",
	}
	if err := env.db.Create(hint).Error; err != nil {
		t.Fatalf("create hint: %v", err)
	}

	writeup := &challengeentity.ChallengeWriteup{
		ChallengeID: env.challenge.ID,
		Title:       "题解",
		Content:     "writeup content",
		Visibility:  challengeentity.WriteupVisibilityPublic,
		CreatedBy:   &env.admin.ID,
	}
	if err := env.db.Create(writeup).Error; err != nil {
		t.Fatalf("create writeup: %v", err)
	}

	spec, err := challengecontracts.EncodeTopologySpec(challengecontracts.TopologySpec{
		Networks: []challengecontracts.TopologyNetwork{{Key: challengecontracts.TopologyDefaultNetworkKey, Name: "default"}},
		Nodes: []challengecontracts.TopologyNode{{
			Key:         "web",
			Name:        "Web Node",
			ImageID:     env.image.ID,
			ServicePort: 80,
			InjectFlag:  true,
			Tier:        challengecontracts.TopologyTierPublic,
			NetworkKeys: []string{challengecontracts.TopologyDefaultNetworkKey},
		}},
	})
	if err != nil {
		t.Fatalf("encode topology: %v", err)
	}

	env.template = &challengeentity.EnvironmentTemplate{
		Name:         "Matrix Template",
		Description:  "template for integration tests",
		EntryNodeKey: "web",
		Spec:         spec,
	}
	if err := env.db.Create(env.template).Error; err != nil {
		t.Fatalf("create template: %v", err)
	}

	now := time.Now()
	env.contest = &contestcontracts.Contest{
		Title:       "Matrix Jeopardy",
		Description: "contest",
		Mode:        contestcontracts.ContestModeJeopardy,
		StartTime:   now.Add(-time.Hour),
		EndTime:     now.Add(time.Hour),
		Status:      contestcontracts.ContestStatusRunning,
	}
	if err := env.db.Create(env.contest).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}

	env.awdContest = &contestcontracts.Contest{
		Title:       "Matrix AWD",
		Description: "awd contest",
		Mode:        contestcontracts.ContestModeAWD,
		StartTime:   now.Add(-time.Hour),
		EndTime:     now.Add(time.Hour),
		Status:      contestcontracts.ContestStatusRunning,
	}
	if err := env.db.Create(env.awdContest).Error; err != nil {
		t.Fatalf("create awd contest: %v", err)
	}

	contestChallenge := &contestcontracts.ContestChallenge{
		ContestID:   env.contest.ID,
		ChallengeID: env.challenge.ID,
		Points:      100,
		Order:       1,
		IsVisible:   true,
	}
	if err := env.db.Create(contestChallenge).Error; err != nil {
		t.Fatalf("create contest challenge: %v", err)
	}
	awdContestChallenge := &contestcontracts.ContestChallenge{
		ContestID:   env.awdContest.ID,
		ChallengeID: env.challenge.ID,
		Points:      100,
		Order:       1,
		IsVisible:   true,
	}
	if err := env.db.Create(awdContestChallenge).Error; err != nil {
		t.Fatalf("create awd contest challenge: %v", err)
	}

	env.registration = &contestcontracts.ContestRegistration{
		ContestID: env.contest.ID,
		UserID:    env.student.ID,
		Status:    contestcontracts.ContestRegistrationStatusApproved,
	}
	if err := env.db.Create(env.registration).Error; err != nil {
		t.Fatalf("create registration: %v", err)
	}
	awdRegistration := &contestcontracts.ContestRegistration{
		ContestID: env.awdContest.ID,
		UserID:    env.student.ID,
		Status:    contestcontracts.ContestRegistrationStatusApproved,
	}
	if err := env.db.Create(awdRegistration).Error; err != nil {
		t.Fatalf("create awd registration: %v", err)
	}

	env.announcement = &contestentity.ContestAnnouncement{
		ContestID: env.contest.ID,
		Title:     "公告",
		Content:   "contest starts",
		CreatedBy: &env.admin.ID,
	}
	if err := env.db.Create(env.announcement).Error; err != nil {
		t.Fatalf("create announcement: %v", err)
	}

	env.team = &contestcontracts.Team{
		ContestID:  env.contest.ID,
		Name:       "Matrix Team",
		CaptainID:  env.student.ID,
		InviteCode: "MATRIX123",
		MaxMembers: 4,
	}
	if err := env.db.Create(env.team).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := env.db.Create(&contestcontracts.TeamMember{
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

	env.awdRound = &contestcontracts.AWDRound{
		ContestID:    env.awdContest.ID,
		RoundNumber:  1,
		Status:       contestcontracts.AWDRoundStatusRunning,
		StartedAt:    &now,
		AttackScore:  50,
		DefenseScore: 50,
	}
	if err := env.db.Create(env.awdRound).Error; err != nil {
		t.Fatalf("create awd round: %v", err)
	}
	if err := env.db.Create(&contestcontracts.AWDTeamService{
		RoundID:        env.awdRound.ID,
		TeamID:         env.team.ID,
		AWDChallengeID: env.challenge.ID,
		ServiceStatus:  contestcontracts.AWDServiceStatusUp,
		CheckResult:    `{"status":"ok"}`,
	}).Error; err != nil {
		t.Fatalf("create awd team service: %v", err)
	}
	if err := env.db.Create(&contestcontracts.AWDAttackLog{
		RoundID:        env.awdRound.ID,
		AttackerTeamID: env.team.ID,
		VictimTeamID:   env.team.ID,
		AWDChallengeID: env.challenge.ID,
		AttackType:     contestcontracts.AWDAttackTypeFlagCapture,
		Source:         contestcontracts.AWDAttackSourceManual,
		IsSuccess:      false,
	}).Error; err != nil {
		t.Fatalf("create awd attack log: %v", err)
	}

	runtimeDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(runtimecontracts.InstanceRuntimeDetails{
		Containers: []runtimecontracts.InstanceRuntimeContainer{{
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
	env.instance = &instancecontracts.Instance{
		UserID:         env.student.ID,
		ChallengeID:    env.challenge.ID,
		ContainerID:    "ctf-instance",
		NetworkID:      "ctf-network",
		RuntimeDetails: runtimeDetails,
		Status:         instancecontracts.InstanceStatusRunning,
		AccessURL:      "http://127.0.0.1:30001",
		Nonce:          "matrix-nonce",
		ExpiresAt:      now.Add(2 * time.Hour),
		MaxExtends:     2,
	}
	if err := env.db.Create(env.instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}

	if err := env.db.Create(&contestcontracts.Submission{
		UserID:      env.student.ID,
		ChallengeID: env.challenge.ID,
		IsCorrect:   true,
		SubmittedAt: now.Add(-10 * time.Minute),
	}).Error; err != nil {
		t.Fatalf("create submission: %v", err)
	}
	if err := env.db.Create(&practiceentity.UserScore{
		UserID:     env.student.ID,
		TotalScore: 100,
	}).Error; err != nil {
		t.Fatalf("create user score: %v", err)
	}
	if err := env.db.Create(&assessmententity.SkillProfile{
		UserID:    env.student.ID,
		Dimension: taxonomy.DimensionWeb,
		Score:     0.3,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create skill profile: %v", err)
	}

	env.notification = &opsentity.Notification{
		UserID:    env.student.ID,
		Type:      opsentity.NotificationTypeSystem,
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
	env.report = &assessmententity.Report{
		Type:        assessmententity.ReportTypePersonal,
		Format:      assessmententity.ReportFormatPDF,
		UserID:      &env.student.ID,
		Status:      assessmententity.ReportStatusReady,
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

	roles := []*identitycontracts.Role{
		{Code: identitycontracts.RoleAdmin, Name: "管理员"},
		{Code: identitycontracts.RoleTeacher, Name: "教师"},
		{Code: identitycontracts.RoleStudent, Name: "学生"},
	}
	for _, role := range roles {
		if err := db.Create(role).Error; err != nil {
			t.Fatalf("create role %s: %v", role.Code, err)
		}
	}
}

func createFullRouterUser(t *testing.T, db *gorm.DB, username, password, role, className string) *identitycontracts.User {
	t.Helper()

	user := &identitycontracts.User{
		Username:  username,
		Email:     fmt.Sprintf("%s@example.com", username),
		Role:      role,
		Status:    identitycontracts.UserStatusActive,
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
