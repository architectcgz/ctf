package app

import (
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	healthHandler "ctf-platform/internal/handler/health"
	"ctf-platform/internal/middleware"
	"ctf-platform/internal/model"
	authModule "ctf-platform/internal/module/auth"
	challengeModule "ctf-platform/internal/module/challenge"
	containerModule "ctf-platform/internal/module/container"
	practiceModule "ctf-platform/internal/module/practice"
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
	authHandler := authModule.NewHandler(authService, tokenService, authModule.CookieConfig{
		Name:     cfg.Auth.RefreshCookieName,
		Path:     cfg.Auth.RefreshCookiePath,
		Secure:   cfg.Auth.RefreshCookieSecure,
		HTTPOnly: cfg.Auth.RefreshCookieHTTPOnly,
		SameSite: cfg.Auth.CookieSameSite(),
		MaxAge:   cfg.Auth.RefreshTokenTTL,
	}, log.Named("auth_handler"))

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

	// 镜像管理（仅管理员）
	imageRepo := challengeModule.NewImageRepository(db)
	imageService := challengeModule.NewImageService(imageRepo, nil, cfg, log.Named("image_service"))
	imageHandler := challengeModule.NewImageHandler(imageService)
	adminOnly.POST("/images", imageHandler.CreateImage)
	adminOnly.GET("/images", imageHandler.ListImages)
	adminOnly.GET("/images/:id", imageHandler.GetImage)
	adminOnly.PUT("/images/:id", imageHandler.UpdateImage)
	adminOnly.DELETE("/images/:id", imageHandler.DeleteImage)

	// 靶场管理（仅管理员）
	challengeRepo := challengeModule.NewRepository(db)
	challengeService := challengeModule.NewService(challengeRepo, imageRepo, cache)
	challengeHandler := challengeModule.NewHandler(challengeService)
	adminOnly.POST("/challenges", challengeHandler.CreateChallenge)
	adminOnly.GET("/challenges", challengeHandler.ListChallenges)
	adminOnly.GET("/challenges/:id", challengeHandler.GetChallenge)
	adminOnly.PUT("/challenges/:id", challengeHandler.UpdateChallenge)
	adminOnly.DELETE("/challenges/:id", challengeHandler.DeleteChallenge)
	adminOnly.PUT("/challenges/:id/publish", challengeHandler.PublishChallenge)

	// Flag 提交（学员）
	practiceRepo := practiceModule.NewRepository(db)
	containerRepo := containerModule.NewRepository(db)
	practiceService := practiceModule.NewService(
		practiceRepo,
		challengeRepo,
		containerRepo,
		cache,
		log,
		cfg.Container.FlagGlobalSecret,
		cfg.RateLimit.FlagSubmit.Limit,
		cfg.RateLimit.FlagSubmit.Window,
		cfg.Cache.ProgressTTL,
	)
	practiceHandler := practiceModule.NewHandler(practiceService)

	challengeGroup := protected.Group("/challenges")
	challengeGroup.POST("/:id/submit", middleware.ParseChallengeID(), practiceHandler.SubmitFlag)

	// 个人进度（学员）
	usersGroup := protected.Group("/users")
	usersGroup.GET("/me/progress", practiceHandler.GetProgress)
	usersGroup.GET("/me/timeline", practiceHandler.GetTimeline)

	return engine, nil
}
