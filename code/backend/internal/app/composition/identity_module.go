package composition

import (
	authinfra "ctf-platform/internal/module/auth/infrastructure"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	identitycmd "ctf-platform/internal/module/identity/application/commands"
	identityqry "ctf-platform/internal/module/identity/application/queries"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
	jwtpkg "ctf-platform/pkg/jwt"
)

type IdentityModule struct {
	AdminHandler    *identityhttp.Handler
	ProfileCommands identitycontracts.ProfileCommandService
	ProfileQueries  identitycontracts.ProfileQueryService
	TokenService    identitycontracts.Authenticator

	users identitycontracts.UserRepository
}

type identityModuleDeps struct {
	users        identitycontracts.UserRepository
	tokenService identitycontracts.Authenticator
}

func BuildIdentityModule(root *Root) (*IdentityModule, error) {
	cfg := root.Config()
	log := root.Logger()

	jwtManager, err := jwtpkg.NewManager(cfg.Auth, cfg.App.Name)
	if err != nil {
		return nil, err
	}

	deps := buildIdentityModuleDeps(root, jwtManager)
	adminCommandService := identitycmd.NewAdminService(deps.users, log.Named("identity_admin_command_service"))
	adminQueryService := identityqry.NewAdminService(deps.users, cfg.Pagination, log.Named("identity_admin_query_service"))
	profileCommandService := identitycmd.NewProfileService(deps.users, log.Named("identity_profile_command_service"))
	profileQueryService := identityqry.NewProfileService(deps.users)

	return &IdentityModule{
		AdminHandler:    identityhttp.NewHandler(adminCommandService, adminQueryService),
		ProfileCommands: profileCommandService,
		ProfileQueries:  profileQueryService,
		TokenService:    deps.tokenService,
		users:           deps.users,
	}, nil
}

func buildIdentityModuleDeps(root *Root, jwtManager *jwtpkg.Manager) identityModuleDeps {
	cfg := root.Config()
	return identityModuleDeps{
		users:        identityinfra.NewRepository(root.DB()),
		tokenService: identitycmd.NewAuthenticatorService(authinfra.NewTokenService(cfg.Auth, cfg.WebSocket, root.Cache(), jwtManager)),
	}
}
