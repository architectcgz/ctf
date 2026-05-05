package composition

import (
	authinfra "ctf-platform/internal/module/auth/infrastructure"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
	identityruntime "ctf-platform/internal/module/identity/runtime"
)

type IdentityModule struct {
	AdminHandler    *identityhttp.Handler
	ProfileCommands identitycontracts.ProfileCommandService
	ProfileQueries  identitycontracts.ProfileQueryService
	TokenService    identitycontracts.Authenticator
	Users           identitycontracts.UserLookupRepository

	userRepo *identityinfra.Repository
}

type identityModuleDeps struct {
	users        *identityinfra.Repository
	tokenService identitycontracts.Authenticator
}

func BuildIdentityModule(root *Root) (*IdentityModule, error) {
	deps := buildIdentityModuleDeps(root)
	module := identityruntime.Build(identityruntime.Deps{
		Config:       root.Config(),
		Logger:       root.Logger(),
		DB:           root.DB(),
		Cache:        root.Cache(),
		TokenService: authinfra.NewTokenService(root.Config().Auth, root.Config().WebSocket, root.Cache()),
	})
	return &IdentityModule{
		AdminHandler:    module.AdminHandler,
		ProfileCommands: module.ProfileCommands,
		ProfileQueries:  module.ProfileQueries,
		TokenService:    deps.tokenService,
		Users:           deps.users,
		userRepo:        deps.users,
	}, nil
}

func buildIdentityModuleDeps(root *Root) identityModuleDeps {
	return identityModuleDeps{
		users:        identityinfra.NewRepository(root.DB()),
		tokenService: authinfra.NewTokenService(root.Config().Auth, root.Config().WebSocket, root.Cache()),
	}
}
