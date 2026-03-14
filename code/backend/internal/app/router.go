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
	writeupService := challengeModule.NewWriteupService(challengeRepo)
	writeupHandler := challengeModule.NewWriteupHandler(writeupService)
	templateRepo := challengeModule.NewTemplateRepository(db)
	topologyService := challengeModule.NewTopologyService(challengeRepo, templateRepo, imageRepo)
	topologyHandler := challengeModule.NewTopologyHandler(topologyService)
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
	adminOnly.GET("/challenges/:id/writeup", writeupHandler.GetAdmin)
	adminOnly.PUT("/challenges/:id/writeup",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge_writeup",
			ResourceIDParam: "id",
		}, auditLogger),
		writeupHandler.Upsert,
	)
	adminOnly.DELETE("/challenges/:id/writeup",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "challenge_writeup",
			ResourceIDParam: "id",
		}, auditLogger),
		writeupHandler.Delete,
	)
	adminOnly.GET("/challenges/:id/topology", topologyHandler.GetChallengeTopology)
	adminOnly.PUT("/challenges/:id/topology",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "challenge_topology",
			ResourceIDParam: "id",
		}, auditLogger),
		topologyHandler.SaveChallengeTopology,
	)
	adminOnly.DELETE("/challenges/:id/topology",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "challenge_topology",
			ResourceIDParam: "id",
		}, auditLogger),
		topologyHandler.DeleteChallengeTopology,
	)
	adminOnly.GET("/environment-templates", topologyHandler.ListTemplates)
	adminOnly.POST("/environment-templates",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "environment_template",
		}, auditLogger),
		topologyHandler.CreateTemplate,
	)
	adminOnly.GET("/environment-templates/:id", topologyHandler.GetTemplate)
	adminOnly.PUT("/environment-templates/:id",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "environment_template",
			ResourceIDParam: "id",
		}, auditLogger),
		topologyHandler.UpdateTemplate,
	)
	adminOnly.DELETE("/environment-templates/:id",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "environment_template",
			ResourceIDParam: "id",
		}, auditLogger),
		topologyHandler.DeleteTemplate,
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
	riskRepo := systemModule.NewRiskRepository(db)
	riskService := systemModule.NewRiskService(riskRepo, log.Named("risk_service"))
	riskHandler := systemModule.NewRiskHandler(riskService)
	adminOnly.GET("/cheat-detection", riskHandler.GetCheatDetection)

	adminUserRepo := adminUserModule.NewRepository(db)
	adminUserService := adminUserModule.NewService(adminUserRepo, cfg.Pagination, log.Named("admin_user_service"))
	adminUserHandler := adminUserModule.NewHandler(adminUserService)
	adminOnly.GET("/users", adminUserHandler.ListUsers)
	adminOnly.POST("/users",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "user",
		}, auditLogger),
		adminUserHandler.CreateUser,
	)
	adminOnly.PUT("/users/:id",
		middleware.ParseInt64Param("id"),
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "user",
			ResourceIDParam: "id",
		}, auditLogger),
		adminUserHandler.UpdateUser,
	)
	adminOnly.DELETE("/users/:id",
		middleware.ParseInt64Param("id"),
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "user",
			ResourceIDParam: "id",
		}, auditLogger),
		adminUserHandler.DeleteUser,
	)
	adminOnly.POST("/users/import",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:       model.AuditActionCreate,
			ResourceType: "user_import",
		}, auditLogger),
		adminUserHandler.ImportUsers,
	)

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
	awdService := contestModule.NewAWDService(db, contestRepo, cache, cfg.Container.FlagGlobalSecret, cfg.Contest.AWD, log.Named("contest_awd_service"))
	awdHandler := contestModule.NewAWDHandler(awdService)
	contestChallengeRepo := contestModule.NewChallengeRepository(db)
	contestChallengeService := contestModule.NewChallengeService(contestChallengeRepo, challengeRepo, contestRepo)
	contestChallengeHandler := contestModule.NewChallengeHandler(contestChallengeService)
	teamRepo := contestModule.NewTeamRepository(db)
	teamService := contestModule.NewTeamService(teamRepo, contestRepo)
	teamHandler := contestModule.NewTeamHandler(teamService)
	participationService := contestModule.NewParticipationService(db, contestRepo, teamRepo)
	participationHandler := contestModule.NewParticipationHandler(participationService)
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
	adminOnly.GET("/contests/:id/registrations", participationHandler.ListRegistrations)
	adminOnly.PUT("/contests/:id/registrations/:rid",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionUpdate,
			ResourceType:    "contest_registration",
			ResourceIDParam: "rid",
			DetailBuilder:   middleware.DetailFromParams("id", "rid"),
		}, auditLogger),
		participationHandler.ReviewRegistration,
	)
	adminOnly.GET("/contests/:id/announcements", participationHandler.ListAnnouncements)
	adminOnly.POST("/contests/:id/announcements",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "contest_announcement",
			DetailBuilder: middleware.DetailFromParams("id"),
		}, auditLogger),
		participationHandler.CreateAnnouncement,
	)
	adminOnly.DELETE("/contests/:id/announcements/:aid",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "contest_announcement",
			ResourceIDParam: "aid",
			DetailBuilder:   middleware.DetailFromParams("id", "aid"),
		}, auditLogger),
		participationHandler.DeleteAnnouncement,
	)
	adminOnly.GET("/contests/:id/awd/rounds",
		middleware.ParseInt64Param("id"),
		awdHandler.ListRounds,
	)
	adminOnly.GET("/contests/:id/scoreboard/live", contestHandler.GetLiveScoreboard)
	adminOnly.POST("/contests/:id/awd/rounds",
		middleware.ParseInt64Param("id"),
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "awd_round",
			DetailBuilder: middleware.DetailFromParams("id"),
		}, auditLogger),
		awdHandler.CreateRound,
	)
	adminOnly.POST("/contests/:id/awd/current-round/check",
		middleware.ParseInt64Param("id"),
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:        model.AuditActionUpdate,
			ResourceType:  "awd_checker_run",
			DetailBuilder: middleware.DetailFromParams("id"),
		}, auditLogger),
		awdHandler.RunCurrentRoundChecks,
	)
	adminOnly.POST("/contests/:id/awd/rounds/:rid/check",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:        model.AuditActionUpdate,
			ResourceType:  "awd_checker_run",
			DetailBuilder: middleware.DetailFromParams("id", "rid"),
		}, auditLogger),
		awdHandler.RunRoundChecks,
	)
	adminOnly.GET("/contests/:id/awd/rounds/:rid/services",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		awdHandler.ListServices,
	)
	adminOnly.POST("/contests/:id/awd/rounds/:rid/services/check",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:        model.AuditActionUpdate,
			ResourceType:  "awd_service_check",
			DetailBuilder: middleware.DetailFromParams("id", "rid"),
		}, auditLogger),
		awdHandler.UpsertServiceCheck,
	)
	adminOnly.GET("/contests/:id/awd/rounds/:rid/attacks",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		awdHandler.ListAttackLogs,
	)
	adminOnly.POST("/contests/:id/awd/rounds/:rid/attacks",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "awd_attack_log",
			DetailBuilder: middleware.DetailFromParams("id", "rid"),
		}, auditLogger),
		awdHandler.CreateAttackLog,
	)
	adminOnly.GET("/contests/:id/awd/rounds/:rid/summary",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("rid"),
		awdHandler.GetRoundSummary,
	)

	contestGroup := apiV1.Group("/contests")
	contestGroup.GET("", contestHandler.ListContests)
	contestGroup.GET("/:id", middleware.ParseInt64Param("id"), contestHandler.GetContest)
	contestGroup.GET("/:id/scoreboard", contestHandler.GetScoreboard)
	contestGroup.GET("/:id/announcements", participationHandler.ListAnnouncements)
	protected.POST("/contests/:id/register",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "contest_registration",
			DetailBuilder: middleware.DetailFromParams("id"),
		}, auditLogger),
		participationHandler.RegisterContest,
	)
	protected.GET("/contests/:id/challenges", contestChallengeHandler.ListChallenges)
	protected.GET("/contests/:id/my-progress", participationHandler.GetMyProgress)
	protected.POST("/contests/:id/challenges/:cid/submissions",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionSubmit,
			ResourceType:    "contest_submission",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}, auditLogger),
		submissionHandler.SubmitFlag,
	)
	protected.POST("/contests/:id/awd/challenges/:cid/submissions",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("cid"),
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionSubmit,
			ResourceType:    "awd_attack_submission",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}, auditLogger),
		awdHandler.SubmitAttack,
	)
	protected.GET("/contests/:id/teams", teamHandler.ListTeams)
	protected.GET("/contests/:id/my-team", teamHandler.GetMyTeam)
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
	protected.DELETE("/contests/:id/teams/:tid/members/:uid",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "team_membership",
			ResourceIDParam: "uid",
			DetailBuilder:   middleware.DetailFromParams("id", "tid", "uid"),
		}, auditLogger),
		teamHandler.KickMember,
	)

	// 实践模块（学员）
	practiceRepo := practiceModule.NewRepository(db)
	instanceRepo := containerModule.NewRepository(db)
	var containerEngine *containerModule.Engine
	if cfg.App.Env == "test" {
		log.Info("container_engine_disabled_in_test_env")
	} else if engine, err := containerModule.NewEngine(&cfg.Container); err != nil {
		log.Warn("container_engine_init_failed", zap.Error(err))
	} else {
		containerEngine = engine
	}
	containerService := containerModule.NewService(instanceRepo, containerEngine, &cfg.Container, log.Named("container_service"))
	proxyTicketService := containerModule.NewProxyTicketService(cache, &cfg.Container)
	scoreService := practiceModule.NewScoreService(db, cache, log.Named("score_service"), &cfg.Score)
	practiceService := practiceModule.NewService(
		db,
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
	protected.GET("/challenges", challengeHandler.ListPublishedChallenges)
	protected.GET("/challenges/:id",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionRead,
			ResourceType:    "challenge_detail",
			ResourceIDParam: "id",
		}, auditLogger),
		challengeHandler.GetPublishedChallenge,
	)
	protected.GET("/challenges/:id/writeup", writeupHandler.GetPublished)
	protected.POST("/challenges/:id/instances",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:        model.AuditActionCreate,
			ResourceType:  "instance",
			DetailBuilder: middleware.DetailFromParams("id"),
		}, auditLogger),
		practiceHandler.StartChallenge,
	)
	protected.POST("/contests/:id/challenges/:cid/instances",
		middleware.ParseInt64Param("id"),
		middleware.ParseInt64Param("cid"),
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionCreate,
			ResourceType:    "contest_instance",
			ResourceIDParam: "cid",
			DetailBuilder:   middleware.DetailFromParams("id", "cid"),
		}, auditLogger),
		practiceHandler.StartContestChallenge,
	)
	protected.POST("/challenges/:id/submit",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionSubmit,
			ResourceType:    "challenge_submission",
			ResourceIDParam: "id",
		}, auditLogger),
		practiceHandler.SubmitFlag,
	)
	protected.POST("/challenges/:id/hints/:level/unlock",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionCreate,
			ResourceType:    "challenge_hint_unlock",
			ResourceIDParam: "id",
			DetailBuilder:   middleware.DetailFromParams("id", "level"),
		}, auditLogger),
		practiceHandler.UnlockHint,
	)
	protected.GET("/instances", practiceHandler.ListUserInstances)
	protected.GET("/instances/:id", practiceHandler.GetInstance)

	containerHandler := containerModule.NewHandler(containerService, proxyTicketService, auditService, containerModule.ProxyCookieConfig{
		Secure:   cfg.Auth.RefreshCookieSecure,
		SameSite: cfg.Auth.CookieSameSite(),
	})
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
	protected.POST("/instances/:id/access", containerHandler.AccessInstance)
	apiV1.GET("/instances/:id/proxy", containerHandler.ProxyInstance)
	apiV1.Any("/instances/:id/proxy/*proxyPath", containerHandler.ProxyInstance)

	usersGroup := protected.Group("/users")
	usersGroup.GET("/me/progress", practiceHandler.GetProgress)
	usersGroup.GET("/me/timeline", practiceHandler.GetTimeline)
	usersGroup.GET("/me/skill-profile", assessmentHandler.GetMySkillProfile)
	usersGroup.GET("/me/recommendations", assessmentHandler.GetRecommendations)
	usersGroup.GET("/:id/skill-profile", middleware.RequireRole(model.RoleTeacher), assessmentHandler.GetStudentSkillProfile)
	teacherOrAbove.GET("/classes", teacherHandler.ListClasses)
	teacherOrAbove.GET("/classes/:name/students", teacherHandler.ListClassStudents)
	teacherOrAbove.GET("/classes/:name/summary", teacherHandler.GetClassSummary)
	teacherOrAbove.GET("/classes/:name/trend", teacherHandler.GetClassTrend)
	teacherOrAbove.GET("/classes/:name/review", teacherHandler.GetClassReview)
	teacherOrAbove.GET("/instances", containerHandler.ListTeacherInstances)
	teacherOrAbove.DELETE("/instances/:id",
		middleware.Audit(auditService, middleware.AuditOptions{
			Action:          model.AuditActionDelete,
			ResourceType:    "instance",
			ResourceIDParam: "id",
		}, auditLogger),
		containerHandler.DestroyTeacherInstance,
	)
	teacherOrAbove.GET("/students/:id/progress", teacherHandler.GetStudentProgress)
	teacherOrAbove.GET("/students/:id/skill-profile", assessmentHandler.GetStudentSkillProfile)
	teacherOrAbove.GET("/students/:id/recommendations", teacherHandler.GetStudentRecommendations)
	teacherOrAbove.GET("/students/:id/timeline", teacherHandler.GetStudentTimeline)
	protected.POST("/reports/personal", reportHandler.CreatePersonalReport)
	protected.GET("/reports/:id", reportHandler.GetReportStatus)
	protected.GET("/reports/:id/download", reportHandler.DownloadReport)
	protected.POST("/reports/class", middleware.RequireRole(model.RoleTeacher), reportHandler.CreateClassReport)
	teacherOrAbove.POST("/reports/class", reportHandler.CreateClassReport)

	return engine, nil
}
