package app

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/middleware"
	"ctf-platform/internal/model"
	authModule "ctf-platform/internal/module/auth"
	challengeModule "ctf-platform/internal/module/challenge"
	containerModule "ctf-platform/internal/module/container"
	practiceModule "ctf-platform/internal/module/practice"
	systemModule "ctf-platform/internal/module/system"
	"ctf-platform/internal/validation"
	"ctf-platform/pkg/errcode"
	jwtpkg "ctf-platform/pkg/jwt"
)

type flowTestEnv struct {
	router  *gin.Engine
	db      *gorm.DB
	cache   *redislib.Client
	admin   *model.User
	student *model.User
	image   *model.Image
}

type flowEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type flowLoginResponse struct {
	AccessToken string `json:"access_token"`
	User        struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"user"`
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
	SolvedCount   int64  `json:"solved_count"`
	TotalAttempts int64  `json:"total_attempts"`
	IsSolved      bool   `json:"is_solved"`
}

type flowSubmissionResponse struct {
	IsCorrect bool   `json:"is_correct"`
	Message   string `json:"message"`
	Points    int    `json:"points"`
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
	} `json:"events"`
}

type flowAuditItem struct {
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *int64                 `json:"resource_id"`
	ActorUserID  *int64                 `json:"actor_user_id"`
	Detail       map[string]interface{} `json:"detail"`
}

func TestPracticeFlow_AdminPublishesChallengeStudentSolvesChallenge(t *testing.T) {
	env := newPracticeFlowTestEnv(t)

	adminToken := loginForToken(t, env.router, "admin_user", "Password123")
	studentToken := loginForToken(t, env.router, "student_user", "Password123")

	createResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/admin/challenges",
		map[string]any{
			"title":       "Web SQLi 101",
			"description": "basic sql injection challenge",
			"category":    model.DimensionWeb,
			"difficulty":  model.ChallengeDifficultyEasy,
			"points":      100,
			"image_id":    env.image.ID,
		},
		bearerHeaders(adminToken),
		nil,
	)
	if createResp.Code != http.StatusOK {
		t.Fatalf("unexpected create challenge status: %d body=%s", createResp.Code, createResp.Body.String())
	}
	createBody := decodeFlowEnvelope(t, createResp)
	challenge := decodeFlowJSON[flowChallengeResponse](t, createBody.Data)

	configureFlagResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPut,
		"/api/v1/admin/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/flag",
		map[string]any{
			"flag_type": "static",
			"flag":      "flag{sqli_success}",
		},
		bearerHeaders(adminToken),
		nil,
	)
	if configureFlagResp.Code != http.StatusOK {
		t.Fatalf("unexpected configure flag status: %d body=%s", configureFlagResp.Code, configureFlagResp.Body.String())
	}

	publishResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPut,
		"/api/v1/admin/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/publish",
		nil,
		bearerHeaders(adminToken),
		nil,
	)
	if publishResp.Code != http.StatusOK {
		t.Fatalf("unexpected publish challenge status: %d body=%s", publishResp.Code, publishResp.Body.String())
	}

	listBeforeResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/challenges",
		nil,
		bearerHeaders(studentToken),
		nil,
	)
	if listBeforeResp.Code != http.StatusOK {
		t.Fatalf("unexpected list challenges status: %d body=%s", listBeforeResp.Code, listBeforeResp.Body.String())
	}
	listBeforeBody := decodeFlowEnvelope(t, listBeforeResp)
	listBefore := decodeFlowJSON[dto.PageResult](t, listBeforeBody.Data)
	listBeforeItems := decodeFlowJSON[[]flowChallengeListItem](t, mustMarshalJSON(t, listBefore.List))
	if len(listBeforeItems) != 1 {
		t.Fatalf("expected 1 published challenge, got %+v", listBeforeItems)
	}
	if listBeforeItems[0].IsSolved {
		t.Fatalf("expected challenge to be unsolved before submission")
	}

	detailResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10),
		nil,
		bearerHeaders(studentToken),
		nil,
	)
	if detailResp.Code != http.StatusOK {
		t.Fatalf("unexpected challenge detail status: %d body=%s", detailResp.Code, detailResp.Body.String())
	}
	detailBody := decodeFlowEnvelope(t, detailResp)
	detail := decodeFlowJSON[flowChallengeDetail](t, detailBody.Data)
	if detail.IsSolved {
		t.Fatalf("expected challenge detail to be unsolved before submission")
	}

	wrongSubmitResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/submit",
		map[string]any{"flag": "flag{wrong_answer}"},
		bearerHeaders(studentToken),
		nil,
	)
	if wrongSubmitResp.Code != http.StatusOK {
		t.Fatalf("unexpected wrong submit status: %d body=%s", wrongSubmitResp.Code, wrongSubmitResp.Body.String())
	}
	wrongSubmitBody := decodeFlowEnvelope(t, wrongSubmitResp)
	wrongSubmission := decodeFlowJSON[flowSubmissionResponse](t, wrongSubmitBody.Data)
	if wrongSubmission.IsCorrect {
		t.Fatalf("expected wrong flag submission to be incorrect")
	}

	correctSubmitResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/submit",
		map[string]any{"flag": "flag{sqli_success}"},
		bearerHeaders(studentToken),
		nil,
	)
	if correctSubmitResp.Code != http.StatusOK {
		t.Fatalf("unexpected correct submit status: %d body=%s", correctSubmitResp.Code, correctSubmitResp.Body.String())
	}
	correctSubmitBody := decodeFlowEnvelope(t, correctSubmitResp)
	correctSubmission := decodeFlowJSON[flowSubmissionResponse](t, correctSubmitBody.Data)
	if !correctSubmission.IsCorrect {
		t.Fatalf("expected correct flag submission to succeed")
	}
	if correctSubmission.Points != 100 {
		t.Fatalf("expected 100 points, got %d", correctSubmission.Points)
	}

	repeatSubmitResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/submit",
		map[string]any{"flag": "flag{sqli_success}"},
		bearerHeaders(studentToken),
		nil,
	)
	if repeatSubmitResp.Code != http.StatusConflict {
		t.Fatalf("expected repeated correct submission to return 409, got %d body=%s", repeatSubmitResp.Code, repeatSubmitResp.Body.String())
	}
	repeatSubmitBody := decodeFlowEnvelope(t, repeatSubmitResp)
	if repeatSubmitBody.Code != errcode.ErrAlreadySolved.Code {
		t.Fatalf("expected already solved code %d, got %d", errcode.ErrAlreadySolved.Code, repeatSubmitBody.Code)
	}

	listAfterResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/challenges",
		nil,
		bearerHeaders(studentToken),
		nil,
	)
	if listAfterResp.Code != http.StatusOK {
		t.Fatalf("unexpected post-submit list status: %d body=%s", listAfterResp.Code, listAfterResp.Body.String())
	}
	listAfterBody := decodeFlowEnvelope(t, listAfterResp)
	listAfter := decodeFlowJSON[dto.PageResult](t, listAfterBody.Data)
	listAfterItems := decodeFlowJSON[[]flowChallengeListItem](t, mustMarshalJSON(t, listAfter.List))
	if len(listAfterItems) != 1 {
		t.Fatalf("expected 1 challenge after submit, got %+v", listAfterItems)
	}
	if !listAfterItems[0].IsSolved {
		t.Fatalf("expected challenge to be solved after correct submission")
	}
	if listAfterItems[0].SolvedCount != 1 {
		t.Fatalf("expected solved_count 1, got %d", listAfterItems[0].SolvedCount)
	}
	if listAfterItems[0].TotalAttempts != 2 {
		t.Fatalf("expected total_attempts 2, got %d", listAfterItems[0].TotalAttempts)
	}

	progressResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/users/me/progress",
		nil,
		bearerHeaders(studentToken),
		nil,
	)
	if progressResp.Code != http.StatusOK {
		t.Fatalf("unexpected progress status: %d body=%s", progressResp.Code, progressResp.Body.String())
	}
	progressBody := decodeFlowEnvelope(t, progressResp)
	progress := decodeFlowJSON[flowProgressResponse](t, progressBody.Data)
	if progress.TotalSolved != 1 {
		t.Fatalf("expected total_solved 1, got %d", progress.TotalSolved)
	}
	if progress.TotalScore != 100 {
		t.Fatalf("expected total_score 100, got %d", progress.TotalScore)
	}
	if progress.Rank != 1 {
		t.Fatalf("expected rank 1, got %d", progress.Rank)
	}

	timelineResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/users/me/timeline",
		nil,
		bearerHeaders(studentToken),
		nil,
	)
	if timelineResp.Code != http.StatusOK {
		t.Fatalf("unexpected timeline status: %d body=%s", timelineResp.Code, timelineResp.Body.String())
	}
	timelineBody := decodeFlowEnvelope(t, timelineResp)
	timeline := decodeFlowJSON[flowTimelineResponse](t, timelineBody.Data)
	if len(timeline.Events) < 2 {
		t.Fatalf("expected at least two timeline events, got %+v", timeline.Events)
	}
	assertTimelineHasSubmit(t, timeline.Events, challenge.ID, false, 0)
	assertTimelineHasSubmit(t, timeline.Events, challenge.ID, true, 100)

	auditResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/admin/audit-logs?action=submit&resource_type=challenge_submission&user_id="+strconv.FormatInt(env.student.ID, 10),
		nil,
		bearerHeaders(adminToken),
		nil,
	)
	if auditResp.Code != http.StatusOK {
		t.Fatalf("unexpected audit status: %d body=%s", auditResp.Code, auditResp.Body.String())
	}
	auditBody := decodeFlowEnvelope(t, auditResp)
	auditPage := decodeFlowJSON[flowPageResponse[flowAuditItem]](t, auditBody.Data)
	if len(auditPage.List) != 2 {
		t.Fatalf("expected 2 submit audit logs, got %+v", auditPage.List)
	}

	var submissions []model.Submission
	if err := env.db.Order("submitted_at ASC").Find(&submissions).Error; err != nil {
		t.Fatalf("query submissions: %v", err)
	}
	if len(submissions) != 2 {
		t.Fatalf("expected 2 submission records, got %d", len(submissions))
	}
	if submissions[0].IsCorrect {
		t.Fatalf("expected first submission to be incorrect")
	}
	if !submissions[1].IsCorrect {
		t.Fatalf("expected second submission to be correct")
	}
}

func TestPracticeFlow_UnpublishedChallengeCannotBeSolved(t *testing.T) {
	env := newPracticeFlowTestEnv(t)

	adminToken := loginForToken(t, env.router, "admin_user", "Password123")
	studentToken := loginForToken(t, env.router, "student_user", "Password123")

	createResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/admin/challenges",
		map[string]any{
			"title":       "Draft Crypto",
			"description": "not published yet",
			"category":    model.DimensionCrypto,
			"difficulty":  model.ChallengeDifficultyMedium,
			"points":      150,
			"image_id":    env.image.ID,
		},
		bearerHeaders(adminToken),
		nil,
	)
	if createResp.Code != http.StatusOK {
		t.Fatalf("unexpected create challenge status: %d body=%s", createResp.Code, createResp.Body.String())
	}
	createBody := decodeFlowEnvelope(t, createResp)
	challenge := decodeFlowJSON[flowChallengeResponse](t, createBody.Data)

	configureFlagResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPut,
		"/api/v1/admin/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/flag",
		map[string]any{
			"flag_type": "static",
			"flag":      "flag{draft_secret}",
		},
		bearerHeaders(adminToken),
		nil,
	)
	if configureFlagResp.Code != http.StatusOK {
		t.Fatalf("unexpected configure draft flag status: %d body=%s", configureFlagResp.Code, configureFlagResp.Body.String())
	}

	listResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/challenges",
		nil,
		bearerHeaders(studentToken),
		nil,
	)
	if listResp.Code != http.StatusOK {
		t.Fatalf("unexpected list challenges status: %d body=%s", listResp.Code, listResp.Body.String())
	}
	listBody := decodeFlowEnvelope(t, listResp)
	listPage := decodeFlowJSON[dto.PageResult](t, listBody.Data)
	listItems := decodeFlowJSON[[]flowChallengeListItem](t, mustMarshalJSON(t, listPage.List))
	if len(listItems) != 0 {
		t.Fatalf("expected unpublished challenge to stay hidden, got %+v", listItems)
	}

	detailResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodGet,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10),
		nil,
		bearerHeaders(studentToken),
		nil,
	)
	if detailResp.Code != http.StatusForbidden {
		t.Fatalf("expected unpublished challenge detail to return 403, got %d body=%s", detailResp.Code, detailResp.Body.String())
	}

	submitResp := performFlowJSONRequest(
		t,
		env.router,
		http.MethodPost,
		"/api/v1/challenges/"+strconv.FormatInt(challenge.ID, 10)+"/submit",
		map[string]any{"flag": "flag{draft_secret}"},
		bearerHeaders(studentToken),
		nil,
	)
	if submitResp.Code != http.StatusForbidden {
		t.Fatalf("expected unpublished challenge submit to return 403, got %d body=%s", submitResp.Code, submitResp.Body.String())
	}
	submitBody := decodeFlowEnvelope(t, submitResp)
	if submitBody.Code != errcode.ErrChallengeNotPublish.Code {
		t.Fatalf("expected challenge not published code %d, got %d", errcode.ErrChallengeNotPublish.Code, submitBody.Code)
	}
}

func newPracticeFlowTestEnv(t *testing.T) *flowTestEnv {
	t.Helper()

	gin.SetMode(gin.TestMode)
	t.Setenv("CTF_FLAG_SECRET", "12345678901234567890123456789012")

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

	dbPath := filepath.Join(t.TempDir(), "practice-flow.sqlite")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Role{},
		&model.User{},
		&model.UserRole{},
		&model.AuditLog{},
		&model.Image{},
		&model.Challenge{},
		&model.Submission{},
		&model.Instance{},
		&model.SkillProfile{},
		&model.UserScore{},
	); err != nil {
		t.Fatalf("auto migrate test schema: %v", err)
	}

	cfg := newPracticeFlowTestConfig(t)
	logger := zap.NewNop()

	jwtManager, err := jwtpkg.NewManager(cfg.Auth, cfg.App.Name)
	if err != nil {
		t.Fatalf("create jwt manager: %v", err)
	}
	tokenService := authModule.NewTokenService(cfg.Auth, cache, jwtManager)
	authRepo := authModule.NewRepository(db)
	authService := authModule.NewService(authRepo, tokenService, logger)
	auditRepo := systemModule.NewAuditRepository(db)
	auditService := systemModule.NewAuditService(auditRepo, cfg.Pagination, logger)
	authHandler := authModule.NewHandler(authService, tokenService, authModule.CookieConfig{
		Name:     cfg.Auth.RefreshCookieName,
		Path:     cfg.Auth.RefreshCookiePath,
		Secure:   cfg.Auth.RefreshCookieSecure,
		HTTPOnly: cfg.Auth.RefreshCookieHTTPOnly,
		SameSite: cfg.Auth.CookieSameSite(),
		MaxAge:   cfg.Auth.RefreshTokenTTL,
	}, logger, auditService)
	auditHandler := systemModule.NewAuditHandler(auditService)

	challengeRepo := challengeModule.NewRepository(db)
	imageRepo := challengeModule.NewImageRepository(db)
	challengeService := challengeModule.NewService(challengeRepo, imageRepo, cache, &challengeModule.Config{
		SolvedCountCacheTTL: cfg.Challenge.SolvedCountCacheTTL,
	}, logger)
	challengeHandler := challengeModule.NewHandler(challengeService)

	flagService, err := challengeModule.NewFlagService(db)
	if err != nil {
		t.Fatalf("create flag service: %v", err)
	}
	flagHandler := challengeModule.NewFlagHandler(flagService)

	practiceRepo := practiceModule.NewRepository(db)
	instanceRepo := containerModule.NewRepository(db)
	practiceService := practiceModule.NewService(
		practiceRepo,
		challengeRepo,
		imageRepo,
		instanceRepo,
		nil,
		nil,
		nil,
		cache,
		cfg,
		logger,
	)
	practiceHandler := practiceModule.NewHandler(practiceService)

	admin := createFlowUser(t, db, "admin_user", "Password123", model.RoleAdmin)
	student := createFlowUser(t, db, "student_user", "Password123", model.RoleStudent)
	image := createFlowImage(t, db)

	router := gin.New()
	router.Use(middleware.RequestID())

	apiV1 := router.Group("/api/v1")
	authGroup := apiV1.Group("/auth")
	authGroup.POST("/login", authHandler.Login)

	protected := apiV1.Group("")
	protected.Use(middleware.Auth(tokenService))

	adminOnly := protected.Group("/admin")
	adminOnly.Use(middleware.RequireRole(model.RoleAdmin))
	adminOnly.POST("/challenges", challengeHandler.CreateChallenge)
	adminOnly.PUT("/challenges/:id/flag", flagHandler.ConfigureFlag)
	adminOnly.PUT("/challenges/:id/publish", challengeHandler.PublishChallenge)
	adminOnly.GET("/audit-logs", auditHandler.ListAuditLogs)

	protected.GET("/challenges", challengeHandler.ListPublishedChallenges)
	protected.GET("/challenges/:id", challengeHandler.GetPublishedChallenge)
	protected.POST("/challenges/:id/submit",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionSubmit,
			ResourceType:    "challenge_submission",
			ResourceIDParam: "id",
		}, logger),
		practiceHandler.SubmitFlag,
	)
	usersGroup := protected.Group("/users")
	usersGroup.GET("/me/progress", practiceHandler.GetProgress)
	usersGroup.GET("/me/timeline", practiceHandler.GetTimeline)

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

	privateKeyPath, publicKeyPath := writeFlowTestKeyPair(t)
	return &config.Config{
		App: config.AppConfig{
			Name: "ctf-platform-test",
			Env:  "test",
		},
		Auth: config.AuthConfig{
			Issuer:                "ctf-platform-test",
			AccessTokenTTL:        15 * time.Minute,
			RefreshTokenTTL:       24 * time.Hour,
			RefreshCookieName:     "refresh_token",
			RefreshCookiePath:     "/",
			RefreshCookieHTTPOnly: true,
			RefreshCookieSameSite: "lax",
			PrivateKeyPath:        privateKeyPath,
			PublicKeyPath:         publicKeyPath,
			TokenBlacklistPrefix:  "test:blacklist",
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
			FlagGlobalSecret: "12345678901234567890123456789012",
		},
		Assessment: config.AssessmentConfig{
			IncrementalUpdateDelay:   10 * time.Millisecond,
			IncrementalUpdateTimeout: time.Second,
		},
		Pagination: config.PaginationConfig{
			DefaultPageSize: 20,
			MaxPageSize:     100,
		},
	}
}

func writeFlowTestKeyPair(t *testing.T) (string, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicDER,
	})

	keyDir := t.TempDir()
	privatePath := filepath.Join(keyDir, "test_private.pem")
	publicPath := filepath.Join(keyDir, "test_public.pem")
	if err := os.WriteFile(privatePath, privatePEM, 0o600); err != nil {
		t.Fatalf("write private key: %v", err)
	}
	if err := os.WriteFile(publicPath, publicPEM, 0o644); err != nil {
		t.Fatalf("write public key: %v", err)
	}

	return privatePath, publicPath
}

func createFlowUser(t *testing.T, db *gorm.DB, username, password, role string) *model.User {
	t.Helper()

	user := &model.User{
		Username: username,
		Email:    fmt.Sprintf("%s@example.com", username),
		Role:     role,
		Status:   model.UserStatusActive,
	}
	if err := user.SetPassword(password); err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func createFlowImage(t *testing.T, db *gorm.DB) *model.Image {
	t.Helper()

	image := &model.Image{
		Name:   "ctf/web-basic",
		Tag:    "v1",
		Status: model.ImageStatusAvailable,
	}
	if err := db.Create(image).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	return image
}

func loginForToken(t *testing.T, router http.Handler, username, password string) string {
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
	login := decodeFlowJSON[flowLoginResponse](t, body.Data)
	if login.AccessToken == "" {
		t.Fatalf("expected access token for %s", username)
	}
	return login.AccessToken
}

func bearerHeaders(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
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
		return
	}

	t.Fatalf("expected timeline to contain submit event challenge_id=%d is_correct=%t", challengeID, isCorrect)
}
