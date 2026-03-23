package composition

import (
	authModule "ctf-platform/internal/module/auth"
)

type AuthModule struct {
	Handler *authModule.Handler
}

func BuildAuthModule(root *Root, system *SystemModule, identity *IdentityModule) (*AuthModule, error) {
	cfg := root.Config()
	log := root.Logger()
	authService := authModule.NewService(identity.users, identity.TokenService, cfg.RateLimit.Login, log.Named("auth_service"))
	casProvider := authModule.NewCASProvider(cfg.Auth.CAS, identity.users, identity.TokenService, log.Named("cas_provider"), nil)

	return &AuthModule{
		Handler: authModule.NewHandler(
			authService,
			identity.ProfileService,
			identity.TokenService,
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
	}, nil
}
