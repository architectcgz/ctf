package composition

import (
	"ctf-platform/internal/auditlog"
	authhttp "ctf-platform/internal/module/auth/api/http"
	authcmd "ctf-platform/internal/module/auth/application/commands"
	authqry "ctf-platform/internal/module/auth/application/queries"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
)

type AuthModule struct {
	Handler *authhttp.Handler
}

type authModuleDeps struct {
	users           *identityinfra.Repository
	tokenService    authcontracts.TokenService
	profileCommands identitycontracts.ProfileCommandService
	profileQueries  identitycontracts.ProfileQueryService
	auditRecorder   auditlog.Recorder
}

func BuildAuthModule(root *Root, ops *OpsModule, identity *IdentityModule) (*AuthModule, error) {
	cfg := root.Config()
	log := root.Logger()
	deps := buildAuthModuleDeps(ops, identity)
	authService := authcmd.NewService(deps.users, deps.tokenService, cfg.RateLimit.Login, log.Named("auth_service"))
	casCommandService := authcmd.NewCASService(cfg.Auth.CAS, deps.users, deps.tokenService, log.Named("cas_command_service"), nil)
	casQueryService := authqry.NewCASService(cfg.Auth.CAS)

	return &AuthModule{
		Handler: authhttp.NewHandler(
			authService,
			deps.profileCommands,
			deps.profileQueries,
			deps.tokenService,
			casCommandService,
			casQueryService,
			authhttp.CookieConfig{
				Name:     cfg.Auth.SessionCookieName,
				Path:     cfg.Auth.SessionCookiePath,
				Secure:   cfg.Auth.SessionCookieSecure,
				HTTPOnly: cfg.Auth.SessionCookieHTTPOnly,
				SameSite: cfg.Auth.CookieSameSite(),
				MaxAge:   cfg.Auth.SessionTTL,
			},
			log.Named("auth_handler"),
			deps.auditRecorder,
		),
	}, nil
}

func buildAuthModuleDeps(ops *OpsModule, identity *IdentityModule) authModuleDeps {
	return authModuleDeps{
		users:           identity.userRepo,
		tokenService:    identity.TokenService,
		profileCommands: identity.ProfileCommands,
		profileQueries:  identity.ProfileQueries,
		auditRecorder:   ops.AuditService,
	}
}
