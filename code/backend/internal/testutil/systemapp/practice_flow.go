package systemapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/config"
	"ctf-platform/internal/middleware"
	authhttp "ctf-platform/internal/module/auth/api/http"
	authcmd "ctf-platform/internal/module/auth/application/commands"
	authqry "ctf-platform/internal/module/auth/application/queries"
	authinfra "ctf-platform/internal/module/auth/infrastructure"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecmd "ctf-platform/internal/module/challenge/application/commands"
	challengeqry "ctf-platform/internal/module/challenge/application/queries"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	challengeruntime "ctf-platform/internal/module/challenge/runtime"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycmd "ctf-platform/internal/module/identity/application/commands"
	identityqry "ctf-platform/internal/module/identity/application/queries"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
	instancehttp "ctf-platform/internal/module/instance/api/http"
	instancecmd "ctf-platform/internal/module/instance/application/commands"
	instanceqry "ctf-platform/internal/module/instance/application/queries"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceinfra "ctf-platform/internal/module/instance/infrastructure"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsqry "ctf-platform/internal/module/ops/application/queries"
	opsinfra "ctf-platform/internal/module/ops/infrastructure"
	practicehttp "ctf-platform/internal/module/practice/api/http"
	practicecmd "ctf-platform/internal/module/practice/application/commands"
	practiceqry "ctf-platform/internal/module/practice/application/queries"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	teachingqueryhttp "ctf-platform/internal/module/teaching_query/api/http"
	teachingqueryqueries "ctf-platform/internal/module/teaching_query/application/queries"
	teachingqueryinfra "ctf-platform/internal/module/teaching_query/infrastructure"
	"ctf-platform/internal/shared/taxonomy"
	runtimeadapters "ctf-platform/internal/testutil/runtimeadapters"
	"ctf-platform/internal/validation"
)

type PracticeFlowEnv struct {
	Router  *gin.Engine
	DB      *gorm.DB
	Cache   *redislib.Client
	ImageID int64
}

type FlowEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type FlowLoginResponse struct {
	User struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"user"`
}

type FlowChallengeResponse struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Difficulty  string `json:"difficulty"`
	Points      int    `json:"points"`
	ImageID     int64  `json:"image_id"`
	Status      string `json:"status"`
}

type FlowChallengeListItem struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Category      string `json:"category"`
	Difficulty    string `json:"difficulty"`
	Points        int    `json:"points"`
	SolvedCount   int64  `json:"solved_count"`
	TotalAttempts int64  `json:"total_attempts"`
	IsSolved      bool   `json:"is_solved"`
}

type FlowChallengeHint struct {
	Level   int    `json:"level"`
	Content string `json:"content"`
}

type FlowChallengeDetail struct {
	ID            int64               `json:"id"`
	Title         string              `json:"title"`
	Category      string              `json:"category"`
	Difficulty    string              `json:"difficulty"`
	Points        int                 `json:"points"`
	NeedTarget    bool                `json:"need_target"`
	AttachmentURL string              `json:"attachment_url"`
	Hints         []FlowChallengeHint `json:"hints"`
	SolvedCount   int64               `json:"solved_count"`
	TotalAttempts int64               `json:"total_attempts"`
	IsSolved      bool                `json:"is_solved"`
}

type FlowSubmissionResponse struct {
	IsCorrect bool   `json:"is_correct"`
	Message   string `json:"message"`
	Points    int    `json:"points"`
}

type FlowSubmissionRecord struct {
	ID          int64  `json:"id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	Answer      string `json:"answer"`
	SubmittedAt string `json:"submitted_at"`
}

type FlowInstanceResponse struct {
	ID        int64  `json:"id"`
	AccessURL string `json:"access_url"`
	Status    string `json:"status"`
}

type FlowPageResponse[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type FlowProgressResponse struct {
	TotalScore  int `json:"total_score"`
	TotalSolved int `json:"total_solved"`
	Rank        int `json:"rank"`
}

type FlowTimelineEvent struct {
	Type        string `json:"type"`
	ChallengeID int64  `json:"challenge_id"`
	Title       string `json:"title"`
	IsCorrect   *bool  `json:"is_correct"`
	Points      *int   `json:"points"`
	Detail      string `json:"detail"`
}

type FlowTimelineResponse struct {
	Events []FlowTimelineEvent `json:"events"`
}

type FlowTeacherEvidenceReviewSummary struct {
	TotalEvents       int   `json:"total_events"`
	ProxyRequestCount int   `json:"proxy_request_count"`
	SubmitCount       int   `json:"submit_count"`
	SuccessCount      int   `json:"success_count"`
	ChallengeID       int64 `json:"challenge_id"`
}

type FlowTeacherEvidenceReviewEvent struct {
	Type        string                 `json:"type"`
	ChallengeID int64                  `json:"challenge_id"`
	Title       string                 `json:"title"`
	Detail      string                 `json:"detail"`
	Meta        map[string]interface{} `json:"meta"`
}

type FlowTeacherEvidenceReviewResponse struct {
	Summary FlowTeacherEvidenceReviewSummary `json:"summary"`
	Events  []FlowTeacherEvidenceReviewEvent `json:"events"`
}

type FlowTeacherAttackSessionEvent struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type FlowTeacherAttackSession struct {
	ID          string                          `json:"id"`
	Mode        string                          `json:"mode"`
	ChallengeID *int64                          `json:"challenge_id"`
	Result      string                          `json:"result"`
	EventCount  int                             `json:"event_count"`
	Events      []FlowTeacherAttackSessionEvent `json:"events"`
}

type FlowTeacherAttackSessionSummary struct {
	TotalSessions   int `json:"total_sessions"`
	SuccessCount    int `json:"success_count"`
	FailedCount     int `json:"failed_count"`
	InProgressCount int `json:"in_progress_count"`
	UnknownCount    int `json:"unknown_count"`
	EventCount      int `json:"event_count"`
}

type FlowTeacherAttackSessionResponse struct {
	Summary  FlowTeacherAttackSessionSummary `json:"summary"`
	Sessions []FlowTeacherAttackSession      `json:"sessions"`
}

type FlowAuditItem struct {
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *int64                 `json:"resource_id"`
	ActorUserID  *int64                 `json:"actor_user_id"`
	Detail       map[string]interface{} `json:"detail"`
}

type PracticeFlowScenarioResult struct {
	AdminSession      *http.Cookie
	StudentSession    *http.Cookie
	Challenge         FlowChallengeResponse
	ListBeforeItems   []FlowChallengeListItem
	DetailBody        []byte
	Detail            FlowChallengeDetail
	Instance          FlowInstanceResponse
	ProxyAccess       FlowInstanceResponse
	ProxyLocation     string
	ProxyCookies      []*http.Cookie
	WrongSubmission   FlowSubmissionResponse
	CorrectSubmission FlowSubmissionResponse
	RepeatSubmission  FlowSubmissionResponse
	SubmissionHistory []FlowSubmissionRecord
	ListAfterItems    []FlowChallengeListItem
	Progress          FlowProgressResponse
	Timeline          FlowTimelineResponse
	Evidence          FlowTeacherEvidenceReviewResponse
	AttackSessions    FlowTeacherAttackSessionResponse
	AuditPage         FlowPageResponse[FlowAuditItem]
	Submissions       []contestcontracts.Submission
}

type teachingQueryIdentityLookupAdapter struct {
	users identitycontracts.UserLookupRepository
}

func (a teachingQueryIdentityLookupAdapter) FindUserByID(ctx context.Context, userID int64) (*identitycontracts.User, error) {
	user, err := a.users.FindByID(ctx, userID)
	if errors.Is(err, identitycontracts.ErrUserNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

var testPasswordHashCache sync.Map

func setTestPassword(t *testing.T, user *identitycontracts.User, password string) {
	t.Helper()

	if user == nil {
		t.Fatal("setTestPassword() got nil user")
	}

	hash, err := testPasswordHash(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user.PasswordHash = hash
}

func testPasswordHash(password string) (string, error) {
	if cached, ok := testPasswordHashCache.Load(password); ok {
		return cached.(string), nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	hashValue := string(hash)
	testPasswordHashCache.Store(password, hashValue)
	return hashValue, nil
}

func NewPracticeFlowTestEnv(t *testing.T) *PracticeFlowEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)
	if err := validation.Register(); err != nil {
		t.Fatalf("register validator: %v", err)
	}

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	cache := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	pingCtx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if err := cache.Ping(pingCtx).Err(); err != nil {
		t.Fatalf("ping test redis: %v", err)
	}

	db := OpenInternalAppTestSQLite(t, "practice-flow.sqlite")

	cfg := NewPracticeFlowTestConfig(t)
	logger := zap.NewNop()

	tokenService := authinfra.NewTokenService(cfg.Auth, cfg.WebSocket, cache)
	authRepo := identityinfra.NewRepository(db)
	authService := authcmd.NewService(authRepo, tokenService, cfg.RateLimit.Login, logger)
	casCommandService := authcmd.NewCASService(cfg.Auth.CAS, authRepo, tokenService, logger.Named("cas_command_service"), nil)
	casQueryService := authqry.NewCASService(cfg.Auth.CAS)
	profileCommandService := identitycmd.NewProfileService(authRepo, logger.Named("identity_profile_command_service"))
	profileQueryService := identityqry.NewProfileService(authRepo)
	auditRepo := opsinfra.NewAuditRepository(db)
	auditCommandService := opscmd.NewAuditService(auditRepo, logger)
	auditQueryService := opsqry.NewAuditService(auditRepo, cfg.Pagination, logger)
	authHandler := authhttp.NewHandler(authService, profileCommandService, profileQueryService, tokenService, casCommandService, casQueryService, authhttp.CookieConfig{
		Name:     cfg.Auth.SessionCookieName,
		Path:     cfg.Auth.SessionCookiePath,
		Secure:   cfg.Auth.SessionCookieSecure,
		HTTPOnly: cfg.Auth.SessionCookieHTTPOnly,
		SameSite: cfg.Auth.CookieSameSite(),
		MaxAge:   cfg.Auth.SessionTTL,
	}, logger, auditCommandService)
	auditHandler := opshttp.NewAuditHandler(auditQueryService)

	challengeRepo := challengeinfra.NewRepository(db)
	flagRepo := challengeinfra.NewFlagRepository(challengeRepo)
	imageRepo := challengeinfra.NewImageRepository(db)
	challengeCommandService := challengecmd.NewChallengeService(
		challengeinfra.NewChallengeCommandRepository(challengeRepo),
		challengeinfra.NewImageQueryRepository(imageRepo),
		challengeinfra.NewTopologyServiceRepository(challengeRepo),
		challengeinfra.NewTopologyPackageRevisionRepository(challengeRepo),
		nil,
		challengecmd.SelfCheckConfig{
			RuntimeCreateTimeout: cfg.Container.CreateTimeout,
			FlagGlobalSecret:     cfg.Container.FlagGlobalSecret,
		},
		logger,
	)
	challengeCommandService.SetChallengeImportTxRunner(challengeruntime.NewChallengeImportTxRunner(challengeRepo, nil))
	challengeCommandService.SetChallengePackageExportTxRunner(challengeruntime.NewChallengePackageExportTxRunner(challengeRepo))
	challengeQueryService := challengeqry.NewChallengeService(challengeinfra.NewChallengeQueryRepository(challengeRepo), challengeinfra.NewSolvedCountCache(cache), &challengeqry.Config{
		SolvedCountCacheTTL: cfg.Challenge.SolvedCountCacheTTL,
	}, logger)
	challengeHandler := challengehttp.NewHandler(challengeCommandService, challengeQueryService)

	flagQueryService, err := challengeqry.NewFlagService(flagRepo, cfg.Container.FlagGlobalSecret)
	if err != nil {
		t.Fatalf("create flag query service: %v", err)
	}
	flagCommandService, err := challengecmd.NewFlagService(flagRepo, cfg.Container.FlagGlobalSecret)
	if err != nil {
		t.Fatalf("create flag command service: %v", err)
	}
	flagHandler := challengehttp.NewFlagHandler(flagCommandService, flagQueryService)

	practiceRepo := practiceinfra.NewRepositoryWithRuntimePortOwner(db, func(db *gorm.DB) runtimeports.PortReservationOwner {
		return runtimeinfrarepo.NewAllocationRepository(db)
	})
	instanceRepo := instanceinfra.NewRepository(db)
	proxyTicketInstanceRepo := instanceinfra.NewRepository(db)
	root, err := composition.BuildRoot(cfg, logger, db, cache)
	if err != nil {
		t.Fatalf("build composition root: %v", err)
	}
	runtimeModule, err := composition.BuildRuntimeModule(root)
	if err != nil {
		t.Fatalf("build runtime module: %v", err)
	}
	instanceModule := composition.BuildInstanceModule(root, runtimeModule)
	runtimeCleanupService := runtimecmd.NewRuntimeCleanupService(nil, nil, logger)
	runtimeInstanceCommands := instancecmd.NewInstanceService(instanceRepo, systemRuntimeCleanerAdapter{cleaner: runtimeCleanupService}, &cfg.Container, logger)
	runtimeInstanceQueries := instanceqry.NewInstanceService(instanceRepo, &cfg.Container)
	runtimeProxyTicketService := instanceqry.NewProxyTicketService(instanceinfra.NewProxyTicketStore(cache), proxyTicketInstanceRepo, cfg.Container.ProxyTicketTTL)
	runtimeService := runtimeadapters.NewHTTPService(
		runtimeInstanceCommands,
		runtimeInstanceQueries,
		runtimeProxyTicketService,
		cfg.Container.ProxyBodyPreviewSize,
	)
	scoreStateStore := practiceinfra.NewScoreStateStore(cache)
	flagSubmitRateLimitStore := practiceinfra.NewFlagSubmitRateLimitStore(cache, cfg.RateLimit.RedisKeyPrefix)
	practiceScoreCommandService := practicecmd.NewScoreService(practiceRepo, scoreStateStore, logger, &cfg.Score)
	practiceService := practicecmd.NewService(
		practiceRepo,
		imageRepo,
		instanceModule.PracticeInstanceRepository,
		instanceModule.PracticeRuntimeService,
		practiceScoreCommandService,
		flagSubmitRateLimitStore,
		cfg,
		logger).
		SetSolvedSubmissionRepository(practiceinfra.NewSolvedSubmissionRepository(practiceRepo)).
		SetManualReviewRepository(practiceinfra.NewManualReviewRepository(practiceRepo)).
		SetContestScopeRepository(practiceinfra.NewContestScopeRepository(practiceRepo)).
		SetRuntimeSubjectRepository(practiceinfra.NewRuntimeSubjectRepository(challengeRepo)).
		SetInstanceReadinessProbe(practiceinfra.NewInstanceReadinessProbe())

	practiceScoreQueryService := practiceqry.NewScoreService(practiceinfra.NewScoreQueryRepository(practiceRepo), scoreStateStore, logger, &cfg.Score)
	practiceProgressTimelineService := practiceqry.NewProgressTimelineService(
		practiceRepo,
		practiceinfra.NewProgressCache(cache),
		cfg.Cache.ProgressTTL,
		logger,
	)
	practiceHandler := practicehttp.NewHandler(practiceService, practiceScoreQueryService, practiceProgressTimelineService)
	teachingQueryRepo := teachingqueryinfra.NewRepository(db)
	teachingQueryUsers := teachingQueryIdentityLookupAdapter{users: authRepo}
	teachingQueryService := teachingqueryqueries.NewQueryService(teachingQueryUsers, teachingQueryRepo, cfg.Pagination)
	teachingQueryOverviewService := teachingqueryqueries.NewOverviewService(teachingQueryUsers, teachingQueryRepo)
	teachingQueryClassInsightService := teachingqueryqueries.NewClassInsightService(teachingQueryUsers, teachingQueryRepo, nil, logger)
	teachingQueryStudentReviewService := teachingqueryqueries.NewStudentReviewService(teachingQueryUsers, teachingQueryRepo, nil)
	teachingQueryHandler := teachingqueryhttp.NewHandler(
		teachingQueryService,
		teachingQueryOverviewService,
		teachingQueryClassInsightService,
		teachingQueryStudentReviewService,
	)
	runtimeHandler := instancehttp.NewHandler(runtimeService, cfg.Container.PublicHost, cfg.Container.AccessHost, auditCommandService, instancehttp.CookieConfig{}, nil)

	createFlowUser(t, db, "admin_user", "Password123", identitycontracts.RoleAdmin)
	createFlowUser(t, db, "student_user", "Password123", identitycontracts.RoleStudent)
	imageID := createFlowImage(t, db)

	router := gin.New()
	router.Use(middleware.RequestID())

	apiV1 := router.Group("/api/v1")
	authGroup := apiV1.Group("/auth")
	authGroup.POST("/login", authHandler.Login)

	protected := apiV1.Group("")
	protected.Use(middleware.Auth(tokenService, cfg.Auth.SessionCookieName))

	authoringOnly := protected.Group("/authoring")
	authoringOnly.Use(middleware.RequireRole(identitycontracts.RoleTeacher))
	authoringOnly.POST("/challenges", challengeHandler.CreateChallenge)
	authoringOnly.PUT("/challenges/:id/flag", flagHandler.ConfigureFlag)

	adminOnly := protected.Group("/admin")
	adminOnly.Use(middleware.RequireRole(identitycontracts.RoleAdmin))
	adminOnly.GET("/audit-logs", auditHandler.ListAuditLogs)

	protected.GET("/challenges", challengeHandler.ListPublishedChallenges)
	protected.GET("/challenges/:id",
		middleware.Audit(auditCommandService, middleware.AuditOptions{
			Action:          auditlog.ActionRead,
			ResourceType:    "challenge_detail",
			ResourceIDParam: "id",
		}, logger),
		challengeHandler.GetPublishedChallenge,
	)
	protected.POST("/challenges/:id/submit",
		middleware.Audit(auditCommandService, middleware.AuditOptions{
			Action:          auditlog.ActionSubmit,
			ResourceType:    "challenge_submission",
			ResourceIDParam: "id",
		}, logger),
		practiceHandler.SubmitFlag,
	)
	protected.GET("/challenges/:id/submissions/mine", practiceHandler.ListMyChallengeSubmissions)
	protected.POST("/challenges/:id/instances", practiceHandler.StartChallenge)
	protected.POST("/instances/:id/access", runtimeHandler.AccessInstance)
	apiV1.GET("/instances/:id/proxy", runtimeHandler.ProxyInstance)
	apiV1.Any("/instances/:id/proxy/*proxyPath", runtimeHandler.ProxyInstance)
	usersGroup := protected.Group("/users")
	usersGroup.GET("/me/progress", practiceHandler.GetProgress)
	usersGroup.GET("/me/timeline", practiceHandler.GetTimeline)
	teacherGroup := protected.Group("/teacher")
	teacherGroup.Use(middleware.RequireRole(identitycontracts.RoleTeacher, identitycontracts.RoleAdmin))
	teacherGroup.GET("/students/:id/evidence", teachingQueryHandler.GetStudentEvidence)
	teacherGroup.GET("/students/:id/attack-sessions", teachingQueryHandler.GetStudentAttackSessions)

	t.Cleanup(func() {
		if sqlDB, sqlErr := db.DB(); sqlErr == nil {
			_ = sqlDB.Close()
		}
		_ = cache.Close()
		mini.Close()
	})

	return &PracticeFlowEnv{
		Router:  router,
		DB:      db,
		Cache:   cache,
		ImageID: imageID,
	}
}

type systemRuntimeCleanerAdapter struct {
	cleaner *runtimecmd.RuntimeCleanupService
}

func (a systemRuntimeCleanerAdapter) CleanupRuntime(ctx context.Context, instance *instancecontracts.Instance) error {
	if a.cleaner == nil || instance == nil {
		return nil
	}
	return a.cleaner.CleanupRuntime(ctx, runtimecontracts.RuntimeCleanupTarget{
		InstanceID:     instance.ID,
		NodeID:         instance.NodeID,
		ContainerID:    instance.ContainerID,
		NetworkID:      instance.NetworkID,
		HostPort:       instance.HostPort,
		RuntimeDetails: instance.RuntimeDetails,
	})
}

func NewPracticeFlowTestConfig(t *testing.T) *config.Config {
	t.Helper()

	portRangeStart, portRangeEnd := ReserveInternalAppTestPortRange(t, 101)

	return &config.Config{
		App: config.AppConfig{
			Name: "ctf-platform-test",
			Env:  "test",
		},
		Auth: config.AuthConfig{
			SessionTTL:            24 * time.Hour,
			SessionCookieName:     "ctf_session",
			SessionCookiePath:     "/",
			SessionCookieHTTPOnly: true,
			SessionCookieSameSite: "lax",
			SessionKeyPrefix:      "test:session",
		},
		RateLimit: config.RateLimitConfig{
			RedisKeyPrefix: "test:rate_limit",
			FlagSubmit: config.RateLimitPolicyConfig{
				Enabled: true,
				Limit:   10,
				Window:  time.Minute,
			},
		},
		Challenge: config.ChallengeConfig{
			SolvedCountCacheTTL: time.Minute,
		},
		Cache: config.CacheConfig{
			ProgressTTL: time.Minute,
		},
		Container: config.ContainerConfig{
			FlagGlobalSecret:     "12345678901234567890123456789012",
			MaxConcurrentPerUser: 3,
			MaxExtends:           2,
			DefaultTTL:           2 * time.Hour,
			ExtendDuration:       30 * time.Minute,
			CreateTimeout:        5 * time.Second,
			PublicHost:           "127.0.0.1",
			DefaultExposedPort:   80,
			PortRangeStart:       portRangeStart,
			PortRangeEnd:         portRangeEnd,
			ProxyTicketTTL:       15 * time.Minute,
			ProxyBodyPreviewSize: 1024,
		},
		Assessment: config.AssessmentConfig{
			IncrementalUpdateDelay:   10 * time.Millisecond,
			IncrementalUpdateTimeout: time.Second,
		},
		Pagination: config.PaginationConfig{
			DefaultPageSize: 20,
			MaxPageSize:     100,
		},
		WebSocket: config.WebSocketConfig{
			TicketTTL:         30 * time.Second,
			TicketKeyPrefix:   "test:ws:ticket",
			HeartbeatInterval: 100 * time.Millisecond,
			ReadTimeout:       time.Second,
			RetryInitialDelay: time.Second,
			RetryMaxDelay:     5 * time.Second,
		},
	}
}

func createFlowUser(t *testing.T, db *gorm.DB, username, password, role string) *identitycontracts.User {
	t.Helper()

	user := &identitycontracts.User{
		Username: username,
		Email:    fmt.Sprintf("%s@example.com", username),
		Role:     role,
		Status:   identitycontracts.UserStatusActive,
	}
	setTestPassword(t, user, password)
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func createFlowImage(t *testing.T, db *gorm.DB) int64 {
	t.Helper()

	image := &appImageRow{
		Name:   "ctf/web-basic",
		Tag:    "v1",
		Status: challengecontracts.ImageStatusAvailable,
	}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	return image.ID
}

func LoginForSession(t *testing.T, router http.Handler, username, password string) *http.Cookie {
	t.Helper()

	resp := PerformFlowJSONRequest(
		t,
		router,
		http.MethodPost,
		"/api/v1/auth/login",
		map[string]any{
			"username": username,
			"password": password,
		},
		nil,
		nil,
	)
	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected login status for %s: %d body=%s", username, resp.Code, resp.Body.String())
	}
	body := DecodeFlowEnvelope(t, resp)
	_ = DecodeFlowJSON[FlowLoginResponse](t, body.Data)
	sessionCookie := cloneCookie(resp.Result().Cookies(), "ctf_session")
	if sessionCookie == nil {
		t.Fatalf("expected session cookie for %s", username)
	}
	return sessionCookie
}

func SessionHeaders(cookie *http.Cookie) map[string]string {
	if cookie == nil {
		return nil
	}
	return map[string]string{
		"Cookie": fmt.Sprintf("%s=%s", cookie.Name, cookie.Value),
	}
}

func LoginForToken(t *testing.T, router http.Handler, username, password string) string {
	t.Helper()
	return LoginForSession(t, router, username, password).Value
}

func BearerHeaders(token string) map[string]string {
	if token == "" {
		return nil
	}
	return map[string]string{
		"Cookie": "ctf_session=" + token,
	}
}

func cloneCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			cloned := *cookie
			return &cloned
		}
	}
	return nil
}

func PerformFlowJSONRequest(
	t *testing.T,
	router http.Handler,
	method string,
	target string,
	payload any,
	headers map[string]string,
	cookies []*http.Cookie,
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
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func DecodeFlowEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) FlowEnvelope {
	t.Helper()

	var envelope FlowEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, recorder.Body.String())
	}
	return envelope
}

func DecodeFlowJSON[T any](t *testing.T, data []byte) T {
	t.Helper()

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode payload: %v payload=%s", err, string(data))
	}
	return value
}

func MustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return data
}

func RunPublishedPracticeFlowScenario(t *testing.T) *PracticeFlowScenarioResult {
	t.Helper()

	env := NewPracticeFlowTestEnv(t)
	result := &PracticeFlowScenarioResult{}

	adminSession := LoginForSession(t, env.Router, "admin_user", "Password123")
	studentSession := LoginForSession(t, env.Router, "student_user", "Password123")
	result.AdminSession = adminSession
	result.StudentSession = studentSession

	createResp := PerformFlowJSONRequest(
		t,
		env.Router,
		http.MethodPost,
		"/api/v1/authoring/challenges",
		map[string]any{
			"title":          "Web SQLi 101",
			"description":    "basic sql injection challenge",
			"category":       taxonomy.DimensionWeb,
			"difficulty":     taxonomy.DifficultyEasy,
			"points":         100,
			"image_id":       env.ImageID,
			"attachment_url": "https://example.com/files/web-sqli-101.zip",
			"hints": []map[string]any{{
				"level":   1,
				"title":   "入口提示",
				"content": "先观察登录表单的参数。",
			}},
		},
		SessionHeaders(adminSession),
		nil,
	)
	if createResp.Code != http.StatusOK {
		t.Fatalf("unexpected create challenge status: %d body=%s", createResp.Code, createResp.Body.String())
	}
	createBody := DecodeFlowEnvelope(t, createResp)
	challenge := DecodeFlowJSON[FlowChallengeResponse](t, createBody.Data)
	result.Challenge = challenge

	configureFlagResp := PerformFlowJSONRequest(
		t,
		env.Router,
		http.MethodPut,
		"/api/v1/authoring/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/flag",
		map[string]any{
			"flag_type": "static",
			"flag":      "flag{sqli_success}",
		},
		SessionHeaders(adminSession),
		nil,
	)
	if configureFlagResp.Code != http.StatusOK {
		t.Fatalf("unexpected configure flag status: %d body=%s", configureFlagResp.Code, configureFlagResp.Body.String())
	}

	if err := env.DB.Model(&appChallengeRow{}).
		Where("id = ?", challenge.ID).
		Update("status", challengecontracts.ChallengeStatusPublished).Error; err != nil {
		t.Fatalf("set challenge published: %v", err)
	}

	listBeforeResp := PerformFlowJSONRequest(t, env.Router, http.MethodGet, "/api/v1/challenges", nil, SessionHeaders(studentSession), nil)
	if listBeforeResp.Code != http.StatusOK {
		t.Fatalf("unexpected list challenges status: %d body=%s", listBeforeResp.Code, listBeforeResp.Body.String())
	}
	listBeforeBody := DecodeFlowEnvelope(t, listBeforeResp)
	listBefore := DecodeFlowJSON[practicecontracts.PageResult[json.RawMessage]](t, listBeforeBody.Data)
	result.ListBeforeItems = DecodeFlowJSON[[]FlowChallengeListItem](t, MustMarshalJSON(t, listBefore.List))

	detailResp := PerformFlowJSONRequest(
		t, env.Router, http.MethodGet, "/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10), nil, SessionHeaders(studentSession), nil,
	)
	if detailResp.Code != http.StatusOK {
		t.Fatalf("unexpected challenge detail status: %d body=%s", detailResp.Code, detailResp.Body.String())
	}
	detailBody := DecodeFlowEnvelope(t, detailResp)
	result.DetailBody = detailBody.Data
	result.Detail = DecodeFlowJSON[FlowChallengeDetail](t, detailBody.Data)

	instanceCreateResp := PerformFlowJSONRequest(
		t, env.Router, http.MethodPost, "/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/instances", nil, SessionHeaders(studentSession), nil,
	)
	if instanceCreateResp.Code != http.StatusOK {
		t.Fatalf("unexpected create instance status: %d body=%s", instanceCreateResp.Code, instanceCreateResp.Body.String())
	}
	instanceCreateBody := DecodeFlowEnvelope(t, instanceCreateResp)
	instance := DecodeFlowJSON[FlowInstanceResponse](t, instanceCreateBody.Data)
	result.Instance = instance

	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/submit" {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("submitted"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("target ok"))
	}))
	t.Cleanup(targetServer.Close)
	if err := env.DB.Model(&instancecontracts.Instance{}).
		Where("id = ?", instance.ID).
		Update("access_url", targetServer.URL).Error; err != nil {
		t.Fatalf("update instance access url: %v", err)
	}

	instanceAccessResp := PerformFlowJSONRequest(
		t, env.Router, http.MethodPost, "/api/v1/instances/"+strconv.FormatInt(instance.ID, 10)+"/access", nil, SessionHeaders(studentSession), nil,
	)
	if instanceAccessResp.Code != http.StatusOK {
		t.Fatalf("unexpected instance access status: %d body=%s", instanceAccessResp.Code, instanceAccessResp.Body.String())
	}
	instanceAccessBody := DecodeFlowEnvelope(t, instanceAccessResp)
	result.ProxyAccess = DecodeFlowJSON[FlowInstanceResponse](t, instanceAccessBody.Data)

	proxyBootstrapResp := PerformFlowJSONRequest(t, env.Router, http.MethodGet, result.ProxyAccess.AccessURL, nil, nil, nil)
	if proxyBootstrapResp.Code != http.StatusFound {
		t.Fatalf("expected proxy bootstrap redirect, got %d body=%s", proxyBootstrapResp.Code, proxyBootstrapResp.Body.String())
	}
	result.ProxyLocation = proxyBootstrapResp.Header().Get("Location")
	result.ProxyCookies = proxyBootstrapResp.Result().Cookies()

	proxyPageResp := PerformFlowJSONRequest(t, env.Router, http.MethodGet, result.ProxyLocation, nil, nil, result.ProxyCookies)
	if proxyPageResp.Code != http.StatusOK || !strings.Contains(proxyPageResp.Body.String(), "target ok") {
		t.Fatalf("expected proxied page response, got %d body=%s", proxyPageResp.Code, proxyPageResp.Body.String())
	}

	proxySubmitResp := PerformFlowJSONRequest(
		t, env.Router, http.MethodPost, "/api/v1/instances/"+strconv.FormatInt(instance.ID, 10)+"/proxy/submit", map[string]any{"payload": "' OR 1=1 --"}, nil, result.ProxyCookies,
	)
	if proxySubmitResp.Code != http.StatusCreated {
		t.Fatalf("expected proxied submit response, got %d body=%s", proxySubmitResp.Code, proxySubmitResp.Body.String())
	}

	wrongSubmitResp := PerformFlowJSONRequest(
		t, env.Router, http.MethodPost, "/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/submit", map[string]any{"flag": "flag{wrong_answer}"}, SessionHeaders(studentSession), nil,
	)
	if wrongSubmitResp.Code != http.StatusOK {
		t.Fatalf("unexpected wrong submit status: %d body=%s", wrongSubmitResp.Code, wrongSubmitResp.Body.String())
	}
	result.WrongSubmission = DecodeFlowJSON[FlowSubmissionResponse](t, DecodeFlowEnvelope(t, wrongSubmitResp).Data)

	correctSubmitResp := PerformFlowJSONRequest(
		t, env.Router, http.MethodPost, "/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/submit", map[string]any{"flag": "flag{sqli_success}"}, SessionHeaders(studentSession), nil,
	)
	if correctSubmitResp.Code != http.StatusOK {
		t.Fatalf("unexpected correct submit status: %d body=%s", correctSubmitResp.Code, correctSubmitResp.Body.String())
	}
	result.CorrectSubmission = DecodeFlowJSON[FlowSubmissionResponse](t, DecodeFlowEnvelope(t, correctSubmitResp).Data)

	submissionHistoryResp := PerformFlowJSONRequest(
		t, env.Router, http.MethodGet, "/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/submissions/mine", nil, SessionHeaders(studentSession), nil,
	)
	if submissionHistoryResp.Code != http.StatusOK {
		t.Fatalf("unexpected submission history status: %d body=%s", submissionHistoryResp.Code, submissionHistoryResp.Body.String())
	}
	result.SubmissionHistory = DecodeFlowJSON[[]FlowSubmissionRecord](t, DecodeFlowEnvelope(t, submissionHistoryResp).Data)

	repeatSubmitResp := PerformFlowJSONRequest(
		t, env.Router, http.MethodPost, "/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/submit", map[string]any{"flag": "flag{sqli_success}"}, SessionHeaders(studentSession), nil,
	)
	if repeatSubmitResp.Code != http.StatusOK {
		t.Fatalf("unexpected repeat submit status: %d body=%s", repeatSubmitResp.Code, repeatSubmitResp.Body.String())
	}
	result.RepeatSubmission = DecodeFlowJSON[FlowSubmissionResponse](t, DecodeFlowEnvelope(t, repeatSubmitResp).Data)

	listAfterResp := PerformFlowJSONRequest(t, env.Router, http.MethodGet, "/api/v1/challenges", nil, SessionHeaders(studentSession), nil)
	if listAfterResp.Code != http.StatusOK {
		t.Fatalf("unexpected list challenges after submit status: %d body=%s", listAfterResp.Code, listAfterResp.Body.String())
	}
	listAfterBody := DecodeFlowEnvelope(t, listAfterResp)
	listAfter := DecodeFlowJSON[practicecontracts.PageResult[json.RawMessage]](t, listAfterBody.Data)
	result.ListAfterItems = DecodeFlowJSON[[]FlowChallengeListItem](t, MustMarshalJSON(t, listAfter.List))

	progressResp := PerformFlowJSONRequest(t, env.Router, http.MethodGet, "/api/v1/users/me/progress", nil, SessionHeaders(studentSession), nil)
	if progressResp.Code != http.StatusOK {
		t.Fatalf("unexpected progress status: %d body=%s", progressResp.Code, progressResp.Body.String())
	}
	result.Progress = DecodeFlowJSON[FlowProgressResponse](t, DecodeFlowEnvelope(t, progressResp).Data)

	timelineResp := PerformFlowJSONRequest(t, env.Router, http.MethodGet, "/api/v1/users/me/timeline", nil, SessionHeaders(studentSession), nil)
	if timelineResp.Code != http.StatusOK {
		t.Fatalf("unexpected timeline status: %d body=%s", timelineResp.Code, timelineResp.Body.String())
	}
	result.Timeline = DecodeFlowJSON[FlowTimelineResponse](t, DecodeFlowEnvelope(t, timelineResp).Data)

	evidenceResp := PerformFlowJSONRequest(
		t, env.Router, http.MethodGet, "/api/v1/teacher/students/2/evidence?challenge_id="+strconv.FormatInt(challenge.ID, 10), nil, SessionHeaders(adminSession), nil,
	)
	if evidenceResp.Code != http.StatusOK {
		t.Fatalf("unexpected evidence status: %d body=%s", evidenceResp.Code, evidenceResp.Body.String())
	}
	result.Evidence = DecodeFlowJSON[FlowTeacherEvidenceReviewResponse](t, DecodeFlowEnvelope(t, evidenceResp).Data)

	attackSessionsResp := PerformFlowJSONRequest(
		t, env.Router, http.MethodGet, "/api/v1/teacher/students/2/attack-sessions?challenge_id="+strconv.FormatInt(challenge.ID, 10)+"&with_events=false", nil, SessionHeaders(adminSession), nil,
	)
	if attackSessionsResp.Code != http.StatusOK {
		t.Fatalf("unexpected attack sessions status: %d body=%s", attackSessionsResp.Code, attackSessionsResp.Body.String())
	}
	result.AttackSessions = DecodeFlowJSON[FlowTeacherAttackSessionResponse](t, DecodeFlowEnvelope(t, attackSessionsResp).Data)

	auditResp := PerformFlowJSONRequest(
		t, env.Router, http.MethodGet, "/api/v1/admin/audit-logs?action=submit&page=1&page_size=10", nil, SessionHeaders(adminSession), nil,
	)
	if auditResp.Code != http.StatusOK {
		t.Fatalf("unexpected audit log status: %d body=%s", auditResp.Code, auditResp.Body.String())
	}
	result.AuditPage = DecodeFlowJSON[FlowPageResponse[FlowAuditItem]](t, DecodeFlowEnvelope(t, auditResp).Data)

	if err := env.DB.Where("challenge_id = ?", challenge.ID).Order("id ASC").Find(&result.Submissions).Error; err != nil {
		t.Fatalf("load submissions: %v", err)
	}

	return result
}

func AssertTimelineHasSubmit(t *testing.T, events []FlowTimelineEvent, challengeID int64, isCorrect bool, points int) {
	t.Helper()

	for _, event := range events {
		if event.Type != "flag_submit" || event.ChallengeID != challengeID || event.IsCorrect == nil || *event.IsCorrect != isCorrect {
			continue
		}
		if isCorrect && (event.Points == nil || *event.Points != points) {
			t.Fatalf("expected correct timeline event to include %d points, got %+v", points, event)
		}
		if event.Detail == "" {
			t.Fatalf("expected timeline submit event to include detail, got %+v", event)
		}
		return
	}

	t.Fatalf("expected timeline to contain submit event challenge_id=%d is_correct=%t", challengeID, isCorrect)
}

func AssertTeacherEvidenceHasEvent(t *testing.T, events []FlowTeacherEvidenceReviewEvent, wantType string, challengeID int64, metaKey, metaValue string) {
	t.Helper()

	for _, event := range events {
		if event.Type != wantType || event.ChallengeID != challengeID {
			continue
		}
		value, ok := event.Meta[metaKey]
		if !ok {
			t.Fatalf("expected evidence event %s to contain meta key %s: %+v", wantType, metaKey, event)
		}
		if value != metaValue {
			t.Fatalf("expected evidence event %s meta[%s]=%s, got %+v", wantType, metaKey, metaValue, event.Meta)
		}
		return
	}
	t.Fatalf("expected evidence to contain event type=%s challenge_id=%d", wantType, challengeID)
}

func AssertTimelineHasChallengeDetailView(t *testing.T, events []FlowTimelineEvent, challengeID int64) {
	t.Helper()
	for _, event := range events {
		if event.Type == "challenge_detail_view" && event.ChallengeID == challengeID {
			if event.Detail == "" {
				t.Fatalf("expected challenge detail view to include detail, got %+v", event)
			}
			return
		}
	}
	t.Fatalf("expected timeline to contain challenge detail view event challenge_id=%d", challengeID)
}

func AssertTimelineHasInstanceAccess(t *testing.T, events []FlowTimelineEvent, challengeID int64) {
	t.Helper()
	for _, event := range events {
		if event.Type == "instance_access" && event.ChallengeID == challengeID {
			if event.Detail == "" {
				t.Fatalf("expected instance access event to include detail, got %+v", event)
			}
			return
		}
	}
	t.Fatalf("expected timeline to contain instance access event challenge_id=%d", challengeID)
}

func AssertTimelineHasProxyTrace(t *testing.T, events []FlowTimelineEvent, challengeID int64) {
	t.Helper()
	for _, event := range events {
		if event.Type == "instance_proxy_request" && event.ChallengeID == challengeID {
			if !strings.Contains(event.Detail, "经平台代理发起") {
				t.Fatalf("expected proxy trace event detail, got %+v", event)
			}
			return
		}
	}
	t.Fatalf("expected timeline to contain proxy trace event challenge_id=%d", challengeID)
}
