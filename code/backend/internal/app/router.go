package app

import (
	"context"

	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/config"
	healthHandler "ctf-platform/internal/handler/health"
	ratelimitpkg "ctf-platform/internal/infrastructure/ratelimit"
	mcpinterface "ctf-platform/internal/interfaces/mcp"
	"ctf-platform/internal/middleware"
	authinfra "ctf-platform/internal/module/auth/infrastructure"
	contesthttp "ctf-platform/internal/module/contest/api/http"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/internal/platform/clustersecret"
	healthService "ctf-platform/internal/service/health"
	"ctf-platform/internal/validation"
)

type routerRuntime struct {
	engine           *gin.Engine
	readiness        *healthService.ReadinessState
	closers          []lifecycleComponent
	assessment       *composition.AssessmentModule
	containerRuntime *composition.ContainerRuntimeModule
	contest          *composition.ContestModule
	instance         *composition.InstanceModule
}

var (
	buildAuthModule             = composition.BuildAuthModule
	buildAssessmentModule       = composition.BuildAssessmentModule
	buildContainerRuntimeModule = composition.BuildContainerRuntimeModule
	buildChallengeModule        = composition.BuildChallengeModule
	buildContestModule          = composition.BuildContestModule
	buildIdentityModule         = composition.BuildIdentityModule
	buildInstanceModule         = composition.BuildInstanceModule
	buildOpsModule              = composition.BuildOpsModule
	buildPracticeModule         = composition.BuildPracticeModule
	buildTeachingAnalysisModule = composition.BuildTeachingAnalysisModule
)

func NewRouter(cfg *config.Config, log *zap.Logger, db *gorm.DB, cache *redislib.Client) (*gin.Engine, error) {
	root, err := composition.BuildRoot(cfg, log, db, cache)
	if err != nil {
		return nil, err
	}

	runtime, err := buildRouterRuntime(root)
	if err != nil {
		return nil, err
	}
	return runtime.engine, nil
}

func buildRouterRuntime(root *composition.Root) (*routerRuntime, error) {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	if err := validation.Register(); err != nil {
		return nil, err
	}

	engine := gin.New()
	engine.Use(middleware.Recovery(log))
	engine.Use(middleware.RequestID())
	engine.Use(middleware.CORS(cfg.CORS))
	engine.Use(middleware.SecurityHeaders())
	engine.Use(middleware.AccessLog(log))

	rateChecker := ratelimitpkg.NewChecker(cache, cfg.RateLimit.RedisKeyPrefix)

	readiness := healthService.NewReadinessState()
	healthSvc := healthService.NewService(cfg, db, cache, readiness, containerFlagSecretDependencyCheck(cfg, db)...)
	health := healthHandler.NewHandler(healthSvc)
	engine.GET("/live", health.GetLive)
	engine.GET("/ready", health.GetReady)
	engine.GET("/health", health.Get)
	engine.GET("/health/db", health.GetDB)
	engine.GET("/health/redis", health.GetRedis)

	containerRuntimeModule, err := buildContainerRuntimeModule(root)
	if err != nil {
		return nil, err
	}
	opsModule := buildOpsModule(root, containerRuntimeModule)
	instanceModule := buildInstanceModule(root, containerRuntimeModule)

	identityModule, err := buildIdentityModule(root)
	if err != nil {
		return nil, err
	}

	tokenService := authinfra.NewTokenService(cfg.Auth, cfg.WebSocket, cache)

	authModule, err := buildAuthModule(root, opsModule, identityModule, tokenService)
	if err != nil {
		return nil, err
	}

	apiV1 := engine.Group("/api/v1")
	apiV1.GET("/live", health.GetLive)
	apiV1.GET("/ready", health.GetReady)
	apiV1.GET("/health", health.Get)
	apiV1.GET("/health/db", health.GetDB)
	apiV1.GET("/health/redis", health.GetRedis)

	authGroup := apiV1.Group("/auth")
	if cfg.RateLimit.Anonymous.Enabled {
		authGroup.Use(middleware.RateLimitByIP(rateChecker, "auth:anonymous", cfg.RateLimit.Anonymous.Limit, cfg.RateLimit.Anonymous.Window))
	}
	authGroup.POST("/register", authModule.Handler.Register)
	loginHandlers := make([]gin.HandlerFunc, 0, 3)
	if cfg.RateLimit.LoginIP.Enabled {
		loginHandlers = append(loginHandlers, middleware.RateLimitByIP(rateChecker, "auth:login_ip", cfg.RateLimit.LoginIP.Limit, cfg.RateLimit.LoginIP.Window))
	}
	if cfg.RateLimit.Login.Enabled {
		loginHandlers = append(loginHandlers, middleware.RateLimitByLoginPrincipalAndIP(rateChecker, "auth:login_principal", cfg.RateLimit.Login.Limit, cfg.RateLimit.Login.Window))
	}
	loginHandlers = append(loginHandlers, authModule.Handler.Login)
	authGroup.POST("/login", loginHandlers...)
	authGroup.GET("/cas/status", authModule.Handler.CASStatus)
	authGroup.GET("/cas/login", authModule.Handler.CASLogin)
	authGroup.GET("/cas/callback", authModule.Handler.CASCallback)

	protected := apiV1.Group("")
	protected.Use(middleware.Auth(tokenService, cfg.Auth.SessionCookieName, identityModule.Users))
	if cfg.RateLimit.Global.Enabled {
		protected.Use(middleware.RateLimitByUser(rateChecker, "global", cfg.RateLimit.Global.Limit, cfg.RateLimit.Global.Window))
	}
	protected.POST("/auth/logout", authModule.Handler.Logout)
	protected.GET("/auth/profile", authModule.Handler.Profile)
	protected.PUT("/auth/password", authModule.Handler.ChangePassword)
	protected.POST("/auth/ws-ticket", authModule.Handler.IssueWSTicket)
	protected.POST("/auth/mcp-token", authModule.Handler.IssueMCPToken)

	opsModule.BuildNotificationHandler(root, tokenService)
	protected.GET("/notifications", opsModule.NotificationHandler.ListNotifications)
	protected.PUT("/notifications/:id/read", middleware.ParseInt64Param("id"), opsModule.NotificationHandler.MarkAsRead)
	engine.GET("/ws/notifications", opsModule.NotificationHandler.ServeWS)

	teacherOrAbove := protected.Group("/teacher")
	teacherOrAbove.Use(middleware.RequireRole(identitycontracts.RoleTeacher))
	teacherOrAbove.GET("/ping", middleware.RoleGuardPing("teacher"))

	authoring := protected.Group("/authoring")
	authoring.Use(middleware.RequireRole(identitycontracts.RoleTeacher))

	adminOnly := protected.Group("/admin")
	adminOnly.Use(middleware.RequireRole(identitycontracts.RoleAdmin))
	adminOnly.GET("/ping", middleware.RoleGuardPing("admin"))
	challengeModule, err := buildChallengeModule(root, containerRuntimeModule)
	if err != nil {
		return nil, err
	}
	assessmentModule := buildAssessmentModule(root, challengeModule)
	teachingAnalysisModule := buildTeachingAnalysisModule(root, assessmentModule, identityModule)
	contestModule := buildContestModule(root, challengeModule, containerRuntimeModule)
	contestRealtimeHandler := contesthttp.NewRealtimeHandler(
		tokenService,
		opsModule.WebSocketManager,
		log.Named("contest_realtime_handler"),
	)
	practiceModule := buildPracticeModule(root, challengeModule, instanceModule, containerRuntimeModule)
	composition.WireRuntimeNodeFailover(containerRuntimeModule, instanceModule, practiceModule)
	instanceModule.BuildHandler(root, opsModule)
	mcpHandler := mcpinterface.NewHandler(mcpinterface.Deps{
		Instances:  instanceModule.QueryService,
		Challenges: challengeModule.PublishedQuery,
		Tokens:     tokenService,
		LoginURL:   "/login",
		TokenURL:   "/api/v1/auth/mcp-token",
	})
	engine.POST("/mcp", mcpHandler.ServeHTTP)

	registerTeacherAuthoringRoutes(authoring, adminRouteDeps{
		identityHandler: identityModule.AdminHandler,
		auditLogger:     composition.NamedAuditLogger(log),
		auditRecorder:   opsModule.AuditService,
		assessment:      assessmentModule,
		challenge:       challengeModule,
		contest:         contestModule,
		ops:             opsModule,
		tokenService:    tokenService,
	})
	registerAdminRoutes(adminOnly, adminRouteDeps{
		identityHandler: identityModule.AdminHandler,
		auditLogger:     composition.NamedAuditLogger(log),
		auditRecorder:   opsModule.AuditService,
		assessment:      assessmentModule,
		challenge:       challengeModule,
		contest:         contestModule,
		ops:             opsModule,
		practice:        practiceModule,
		tokenService:    tokenService,
	})
	registerUserRoutes(apiV1, protected, teacherOrAbove, userRouteDeps{
		auditLogger:      composition.NamedAuditLogger(log),
		auditRecorder:    opsModule.AuditService,
		assessment:       assessmentModule,
		challenge:        challengeModule,
		contest:          contestModule,
		practice:         practiceModule,
		instance:         instanceModule,
		teachingAnalysis: teachingAnalysisModule,
	})
	engine.GET("/ws/contests/:id/announcements", contestRealtimeHandler.ServeAnnouncementWS)
	engine.GET("/ws/contests/:id/scoreboard", contestRealtimeHandler.ServeScoreboardWS)
	engine.GET("/ws/contests/:id/awd-preview", contestRealtimeHandler.ServeAWDPreviewWS)

	return &routerRuntime{
		engine:           engine,
		readiness:        readiness,
		assessment:       assessmentModule,
		containerRuntime: containerRuntimeModule,
		contest:          contestModule,
		instance:         instanceModule,
		closers: []lifecycleComponent{
			{name: "report_export_tasks", closer: assessmentModule.BackgroundTasks},
			{name: "image_cleanup_tasks", closer: challengeModule.BackgroundTasks},
			{name: "practice_async_tasks", closer: practiceModule.BackgroundTasks},
			{name: "runtime_execution_bridge", closer: containerRuntimeModule.LifecycleCloser},
		},
	}, nil
}

func containerFlagSecretDependencyCheck(cfg *config.Config, db *gorm.DB) []healthService.DependencyCheck {
	if cfg == nil || db == nil || cfg.Container.FlagGlobalSecret == "" {
		return nil
	}
	keyID := cfg.Container.ResolvedFlagSecretKeyID
	if keyID == "" {
		keyID = cfg.Container.FlagGlobalSecretKeyID
	}
	if keyID == "" {
		keyID = "default"
	}
	secret := cfg.Container.FlagGlobalSecret
	secrets := cfg.Container.ResolvedFlagSecrets
	if secrets == nil {
		secrets = map[string]string{keyID: secret}
	}
	return []healthService.DependencyCheck{{
		Name: "container_flag_secret",
		Check: func(ctx context.Context) error {
			requiredKeyIDs, err := clustersecret.RequiredContainerFlagSecretKeyIDs(ctx, db)
			if err != nil {
				return err
			}
			return clustersecret.CheckContainerFlagSecretKeyring(ctx, db, clustersecret.ContainerFlagSecretKeyring{
				ActiveKeyID:    keyID,
				ActiveSecret:   secret,
				Secrets:        secrets,
				RequiredKeyIDs: requiredKeyIDs,
			})
		},
	}}
}
