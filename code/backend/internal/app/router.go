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
	healthService "ctf-platform/internal/service/health"
	"ctf-platform/internal/validation"
	ratelimitpkg "ctf-platform/pkg/ratelimit"
)

type routerRuntime struct {
	engine     *gin.Engine
	closers    []lifecycleComponent
	assessment *composition.AssessmentModule
	contest    *composition.ContestModule
	runtime    *composition.RuntimeModule
}

var (
	buildAuthModule              = composition.BuildAuthModule
	buildAssessmentModule        = composition.BuildAssessmentModule
	buildChallengeModule         = composition.BuildChallengeModule
	buildContestModule           = composition.BuildContestModule
	buildIdentityModule          = composition.BuildIdentityModule
	buildOpsModule               = composition.BuildOpsModule
	buildPracticeModule          = composition.BuildPracticeModule
	buildPracticeReadmodelModule = composition.BuildPracticeReadmodelModule
	buildRuntimeModule           = composition.BuildRuntimeModule
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
	if cfg.RateLimit.Global.Enabled {
		engine.Use(middleware.RateLimitByIP(rateChecker, "global", cfg.RateLimit.Global.Limit, cfg.RateLimit.Global.Window))
	}

	healthSvc := healthService.NewService(cfg, db, cache)
	health := healthHandler.NewHandler(healthSvc)
	engine.GET("/health", health.Get)
	engine.GET("/health/db", health.GetDB)
	engine.GET("/health/redis", health.GetRedis)

	runtimeModule := buildRuntimeModule(root)
	opsModule := buildOpsModule(root, runtimeModule)

	identityModule, err := buildIdentityModule(root)
	if err != nil {
		return nil, err
	}

	authModule, err := buildAuthModule(root, opsModule, identityModule)
	if err != nil {
		return nil, err
	}

	apiV1 := engine.Group("/api/v1")
	apiV1.GET("/health", health.Get)
	apiV1.GET("/health/db", health.GetDB)
	apiV1.GET("/health/redis", health.GetRedis)

	authGroup := apiV1.Group("/auth")
	if cfg.RateLimit.Login.Enabled {
		authGroup.Use(middleware.RateLimitByIP(rateChecker, "auth", cfg.RateLimit.Login.Limit, cfg.RateLimit.Login.Window))
	}
	authGroup.POST("/register", authModule.Handler.Register)
	authGroup.POST("/login", authModule.Handler.Login)
	authGroup.POST("/refresh", authModule.Handler.Refresh)
	authGroup.GET("/cas/status", authModule.Handler.CASStatus)
	authGroup.GET("/cas/login", authModule.Handler.CASLogin)
	authGroup.GET("/cas/callback", authModule.Handler.CASCallback)

	protected := apiV1.Group("")
	protected.Use(middleware.Auth(identityModule.TokenService))
	protected.POST("/auth/logout", authModule.Handler.Logout)
	protected.GET("/auth/profile", authModule.Handler.Profile)
	protected.PUT("/auth/password", authModule.Handler.ChangePassword)
	protected.POST("/auth/ws-ticket", authModule.Handler.IssueWSTicket)

	opsModule.BuildNotificationHandler(root, identityModule.TokenService)
	protected.GET("/notifications", opsModule.NotificationHandler.ListNotifications)
	protected.PUT("/notifications/:id/read", middleware.ParseInt64Param("id"), opsModule.NotificationHandler.MarkAsRead)
	engine.GET("/ws/notifications", opsModule.NotificationHandler.ServeWS)

	teacherOrAbove := protected.Group("/teacher")
	teacherOrAbove.Use(middleware.RequireRole(model.RoleTeacher))
	teacherOrAbove.GET("/ping", middleware.RoleGuardPing("teacher"))

	adminAuthoring := protected.Group("/admin")
	adminAuthoring.Use(middleware.RequireRole(model.RoleTeacher))

	adminOnly := protected.Group("/admin")
	adminOnly.Use(middleware.RequireRole(model.RoleAdmin))
	adminOnly.GET("/ping", middleware.RoleGuardPing("admin"))
	challengeModule, err := buildChallengeModule(root, runtimeModule)
	if err != nil {
		return nil, err
	}
	assessmentModule := buildAssessmentModule(root, challengeModule)
	teachingReadmodelModule := buildTeachingReadmodelModule(root, assessmentModule)
	contestModule := buildContestModule(root, challengeModule, runtimeModule)
	practiceModule := buildPracticeModule(root, challengeModule, runtimeModule, assessmentModule)
	practiceReadmodelModule := buildPracticeReadmodelModule(root)
	runtimeModule.BuildHandler(root, opsModule)

	registerTeacherAuthoringRoutes(adminAuthoring, adminRouteDeps{
		identityHandler: identityModule.AdminHandler,
		auditLogger:     composition.NamedAuditLogger(log),
		auditRecorder:   opsModule.AuditService,
		challenge:       challengeModule,
		contest:         contestModule,
		ops:             opsModule,
	})
	registerAdminRoutes(adminOnly, adminRouteDeps{
		identityHandler: identityModule.AdminHandler,
		auditLogger:     composition.NamedAuditLogger(log),
		auditRecorder:   opsModule.AuditService,
		challenge:       challengeModule,
		contest:         contestModule,
		ops:             opsModule,
	})
	registerUserRoutes(apiV1, protected, teacherOrAbove, userRouteDeps{
		auditLogger:       composition.NamedAuditLogger(log),
		auditRecorder:     opsModule.AuditService,
		assessment:        assessmentModule,
		challenge:         challengeModule,
		contest:           contestModule,
		practice:          practiceModule,
		practiceReadmodel: practiceReadmodelModule,
		runtime:           runtimeModule,
		teachingReadmodel: teachingReadmodelModule,
	})

	return &routerRuntime{
		engine:     engine,
		assessment: assessmentModule,
		contest:    contestModule,
		runtime:    runtimeModule,
		closers: []lifecycleComponent{
			{name: "report_service", closer: assessmentModule.BackgroundCloser},
			{name: "image_service", closer: challengeModule.BackgroundCloser},
			{name: "practice_service", closer: practiceModule.BackgroundCloser},
		},
	}, nil
}
