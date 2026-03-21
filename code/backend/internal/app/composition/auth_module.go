package composition

import (
	authModule "ctf-platform/internal/module/auth"
	jwtpkg "ctf-platform/pkg/jwt"
)

type AuthModule struct {
	Handler      *authModule.Handler
	TokenService authModule.TokenService
}

func BuildAuthModule(root *Root, system *SystemModule) (*AuthModule, error) {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	jwtManager, err := jwtpkg.NewManager(cfg.Auth, cfg.App.Name)
	if err != nil {
		return nil, err
	}

	authRepository := authModule.NewRepository(db)
	tokenService := authModule.NewTokenService(cfg.Auth, cfg.WebSocket, cache, jwtManager)
	authService := authModule.NewService(authRepository, tokenService, cfg.RateLimit.Login, log.Named("auth_service"))
	casProvider := authModule.NewCASProvider(cfg.Auth.CAS, authRepository, tokenService, log.Named("cas_provider"), nil)

	return &AuthModule{
		Handler: authModule.NewHandler(
			authService,
			tokenService,
			casProvider,
			authModule.CookieConfig{
				Name:     cfg.Auth.RefreshCookieName,
				Path:     cfg.Auth.RefreshCookiePath,
				Secure:   cfg.Auth.RefreshCookieSecure,
				HTTPOnly: cfg.Auth.RefreshCookieHTTPOnly,
				SameSite: cfg.Auth.CookieSameSite(),
				MaxAge:   cfg.Auth.RefreshTokenTTL,
			},
			log.Named("auth_handler"),
			system.AuditService,
		),
		TokenService: tokenService,
	}, nil
}
