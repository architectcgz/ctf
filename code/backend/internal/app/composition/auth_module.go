package composition

import (
	authhttp "ctf-platform/internal/module/auth/api/http"
	authcmd "ctf-platform/internal/module/auth/application/commands"
	authqry "ctf-platform/internal/module/auth/application/queries"
)

type AuthModule struct {
	Handler *authhttp.Handler
}

func BuildAuthModule(root *Root, ops *OpsModule, identity *IdentityModule) (*AuthModule, error) {
	cfg := root.Config()
	log := root.Logger()
	authService := authcmd.NewService(identity.users, identity.TokenService, cfg.RateLimit.Login, log.Named("auth_service"))
	casCommandService := authcmd.NewCASService(cfg.Auth.CAS, identity.users, identity.TokenService, log.Named("cas_command_service"), nil)
	casQueryService := authqry.NewCASService(cfg.Auth.CAS)

	return &AuthModule{
		Handler: authhttp.NewHandler(
			authService,
			identity.ProfileCommands,
			identity.ProfileQueries,
			identity.TokenService,
			casCommandService,
			casQueryService,
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
