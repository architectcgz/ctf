package app

import (
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	healthHandler "ctf-platform/internal/handler/health"
	"ctf-platform/internal/middleware"
	"ctf-platform/internal/model"
	assessmentModule "ctf-platform/internal/module/assessment"
	authModule "ctf-platform/internal/module/auth"
	challengeModule "ctf-platform/internal/module/challenge"
	containerModule "ctf-platform/internal/module/container"
	contestModule "ctf-platform/internal/module/contest"
	practiceModule "ctf-platform/internal/module/practice"
	systemModule "ctf-platform/internal/module/system"
	healthService "ctf-platform/internal/service/health"
	"ctf-platform/internal/validation"
	jwtpkg "ctf-platform/pkg/jwt"
	ratelimitpkg "ctf-platform/pkg/ratelimit"
)

func NewRouter(cfg *config.Config, log *zap.Logger, db *gorm.DB, cache *redislib.Client) (*gin.Engine, error) {
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
	tokenService := authModule.NewTokenService(cfg.Auth, cache, jwtManager)
	authService := authModule.NewService(authRepository, tokenService, log.Named("auth_service"))
	auditRepo := systemModule.NewAuditRepository(db)
	auditService := systemModule.NewAuditService(auditRepo, cfg.Pagination, log.Named("audit_service"))
	authHandler := authModule.NewHandler(authService, tokenService, authModule.CookieConfig{
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

	protected := apiV1.Group("")
	protected.Use(middleware.Auth(tokenService))
	protected.POST("/auth/logout", authHandler.Logout)
	protected.GET("/auth/profile", authHandler.Profile)

	teacherOrAbove := protected.Group("/teacher")
	teacherOrAbove.Use(middleware.RequireRole(model.RoleTeacher))
	teacherOrAbove.GET("/ping", middleware.RoleGuardPing("teacher"))

	adminOnly := protected.Group("/admin")
	adminOnly.Use(middleware.RequireRole(model.RoleAdmin))
	adminOnly.GET("/ping", middleware.RoleGuardPing("admin"))
	auditHandler := systemModule.NewAuditHandler(auditService)
	auditLogger := log.Named("audit_middleware")
	dockerClient, dockerErr := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if dockerErr != nil {
		log.Warn("docker_client_init_failed", zap.Error(dockerErr))
	}

	// 镜像管理（仅管理员）
	challengeRepo := challengeModule.NewRepository(db)
	imageRepo := challengeModule.NewImageRepository(db)
	imageService := challengeModule.NewImageService(imageRepo, challengeRepo, dockerClient, cfg, log.Named("image_service"))
	imageHandler := challengeModule.NewImageHandler(imageService)
	adminOnly.POST("/images",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "image",
		}, auditLogger),
		imageHandler.CreateImage,
	)
	adminOnly.GET("/images", imageHandler.ListImages)
	adminOnly.GET("/images/:id", imageHandler.GetImage)
	adminOnly.PUT("/images/:id",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "image",
			ResourceIDParam: "id",
		}, auditLogger),
		imageHandler.UpdateImage,
	)
	adminOnly.DELETE("/images/:id",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "image",
			ResourceIDParam: "id",
		}, auditLogger),
		imageHandler.DeleteImage,
	)

	// 靶场管理（仅管理员）
	challengeConfig := &challengeModule.Config{
		SolvedCountCacheTTL: cfg.Challenge.SolvedCountCacheTTL,
	}
	challengeService := challengeModule.NewService(challengeRepo, imageRepo, cache, challengeConfig, log.Named("challenge_service"))
	challengeHandler := challengeModule.NewHandler(challengeService)
	adminOnly.POST("/challenges",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "challenge",
		}, auditLogger),
		challengeHandler.CreateChallenge,
	)
	adminOnly.GET("/challenges", challengeHandler.ListChallenges)
	adminOnly.GET("/challenges/:id", challengeHandler.GetChallenge)
	adminOnly.PUT("/challenges/:id",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge",
			ResourceIDParam: "id",
		}, auditLogger),
		challengeHandler.UpdateChallenge,
	)
	adminOnly.DELETE("/challenges/:id",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "challenge",
			ResourceIDParam: "id",
		}, auditLogger),
		challengeHandler.DeleteChallenge,
	)
	adminOnly.PUT("/challenges/:id/publish",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionAdminOp,
			ResourceType:    "challenge",
			ResourceIDParam: "id",
		}, auditLogger),
		challengeHandler.PublishChallenge,
	)

	flagService, err := challengeModule.NewFlagService(db)
	if err != nil {
		return nil, err
	}
	flagHandler := challengeModule.NewFlagHandler(flagService)
	adminOnly.PUT("/challenges/:id/flag",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge_flag",
			ResourceIDParam: "id",
		}, auditLogger),
		flagHandler.ConfigureFlag,
	)
	adminOnly.GET("/challenges/:id/flag", flagHandler.GetFlagConfig)
	adminOnly.GET("/audit-logs", auditHandler.ListAuditLogs)

	containerRepoForDashboard := containerModule.NewRepository(db)
	if dockerErr != nil {
		log.Warn("dashboard_docker_client_unavailable", zap.Error(dockerErr))
	}
	dashboardService := systemModule.NewDashboardService(
		containerRepoForDashboard,
		dockerClient,
		cache,
		cfg,
		log.Named("dashboard_service"),
	)
	dashboardHandler := systemModule.NewDashboardHandler(dashboardService)
	adminOnly.GET("/dashboard", dashboardHandler.GetDashboard)

	assessmentRepo := assessmentModule.NewRepository(db)
	assessmentService := assessmentModule.NewService(assessmentRepo, cache, cfg.Assessment, log.Named("assessment_service"))
	recommendationService := assessmentModule.NewRecommendationService(assessmentRepo, challengeRepo, cache, cfg.Recommendation, log.Named("recommendation_service"))
	assessmentHandler := assessmentModule.NewHandler(assessmentService, recommendationService)
	reportRepo := assessmentModule.NewReportRepository(db)
	reportService := assessmentModule.NewReportService(reportRepo, assessmentService, cfg.Report, log.Named("report_service"))
	reportHandler := assessmentModule.NewReportHandler(reportService)

	// 竞赛管理
	contestRepo := contestModule.NewRepository(db)
	scoreboardService := contestModule.NewScoreboardService(contestRepo, cache, &cfg.Contest, log.Named("contest_scoreboard_service"))
	contestService := contestModule.NewService(contestRepo, log.Named("contest_service"))
	contestHandler := contestModule.NewHandler(contestService, scoreboardService)
	contestChallengeRepo := contestModule.NewChallengeRepository(db)
	contestChallengeService := contestModule.NewChallengeService(contestChallengeRepo, challengeRepo, contestRepo)
	contestChallengeHandler := contestModule.NewChallengeHandler(contestChallengeService)
	teamRepo := contestModule.NewTeamRepository(db)
	teamService := contestModule.NewTeamService(teamRepo, contestRepo)
	teamHandler := contestModule.NewTeamHandler(teamService)
	submissionService := contestModule.NewSubmissionService(db, cache, flagService, teamRepo, scoreboardService, cfg)
	submissionHandler := contestModule.NewSubmissionHandler(submissionService)

	adminOnly.POST("/contests",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "contest",
		}, auditLogger),
		contestHandler.CreateContest,
	)
	adminOnly.PUT("/contests/:id",
		middleware.ParseInt64Param("id"),
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "contest",
			ResourceIDParam: "id",
		}, auditLogger),
		contestHandler.UpdateContest,
	)
	adminOnly.GET("/contests", contestHandler.ListContests)
	adminOnly.POST("/contests/:id/freeze",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionAdminOp,
			ResourceType:    "contest",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id"),
		}, auditLogger),
		contestHandler.FreezeScoreboard,
	)
	adminOnly.POST("/contests/:id/unfreeze",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionAdminOp,
			ResourceType:    "contest",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id"),
		}, auditLogger),
		contestHandler.UnfreezeScoreboard,
	)
	adminOnly.GET("/contests/:id/challenges", contestChallengeHandler.ListAdminChallenges)
	adminOnly.POST("/contests/:id/challenges",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "contest_challenge",
			DetailBuilder: middleware.DetailFromParams("id"),
		}, auditLogger),
		contestChallengeHandler.AddChallenge,
	)
	adminOnly.PUT("/contests/:id/challenges/:cid",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "contest_challenge",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}, auditLogger),
		contestChallengeHandler.UpdatePoints,
	)
	adminOnly.DELETE("/contests/:id/challenges/:cid",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "contest_challenge",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}, auditLogger),
		contestChallengeHandler.RemoveChallenge,
	)

	contestGroup := apiV1.Group("/contests")
	contestGroup.GET("", contestHandler.ListContests)
	contestGroup.GET("/:id", middleware.ParseInt64Param("id"), contestHandler.GetContest)
	contestGroup.GET("/:id/scoreboard", contestHandler.GetScoreboard)
	protected.GET("/contests/:id/challenges", contestChallengeHandler.ListChallenges)
	protected.POST("/contests/:id/challenges/:cid/submissions",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionSubmit,
			ResourceType:    "contest_submission",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}, auditLogger),
		submissionHandler.SubmitFlag,
	)
	protected.GET("/contests/:id/teams", teamHandler.ListTeams)
	protected.POST("/contests/:id/teams",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "team",
			DetailBuilder: middleware.DetailFromParams("id"),
		}, auditLogger),
		teamHandler.CreateTeam,
	)
	protected.POST("/contests/:id/teams/:tid/join",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "team_membership",
			ResourceIDParam: "tid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid"),
		}, auditLogger),
		teamHandler.JoinTeam,
	)
	protected.DELETE("/contests/:id/teams/:tid/leave",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "team_membership",
			ResourceIDParam: "tid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid"),
		}, auditLogger),
		teamHandler.LeaveTeam,
	)
	protected.DELETE("/contests/:id/teams/:tid",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "team",
			ResourceIDParam: "tid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid"),
		}, auditLogger),
		teamHandler.DismissTeam,
	)

	// 实践模块（学员）
	practiceRepo := practiceModule.NewRepository(db)
	instanceRepo := containerModule.NewRepository(db)
	var containerEngine *containerModule.Engine
	if engine, err := containerModule.NewEngine(&cfg.Container); err != nil {
		log.Warn("container_engine_init_failed", zap.Error(err))
	} else {
		containerEngine = engine
	}
	containerService := containerModule.NewService(instanceRepo, containerEngine, &cfg.Container, log.Named("container_service"))
	scoreService := practiceModule.NewScoreService(db, cache, log.Named("score_service"), &cfg.Score)
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
	protected.POST("/challenges/:id/instances",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "instance",
			DetailBuilder: middleware.DetailFromParams("id"),
		}, auditLogger),
		practiceHandler.StartChallenge,
	)
	protected.POST("/challenges/:id/submit",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionSubmit,
			ResourceType:    "challenge_submission",
			ResourceIDParam: "id",
		}, auditLogger),
		practiceHandler.SubmitFlag,
	)
	protected.GET("/instances", practiceHandler.ListUserInstances)
	protected.GET("/instances/:id", practiceHandler.GetInstance)

	containerHandler := containerModule.NewHandler(containerService)
	protected.DELETE("/instances/:id",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}, auditLogger),
		containerHandler.DestroyInstance,
	)
	protected.POST("/instances/:id/extend",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}, auditLogger),
		containerHandler.ExtendInstance,
	)

	usersGroup := protected.Group("/users")
	usersGroup.GET("/me/progress", practiceHandler.GetProgress)
	usersGroup.GET("/me/timeline", practiceHandler.GetTimeline)
	usersGroup.GET("/me/skill-profile", assessmentHandler.GetMySkillProfile)
	usersGroup.GET("/me/recommendations", assessmentHandler.GetRecommendations)
	usersGroup.GET("/:id/skill-profile", middleware.RequireRole(model.RoleTeacher), assessmentHandler.GetStudentSkillProfile)
	teacherOrAbove.GET("/students/:id/skill-profile", assessmentHandler.GetStudentSkillProfile)
	protected.POST("/reports/personal", reportHandler.CreatePersonalReport)
	protected.GET("/reports/:id/download", reportHandler.DownloadReport)
	protected.POST("/reports/class", middleware.RequireRole(model.RoleTeacher), reportHandler.CreateClassReport)
	teacherOrAbove.POST("/reports/class", reportHandler.CreateClassReport)

	return engine, nil
}
