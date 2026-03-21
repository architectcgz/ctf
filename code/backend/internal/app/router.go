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
	adminUserModule "ctf-platform/internal/module/adminuser"
	assessmentModule "ctf-platform/internal/module/assessment"
	authModule "ctf-platform/internal/module/auth"
	challengeModule "ctf-platform/internal/module/challenge"
	containerModule "ctf-platform/internal/module/container"
	contestModule "ctf-platform/internal/module/contest"
	practiceModule "ctf-platform/internal/module/practice"
	systemModule "ctf-platform/internal/module/system"
	teacherModule "ctf-platform/internal/module/teacher"
	healthService "ctf-platform/internal/service/health"
	"ctf-platform/internal/validation"
	jwtpkg "ctf-platform/pkg/jwt"
	ratelimitpkg "ctf-platform/pkg/ratelimit"
	websocketpkg "ctf-platform/pkg/websocket"
)

type routerRuntime struct {
	engine            *gin.Engine
	closers           []lifecycleComponent
	containerService  *containerModule.Service
	assessmentService *assessmentModule.Service
}

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

	jwtManager, err := jwtpkg.NewManager(cfg.Auth, cfg.App.Name)
	if err != nil {
		return nil, err
	}

	authRepository := authModule.NewRepository(db)
	tokenService := authModule.NewTokenService(cfg.Auth, cfg.WebSocket, cache, jwtManager)
	authService := authModule.NewService(authRepository, tokenService, cfg.RateLimit.Login, log.Named("auth_service"))
	casProvider := authModule.NewCASProvider(cfg.Auth.CAS, authRepository, tokenService, log.Named("cas_provider"), nil)
	auditRepo := systemModule.NewAuditRepository(db)
	auditService := systemModule.NewAuditService(auditRepo, cfg.Pagination, log.Named("audit_service"))
	wsManager := websocketpkg.NewManager(cfg.WebSocket, log.Named("websocket_manager"))
	authHandler := authModule.NewHandler(authService, tokenService, casProvider, authModule.CookieConfig{
		Name:     cfg.Auth.RefreshCookieName,
		Path:     cfg.Auth.RefreshCookiePath,
		Secure:   cfg.Auth.RefreshCookieSecure,
		HTTPOnly: cfg.Auth.RefreshCookieHTTPOnly,
		SameSite: cfg.Auth.CookieSameSite(),
		MaxAge:   cfg.Auth.RefreshTokenTTL,
	}, log.Named("auth_handler"), auditService)

	apiV1 := engine.Group("/api/v1")
	apiV1.GET("/health", health.Get)
	apiV1.GET("/health/db", health.GetDB)
	apiV1.GET("/health/redis", health.GetRedis)

	authGroup := apiV1.Group("/auth")
	if cfg.RateLimit.Login.Enabled {
		authGroup.Use(middleware.RateLimitByIP(rateChecker, "auth", cfg.RateLimit.Login.Limit, cfg.RateLimit.Login.Window))
	}
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.Refresh)
	authGroup.GET("/cas/status", authHandler.CASStatus)
	authGroup.GET("/cas/login", authHandler.CASLogin)
	authGroup.GET("/cas/callback", authHandler.CASCallback)

	protected := apiV1.Group("")
	protected.Use(middleware.Auth(tokenService))
	protected.POST("/auth/logout", authHandler.Logout)
	protected.GET("/auth/profile", authHandler.Profile)
	protected.PUT("/auth/password", authHandler.ChangePassword)
	protected.POST("/auth/ws-ticket", authHandler.IssueWSTicket)

	notificationRepo := systemModule.NewNotificationRepository(db)
	notificationService := systemModule.NewNotificationService(notificationRepo, cfg.Pagination, wsManager, log.Named("notification_service"))
	notificationHandler := systemModule.NewNotificationHandler(notificationService, tokenService, wsManager, log.Named("notification_handler"))
	protected.GET("/notifications", notificationHandler.ListNotifications)
	protected.PUT("/notifications/:id/read", middleware.ParseInt64Param("id"), notificationHandler.MarkAsRead)
	engine.GET("/ws/notifications", notificationHandler.ServeWS)

	teacherOrAbove := protected.Group("/teacher")
	teacherOrAbove.Use(middleware.RequireRole(model.RoleTeacher))
	teacherOrAbove.GET("/ping", middleware.RoleGuardPing("teacher"))

	adminOnly := protected.Group("/admin")
	adminOnly.Use(middleware.RequireRole(model.RoleAdmin))
	adminOnly.GET("/ping", middleware.RoleGuardPing("admin"))
	auditHandler := systemModule.NewAuditHandler(auditService)
	auditLogger := log.Named("audit_middleware")

	containerRepo := containerModule.NewRepository(db)
	var containerEngine *containerModule.Engine
	if cfg.App.Env == "test" {
		log.Info("container_engine_disabled_in_test_env_for_router")
	} else if engine, err := containerModule.NewEngine(&cfg.Container); err != nil {
		log.Warn("container_engine_init_failed_for_router", zap.Error(err))
	} else {
		containerEngine = engine
	}
	containerService := containerModule.NewService(containerRepo, containerEngine, &cfg.Container, log.Named("container_service"))

	challengeRepo := challengeModule.NewRepository(db)
	imageRepo := challengeModule.NewImageRepository(db)
	imageService := challengeModule.NewImageService(imageRepo, challengeRepo, containerService, cfg, log.Named("image_service"))
	imageHandler := challengeModule.NewImageHandler(imageService)

	challengeConfig := &challengeModule.Config{
		SolvedCountCacheTTL: cfg.Challenge.SolvedCountCacheTTL,
	}
	challengeService := challengeModule.NewService(challengeRepo, imageRepo, cache, challengeConfig, log.Named("challenge_service"))
	challengeHandler := challengeModule.NewHandler(challengeService)
	writeupService := challengeModule.NewWriteupService(challengeRepo)
	writeupHandler := challengeModule.NewWriteupHandler(writeupService)
	templateRepo := challengeModule.NewTemplateRepository(db)
	topologyService := challengeModule.NewTopologyService(challengeRepo, templateRepo, imageRepo)
	topologyHandler := challengeModule.NewTopologyHandler(topologyService)

	flagService, err := challengeModule.NewFlagService(challengeRepo, cfg.Container.FlagGlobalSecret)
	if err != nil {
		return nil, err
	}
	flagHandler := challengeModule.NewFlagHandler(flagService)

	dashboardService := systemModule.NewDashboardService(
		containerRepo,
		containerService,
		cache,
		cfg,
		log.Named("dashboard_service"),
	)
	dashboardHandler := systemModule.NewDashboardHandler(dashboardService)
	riskRepo := systemModule.NewRiskRepository(db)
	riskService := systemModule.NewRiskService(riskRepo, log.Named("risk_service"))
	riskHandler := systemModule.NewRiskHandler(riskService)

	adminUserRepo := adminUserModule.NewRepository(db)
	adminUserService := adminUserModule.NewService(adminUserRepo, cfg.Pagination, log.Named("admin_user_service"))
	adminUserHandler := adminUserModule.NewHandler(adminUserService)

	assessmentRepo := assessmentModule.NewRepository(db)
	assessmentService := assessmentModule.NewService(assessmentRepo, cache, cfg.Assessment, log.Named("assessment_service"))
	recommendationService := assessmentModule.NewRecommendationService(assessmentRepo, challengeRepo, cache, cfg.Recommendation, log.Named("recommendation_service"))
	assessmentHandler := assessmentModule.NewHandler(assessmentService, recommendationService)
	teacherRepo := teacherModule.NewRepository(db)
	teacherService := teacherModule.NewService(teacherRepo, recommendationService, log.Named("teacher_service"))
	teacherHandler := teacherModule.NewHandler(teacherService)
	reportRepo := assessmentModule.NewReportRepository(db)
	reportService := assessmentModule.NewReportService(reportRepo, assessmentService, cfg.Report, log.Named("report_service"))
	reportHandler := assessmentModule.NewReportHandler(reportService)

	// 竞赛管理
	contestRepo := contestModule.NewRepository(db)
	scoreboardService := contestModule.NewScoreboardService(contestRepo, cache, &cfg.Contest, log.Named("contest_scoreboard_service"))
	contestService := contestModule.NewService(contestRepo, log.Named("contest_service"))
	contestHandler := contestModule.NewHandler(contestService, scoreboardService)
	awdService := contestModule.NewAWDService(contestModule.NewAWDRepository(db), contestRepo, cache, cfg.Container.FlagGlobalSecret, cfg.Contest.AWD, log.Named("contest_awd_service"))
	awdHandler := contestModule.NewAWDHandler(awdService)
	contestChallengeRepo := contestModule.NewChallengeRepository(db)
	contestChallengeService := contestModule.NewChallengeService(contestChallengeRepo, challengeRepo, contestRepo)
	contestChallengeHandler := contestModule.NewChallengeHandler(contestChallengeService)
	teamRepo := contestModule.NewTeamRepository(db)
	teamService := contestModule.NewTeamService(teamRepo, contestRepo)
	teamHandler := contestModule.NewTeamHandler(teamService)
	participationRepo := contestModule.NewParticipationRepository(db)
	participationService := contestModule.NewParticipationService(contestRepo, participationRepo, teamRepo)
	participationHandler := contestModule.NewParticipationHandler(participationService)
	submissionRepo := contestModule.NewSubmissionRepository(db)
	submissionService := contestModule.NewSubmissionService(contestRepo, submissionRepo, cache, flagService, teamRepo, scoreboardService, cfg)
	submissionHandler := contestModule.NewSubmissionHandler(submissionService)

	practiceRepo := practiceModule.NewRepository(db)
	instanceRepo := containerRepo
	proxyTicketService := containerModule.NewProxyTicketService(cache, &cfg.Container)
	scoreService := practiceModule.NewScoreService(practiceRepo, cache, log.Named("score_service"), &cfg.Score)
	practiceService := practiceModule.NewService(
		practiceRepo,
		challengeRepo,
		imageRepo,
		instanceRepo,
		containerService,
		scoreService,
		assessmentService,
		cache,
		cfg,
		log.Named("practice_service"),
	)
	practiceHandler := practiceModule.NewHandler(practiceService)

	containerHandler := containerModule.NewHandler(containerService, proxyTicketService, auditService, containerModule.ProxyCookieConfig{
		Secure:   cfg.Auth.RefreshCookieSecure,
		SameSite: cfg.Auth.CookieSameSite(),
	})
	registerAdminRoutes(adminOnly, adminRouteDeps{
		auditRecorder:           auditService,
		auditLogger:             auditLogger,
		imageHandler:            imageHandler,
		challengeHandler:        challengeHandler,
		writeupHandler:          writeupHandler,
		topologyHandler:         topologyHandler,
		flagHandler:             flagHandler,
		auditHandler:            auditHandler,
		dashboardHandler:        dashboardHandler,
		riskHandler:             riskHandler,
		adminUserHandler:        adminUserHandler,
		contestHandler:          contestHandler,
		awdHandler:              awdHandler,
		contestChallengeHandler: contestChallengeHandler,
		participationHandler:    participationHandler,
	})
	registerUserRoutes(apiV1, protected, teacherOrAbove, userRouteDeps{
		auditRecorder:           auditService,
		auditLogger:             auditLogger,
		challengeHandler:        challengeHandler,
		writeupHandler:          writeupHandler,
		practiceHandler:         practiceHandler,
		containerHandler:        containerHandler,
		assessmentHandler:       assessmentHandler,
		teacherHandler:          teacherHandler,
		reportHandler:           reportHandler,
		contestHandler:          contestHandler,
		awdHandler:              awdHandler,
		contestChallengeHandler: contestChallengeHandler,
		participationHandler:    participationHandler,
		submissionHandler:       submissionHandler,
		teamHandler:             teamHandler,
	})

	return &routerRuntime{
		engine:            engine,
		containerService:  containerService,
		assessmentService: assessmentService,
		closers: []lifecycleComponent{
			{name: "report_service", closer: reportService},
			{name: "image_service", closer: imageService},
			{name: "practice_service", closer: practiceService},
		},
	}, nil
}
