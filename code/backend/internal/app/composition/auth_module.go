package composition

import (
	authhttp "ctf-platform/internal/module/auth/api/http"
	authapp "ctf-platform/internal/module/auth/application"
)

type AuthModule struct {
	Handler *authhttp.Handler
}

func BuildAuthModule(root *Root, ops *OpsModule, identity *IdentityModule) (*AuthModule, error) {
	cfg := root.Config()
	log := root.Logger()
	authService := authapp.NewService(identity.users, identity.TokenService, cfg.RateLimit.Login, log.Named("auth_service"))
	casProvider := authapp.NewCASProvider(cfg.Auth.CAS, identity.users, identity.TokenService, log.Named("cas_provider"), nil)

	return &AuthModule{
		Handler: authhttp.NewHandler(
			authService,
			identity.ProfileService,
			identity.TokenService,
			casProvider,
			authhttp.CookieConfig{
				Name:     cfg.Auth.RefreshCookieName,
				Path:     cfg.Auth.RefreshCookiePath,
				Secure:   cfg.Auth.RefreshCookieSecure,
				HTTPOnly: cfg.Auth.RefreshCookieHTTPOnly,
				SameSite: cfg.Auth.CookieSameSite(),
				MaxAge:   cfg.Auth.RefreshTokenTTL,
			},
			log.Named("auth_handler"),
			ops.AuditService,
		),
	}, nil
}
