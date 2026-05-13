package runtime

import (
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/module/auth/api/http"
	authcmd "ctf-platform/internal/module/auth/application/commands"
	authqry "ctf-platform/internal/module/auth/application/queries"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	authinfra "ctf-platform/internal/module/auth/infrastructure"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
)

type Module struct {
	Handler *http.Handler
}

type Deps struct {
	Config          *config.Config
	Logger          *zap.Logger
	Users           identitycontracts.UserRepository
	TokenService    authcontracts.TokenService
	ProfileCommands identitycontracts.ProfileCommandService
	ProfileQueries  identitycontracts.ProfileQueryService
	AuditRecorder   auditlog.Recorder
}

type moduleDeps struct {
	input           Deps
	users           identitycontracts.UserRepository
	tokenService    authcontracts.TokenService
	profileCommands identitycontracts.ProfileCommandService
	profileQueries  identitycontracts.ProfileQueryService
	auditRecorder   auditlog.Recorder
}

func Build(deps Deps) *Module {
	internalDeps := newModuleDeps(deps)
	return &Module{
		Handler: buildHandler(internalDeps),
	}
}

func newModuleDeps(deps Deps) moduleDeps {
	return moduleDeps{
		input:           deps,
		users:           deps.Users,
		tokenService:    deps.TokenService,
		profileCommands: deps.ProfileCommands,
		profileQueries:  deps.ProfileQueries,
		auditRecorder:   deps.AuditRecorder,
	}
}

func buildHandler(deps moduleDeps) *http.Handler {
	cfg := deps.input.Config
	log := deps.input.Logger
	authService := authcmd.NewService(deps.users, deps.tokenService, cfg.RateLimit.Login, log.Named("auth_service"))
	casValidator := authinfra.NewCASTicketValidator(log.Named("cas_ticket_validator"), nil)
	casCommandService := authcmd.NewCASService(cfg.Auth.CAS, deps.users, deps.tokenService, log.Named("cas_command_service"), casValidator)
	casQueryService := authqry.NewCASService(cfg.Auth.CAS)

	return http.NewHandler(
		authService,
		deps.profileCommands,
		deps.profileQueries,
		deps.tokenService,
		casCommandService,
		casQueryService,
		http.CookieConfig{
			Name:     cfg.Auth.SessionCookieName,
			Path:     cfg.Auth.SessionCookiePath,
			Secure:   cfg.Auth.SessionCookieSecure,
			HTTPOnly: cfg.Auth.SessionCookieHTTPOnly,
			SameSite: cfg.Auth.CookieSameSite(),
			MaxAge:   cfg.Auth.SessionTTL,
		},
		log.Named("auth_handler"),
		deps.auditRecorder,
	)
}
