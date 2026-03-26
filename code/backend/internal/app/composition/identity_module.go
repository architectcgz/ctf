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

func BuildIdentityModule(root *Root) (*IdentityModule, error) {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	jwtManager, err := jwtpkg.NewManager(cfg.Auth, cfg.App.Name)
	if err != nil {
		return nil, err
	}

	users := identityinfra.NewRepository(db)
	tokenService := identitycmd.NewAuthenticatorService(authinfra.NewTokenService(cfg.Auth, cfg.WebSocket, cache, jwtManager))
	adminCommandService := identitycmd.NewAdminService(users, log.Named("identity_admin_command_service"))
	adminQueryService := identityqry.NewAdminService(users, cfg.Pagination, log.Named("identity_admin_query_service"))
	profileCommandService := identitycmd.NewProfileService(users, log.Named("identity_profile_command_service"))
	profileQueryService := identityqry.NewProfileService(users)

	return &IdentityModule{
		AdminHandler:    identityhttp.NewHandler(adminCommandService, adminQueryService),
		ProfileCommands: profileCommandService,
		ProfileQueries:  profileQueryService,
		TokenService:    tokenService,
		users:           users,
	}, nil
}
