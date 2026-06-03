package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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
	identitycmd "ctf-platform/internal/module/identity/application/commands"
	identityqry "ctf-platform/internal/module/identity/application/queries"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
	instancecmd "ctf-platform/internal/module/instance/application/commands"
	instanceqry "ctf-platform/internal/module/instance/application/queries"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsqry "ctf-platform/internal/module/ops/application/queries"
	opsinfra "ctf-platform/internal/module/ops/infrastructure"
	practicehttp "ctf-platform/internal/module/practice/api/http"
	practicecmd "ctf-platform/internal/module/practice/application/commands"
	practiceqry "ctf-platform/internal/module/practice/application/queries"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	runtimehttp "ctf-platform/internal/module/runtime/api/http"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	teachingqueryhttp "ctf-platform/internal/module/teaching_query/api/http"
	teachingqueryqueries "ctf-platform/internal/module/teaching_query/application/queries"
	teachingqueryinfra "ctf-platform/internal/module/teaching_query/infrastructure"
	runtimeadapters "ctf-platform/internal/testutil/runtimeadapters"
	"ctf-platform/internal/validation"
)

type flowTestEnv struct {
	router  *gin.Engine
	db      *gorm.DB
	cache   *redislib.Client
	admin   *identitycontracts.User
	student *identitycontracts.User
	image   *appImageRow
}

type flowEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type flowLoginResponse struct {
	User struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"user"`
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

type flowChallengeResponse struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Difficulty  string `json:"difficulty"`
	Points      int    `json:"points"`
	ImageID     int64  `json:"image_id"`
	Status      string `json:"status"`
}

type flowChallengeListItem struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Category      string `json:"category"`
	Difficulty    string `json:"difficulty"`
	Points        int    `json:"points"`
	SolvedCount   int64  `json:"solved_count"`
	TotalAttempts int64  `json:"total_attempts"`
	IsSolved      bool   `json:"is_solved"`
}

type flowChallengeDetail struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Category      string `json:"category"`
	Difficulty    string `json:"difficulty"`
	Points        int    `json:"points"`
	NeedTarget    bool   `json:"need_target"`
	AttachmentURL string `json:"attachment_url"`
	Hints         []struct {
		Level   int    `json:"level"`
		Content string `json:"content"`
	} `json:"hints"`
	SolvedCount   int64 `json:"solved_count"`
	TotalAttempts int64 `json:"total_attempts"`
	IsSolved      bool  `json:"is_solved"`
}

type flowSubmissionResponse struct {
	IsCorrect bool   `json:"is_correct"`
	Message   string `json:"message"`
	Points    int    `json:"points"`
}

type flowSubmissionRecord struct {
	ID          int64  `json:"id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
	Answer      string `json:"answer"`
	SubmittedAt string `json:"submitted_at"`
}

type flowInstanceResponse struct {
	ID        int64  `json:"id"`
	AccessURL string `json:"access_url"`
	Status    string `json:"status"`
}

type flowPageResponse[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type flowProgressResponse struct {
	TotalScore  int `json:"total_score"`
	TotalSolved int `json:"total_solved"`
	Rank        int `json:"rank"`
}

type flowTimelineResponse struct {
	Events []struct {
		Type        string `json:"type"`
		ChallengeID int64  `json:"challenge_id"`
		Title       string `json:"title"`
		IsCorrect   *bool  `json:"is_correct"`
		Points      *int   `json:"points"`
		Detail      string `json:"detail"`
	} `json:"events"`
}

type flowTeacherEvidenceReviewResponse struct {
	Summary struct {
		TotalEvents       int   `json:"total_events"`
		ProxyRequestCount int   `json:"proxy_request_count"`
		SubmitCount       int   `json:"submit_count"`
		SuccessCount      int   `json:"success_count"`
		ChallengeID       int64 `json:"challenge_id"`
	} `json:"summary"`
	Events []struct {
		Type        string                 `json:"type"`
		ChallengeID int64                  `json:"challenge_id"`
		Title       string                 `json:"title"`
		Detail      string                 `json:"detail"`
		Meta        map[string]interface{} `json:"meta"`
	} `json:"events"`
}

type flowTeacherAttackSessionResponse struct {
	Summary struct {
		TotalSessions   int `json:"total_sessions"`
		SuccessCount    int `json:"success_count"`
		FailedCount     int `json:"failed_count"`
		InProgressCount int `json:"in_progress_count"`
		UnknownCount    int `json:"unknown_count"`
		EventCount      int `json:"event_count"`
	} `json:"summary"`
	Sessions []struct {
		ID          string `json:"id"`
		Mode        string `json:"mode"`
		ChallengeID *int64 `json:"challenge_id"`
		Result      string `json:"result"`
		EventCount  int    `json:"event_count"`
		Events      []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		} `json:"events"`
	} `json:"sessions"`
}

type flowAuditItem struct {
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *int64                 `json:"resource_id"`
	ActorUserID  *int64                 `json:"actor_user_id"`
	Detail       map[string]interface{} `json:"detail"`
}

func newPracticeFlowTestEnv(t *testing.T) *flowTestEnv {
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
	if err := cache.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("ping test redis: %v", err)
	}

	db := openInternalAppTestSQLite(t, "practice-flow.sqlite")

	cfg := newPracticeFlowTestConfig(t)
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

	practiceRepo := practiceinfra.NewRepository(db)
	instanceRepo := runtimeinfrarepo.NewRepository(db)
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
	runtimeInstanceCommands := instancecmd.NewInstanceService(instanceRepo, runtimeCleanupService, &cfg.Container, logger)
	runtimeInstanceQueries := instanceqry.NewInstanceService(instanceRepo, &cfg.Container)
	runtimeProxyTicketService := instanceqry.NewProxyTicketService(runtimeinfrarepo.NewProxyTicketStore(cache), instanceRepo, cfg.Container.ProxyTicketTTL)
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
	runtimeHandler := runtimehttp.NewHandler(runtimeService, cfg.Container.PublicHost, cfg.Container.AccessHost, auditCommandService, runtimehttp.CookieConfig{}, nil)

	admin := createFlowUser(t, db, "admin_user", "Password123", identitycontracts.RoleAdmin)
	student := createFlowUser(t, db, "student_user", "Password123", identitycontracts.RoleStudent)
	image := createFlowImage(t, db)

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

	return &flowTestEnv{
		router:  router,
		db:      db,
		cache:   cache,
		admin:   admin,
		student: student,
		image:   image,
	}
}

func newPracticeFlowTestConfig(t *testing.T) *config.Config {
	t.Helper()

	portRangeStart, portRangeEnd := reserveInternalAppTestPortRange(t, 101)

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

func createFlowImage(t *testing.T, db *gorm.DB) *appImageRow {
	t.Helper()

	image := &appImageRow{
		Name:   "ctf/web-basic",
		Tag:    "v1",
		Status: challengecontracts.ImageStatusAvailable,
	}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	return image
}

func loginForSession(t *testing.T, router http.Handler, username, password string) *http.Cookie {
	t.Helper()

	resp := performFlowJSONRequest(
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
	body := decodeFlowEnvelope(t, resp)
	_ = decodeFlowJSON[flowLoginResponse](t, body.Data)
	sessionCookie := cloneCookie(resp.Result().Cookies(), "ctf_session")
	if sessionCookie == nil {
		t.Fatalf("expected session cookie for %s", username)
	}
	return sessionCookie
}

func sessionHeaders(cookie *http.Cookie) map[string]string {
	if cookie == nil {
		return nil
	}
	return map[string]string{
		"Cookie": fmt.Sprintf("%s=%s", cookie.Name, cookie.Value),
	}
}

func loginForToken(t *testing.T, router http.Handler, username, password string) string {
	t.Helper()
	return loginForSession(t, router, username, password).Value
}

func bearerHeaders(token string) map[string]string {
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

func performFlowJSONRequest(
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

func decodeFlowEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) flowEnvelope {
	t.Helper()

	var envelope flowEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v body=%s", err, recorder.Body.String())
	}
	return envelope
}

func decodeFlowJSON[T any](t *testing.T, data []byte) T {
	t.Helper()

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatalf("decode payload: %v payload=%s", err, string(data))
	}
	return value
}

func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	return data
}

func assertTimelineHasSubmit(t *testing.T, events []struct {
	Type        string `json:"type"`
	ChallengeID int64  `json:"challenge_id"`
	Title       string `json:"title"`
	IsCorrect   *bool  `json:"is_correct"`
	Points      *int   `json:"points"`
	Detail      string `json:"detail"`
}, challengeID int64, isCorrect bool, points int) {
	t.Helper()

	for _, event := range events {
		if event.Type != "flag_submit" || event.ChallengeID != challengeID || event.IsCorrect == nil || *event.IsCorrect != isCorrect {
			continue
		}
		if isCorrect {
			if event.Points == nil || *event.Points != points {
				t.Fatalf("expected correct timeline event to include %d points, got %+v", points, event)
			}
		}
		if event.Detail == "" {
			t.Fatalf("expected timeline submit event to include detail, got %+v", event)
		}
		return
	}

	t.Fatalf("expected timeline to contain submit event challenge_id=%d is_correct=%t", challengeID, isCorrect)
}

func assertTeacherEvidenceHasEvent(
	t *testing.T,
	events []struct {
		Type        string                 `json:"type"`
		ChallengeID int64                  `json:"challenge_id"`
		Title       string                 `json:"title"`
		Detail      string                 `json:"detail"`
		Meta        map[string]interface{} `json:"meta"`
	},
	wantType string,
	challengeID int64,
	metaKey string,
	metaValue string,
) {
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

func assertTimelineHasChallengeDetailView(t *testing.T, events []struct {
	Type        string `json:"type"`
	ChallengeID int64  `json:"challenge_id"`
	Title       string `json:"title"`
	IsCorrect   *bool  `json:"is_correct"`
	Points      *int   `json:"points"`
	Detail      string `json:"detail"`
}, challengeID int64) {
	t.Helper()

	for _, event := range events {
		if event.Type != "challenge_detail_view" || event.ChallengeID != challengeID {
			continue
		}
		if event.Detail == "" {
			t.Fatalf("expected challenge detail view to include detail, got %+v", event)
		}
		return
	}

	t.Fatalf("expected timeline to contain challenge detail view event challenge_id=%d", challengeID)
}

func assertTimelineHasInstanceAccess(t *testing.T, events []struct {
	Type        string `json:"type"`
	ChallengeID int64  `json:"challenge_id"`
	Title       string `json:"title"`
	IsCorrect   *bool  `json:"is_correct"`
	Points      *int   `json:"points"`
	Detail      string `json:"detail"`
}, challengeID int64) {
	t.Helper()

	for _, event := range events {
		if event.Type != "instance_access" || event.ChallengeID != challengeID {
			continue
		}
		if event.Detail == "" {
			t.Fatalf("expected instance access event to include detail, got %+v", event)
		}
		return
	}

	t.Fatalf("expected timeline to contain instance access event challenge_id=%d", challengeID)
}

func assertTimelineHasProxyTrace(t *testing.T, events []struct {
	Type        string `json:"type"`
	ChallengeID int64  `json:"challenge_id"`
	Title       string `json:"title"`
	IsCorrect   *bool  `json:"is_correct"`
	Points      *int   `json:"points"`
	Detail      string `json:"detail"`
}, challengeID int64) {
	t.Helper()

	for _, event := range events {
		if event.Type != "instance_proxy_request" || event.ChallengeID != challengeID {
			continue
		}
		if !strings.Contains(event.Detail, "经平台代理发起") {
			t.Fatalf("expected proxy trace event detail, got %+v", event)
		}
		return
	}

	t.Fatalf("expected timeline to contain proxy trace event challenge_id=%d", challengeID)
}
