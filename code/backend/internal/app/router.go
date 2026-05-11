package app

import (
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/config"
	healthHandler "ctf-platform/internal/handler/health"
	"ctf-platform/internal/middleware"
	"ctf-platform/internal/model"
	authinfra "ctf-platform/internal/module/auth/infrastructure"
	contesthttp "ctf-platform/internal/module/contest/api/http"
	healthService "ctf-platform/internal/service/health"
	"ctf-platform/internal/validation"
	ratelimitpkg "ctf-platform/pkg/ratelimit"
)

type routerRuntime struct {
	engine           *gin.Engine
	closers          []lifecycleComponent
	assessment       *composition.AssessmentModule
	containerRuntime *composition.ContainerRuntimeModule
	contest          *composition.ContestModule
	instance         *composition.InstanceModule
}

var (
	buildAuthModule              = composition.BuildAuthModule
	buildAssessmentModule        = composition.BuildAssessmentModule
	buildContainerRuntimeModule  = composition.BuildContainerRuntimeModule
	buildChallengeModule         = composition.BuildChallengeModule
	buildContestModule           = composition.BuildContestModule
	buildIdentityModule          = composition.BuildIdentityModule
	buildInstanceModule          = composition.BuildInstanceModule
	buildOpsModule               = composition.BuildOpsModule
	buildPracticeModule          = composition.BuildPracticeModule
	buildPracticeReadmodelModule = composition.BuildPracticeReadmodelModule
	buildTeachingReadmodelModule = composition.BuildTeachingReadmodelModule
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
	engine.Use(middleware.AccessLog(log))

	rateChecker := ratelimitpkg.NewChecker(cache, cfg.RateLimit.RedisKeyPrefix)

	healthSvc := healthService.NewService(cfg, db, cache)
	health := healthHandler.NewHandler(healthSvc)
	engine.GET("/health", health.Get)
	engine.GET("/health/db", health.GetDB)
	engine.GET("/health/redis", health.GetRedis)

	containerRuntimeModule := buildContainerRuntimeModule(root)
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

	opsModule.BuildNotificationHandler(root, tokenService)
	protected.GET("/notifications", opsModule.NotificationHandler.ListNotifications)
	protected.PUT("/notifications/:id/read", middleware.ParseInt64Param("id"), opsModule.NotificationHandler.MarkAsRead)
	engine.GET("/ws/notifications", opsModule.NotificationHandler.ServeWS)

	teacherOrAbove := protected.Group("/teacher")
	teacherOrAbove.Use(middleware.RequireRole(model.RoleTeacher))
	teacherOrAbove.GET("/ping", middleware.RoleGuardPing("teacher"))

	authoring := protected.Group("/authoring")
	authoring.Use(middleware.RequireRole(model.RoleTeacher))

	adminOnly := protected.Group("/admin")
	adminOnly.Use(middleware.RequireRole(model.RoleAdmin))
	adminOnly.GET("/ping", middleware.RoleGuardPing("admin"))
	challengeModule, err := buildChallengeModule(root, containerRuntimeModule, opsModule)
	if err != nil {
		return nil, err
	}
	assessmentModule := buildAssessmentModule(root, challengeModule)
	teachingReadmodelModule := buildTeachingReadmodelModule(root, assessmentModule)
	contestModule := buildContestModule(root, challengeModule, containerRuntimeModule)
	contestModule.BindRealtimeBroadcaster(opsModule.WebSocketManager)
	contestRealtimeHandler := contesthttp.NewRealtimeHandler(
		tokenService,
		opsModule.WebSocketManager,
		log.Named("contest_realtime_handler"),
	)
	practiceModule := buildPracticeModule(root, challengeModule, instanceModule, assessmentModule)
	practiceReadmodelModule := buildPracticeReadmodelModule(root)
	instanceModule.BuildHandler(root, opsModule)

	registerTeacherAuthoringRoutes(authoring, adminRouteDeps{
		identityHandler: identityModule.AdminHandler,
		auditLogger:     composition.NamedAuditLogger(log),
		auditRecorder:   opsModule.AuditService,
		assessment:      assessmentModule,
		challenge:       challengeModule,
		contest:         contestModule,
		ops:             opsModule,
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
	})
	registerUserRoutes(apiV1, protected, teacherOrAbove, userRouteDeps{
		auditLogger:       composition.NamedAuditLogger(log),
		auditRecorder:     opsModule.AuditService,
		assessment:        assessmentModule,
		challenge:         challengeModule,
		contest:           contestModule,
		practice:          practiceModule,
		practiceReadmodel: practiceReadmodelModule,
		instance:          instanceModule,
		teachingReadmodel: teachingReadmodelModule,
	})
	engine.GET("/ws/contests/:id/announcements", contestRealtimeHandler.ServeAnnouncementWS)
	engine.GET("/ws/contests/:id/scoreboard", contestRealtimeHandler.ServeScoreboardWS)
	engine.GET("/ws/contests/:id/awd-preview", contestRealtimeHandler.ServeAWDPreviewWS)

	return &routerRuntime{
		engine:           engine,
		assessment:       assessmentModule,
		containerRuntime: containerRuntimeModule,
		contest:          contestModule,
		instance:         instanceModule,
		closers: []lifecycleComponent{
			{name: "report_export_tasks", closer: assessmentModule.BackgroundTasks},
			{name: "image_cleanup_tasks", closer: challengeModule.BackgroundTasks},
			{name: "practice_async_tasks", closer: practiceModule.BackgroundTasks},
		},
	}, nil
}
