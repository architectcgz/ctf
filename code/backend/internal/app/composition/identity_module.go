package composition

import (
	authinfra "ctf-platform/internal/module/auth/infrastructure"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	identitycmd "ctf-platform/internal/module/identity/application/commands"
	identityqry "ctf-platform/internal/module/identity/application/queries"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
)

type IdentityModule struct {
	AdminHandler    *identityhttp.Handler
	ProfileCommands identitycontracts.ProfileCommandService
	ProfileQueries  identitycontracts.ProfileQueryService
	TokenService    identitycontracts.Authenticator
	Users           identitycontracts.UserLookupRepository
	userRepo        *identityinfra.Repository
}

type identityModuleDeps struct {
	users        *identityinfra.Repository
	tokenService identitycontracts.Authenticator
}

func BuildIdentityModule(root *Root) (*IdentityModule, error) {
	cfg := root.Config()
	log := root.Logger()
	deps := buildIdentityModuleDeps(root)
	adminCommandService := identitycmd.NewAdminService(deps.users, log.Named("identity_admin_command_service"))
	adminQueryService := identityqry.NewAdminService(deps.users, cfg.Pagination, log.Named("identity_admin_query_service"))
	profileCommandService := identitycmd.NewProfileService(deps.users, log.Named("identity_profile_command_service"))
	profileQueryService := identityqry.NewProfileService(deps.users)

	return &IdentityModule{
		AdminHandler:    identityhttp.NewHandler(adminCommandService, adminQueryService),
		ProfileCommands: profileCommandService,
		ProfileQueries:  profileQueryService,
		TokenService:    deps.tokenService,
		Users:           deps.users,
		userRepo:        deps.users,
	}, nil
}

func buildIdentityModuleDeps(root *Root) identityModuleDeps {
	cfg := root.Config()
	return identityModuleDeps{
		users:        identityinfra.NewRepository(root.DB()),
		tokenService: identitycmd.NewAuthenticatorService(authinfra.NewTokenService(cfg.Auth, cfg.WebSocket, root.Cache())),
	}
}
