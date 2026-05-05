package runtime

import (
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	identitycmd "ctf-platform/internal/module/identity/application/commands"
	identityqry "ctf-platform/internal/module/identity/application/queries"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
)

type Module struct {
	AdminHandler    *identityhttp.Handler
	ProfileCommands identitycontracts.ProfileCommandService
	ProfileQueries  identitycontracts.ProfileQueryService
	TokenService    identitycontracts.Authenticator
	Users           identitycontracts.UserRepository
}

type Deps struct {
	Config       *config.Config
	Logger       *zap.Logger
	DB           *gorm.DB
	Cache        *redislib.Client
	TokenService authcontracts.TokenService
}

type moduleDeps struct {
	input        Deps
	users        identitycontracts.UserRepository
	tokenService identitycontracts.Authenticator
}

func Build(deps Deps) *Module {
	internalDeps := newModuleDeps(deps)
	adminHandler, profileCommands, profileQueries := buildHandlers(internalDeps)

	return &Module{
		AdminHandler:    adminHandler,
		ProfileCommands: profileCommands,
		ProfileQueries:  profileQueries,
		TokenService:    internalDeps.tokenService,
		Users:           internalDeps.users,
	}
}

func newModuleDeps(deps Deps) moduleDeps {
	return moduleDeps{
		input:        deps,
		users:        identityinfra.NewRepository(deps.DB),
		tokenService: identitycmd.NewAuthenticatorService(deps.TokenService),
	}
}

func buildHandlers(deps moduleDeps) (*identityhttp.Handler, identitycontracts.ProfileCommandService, identitycontracts.ProfileQueryService) {
	log := deps.input.Logger
	cfg := deps.input.Config

	adminCommandService := identitycmd.NewAdminService(deps.users, log.Named("identity_admin_command_service"))
	adminQueryService := identityqry.NewAdminService(deps.users, cfg.Pagination, log.Named("identity_admin_query_service"))
	profileCommandService := identitycmd.NewProfileService(deps.users, log.Named("identity_profile_command_service"))
	profileQueryService := identityqry.NewProfileService(deps.users)

	return identityhttp.NewHandler(adminCommandService, adminQueryService), profileCommandService, profileQueryService
}
