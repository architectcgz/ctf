package composition

import (
	identityhttp "ctf-platform/internal/module/identity/api/http"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
	identityruntime "ctf-platform/internal/module/identity/runtime"
)

type IdentityModule struct {
	AdminHandler    *identityhttp.Handler
	ProfileCommands identitycontracts.ProfileCommandService
	ProfileQueries  identitycontracts.ProfileQueryService
	Users           identitycontracts.UserLookupRepository

	userRepo *identityinfra.Repository
}

type identityModuleDeps struct {
	users *identityinfra.Repository
}

func BuildIdentityModule(root *Root) (*IdentityModule, error) {
	deps := buildIdentityModuleDeps(root)
	module := identityruntime.Build(identityruntime.Deps{
		Config: root.Config(),
		Logger: root.Logger(),
		DB:     root.DB(),
	})
	return &IdentityModule{
		AdminHandler:    module.AdminHandler,
		ProfileCommands: module.ProfileCommands,
		ProfileQueries:  module.ProfileQueries,
		Users:           deps.users,
		userRepo:        deps.users,
	}, nil
}

func buildIdentityModuleDeps(root *Root) identityModuleDeps {
	return identityModuleDeps{
		users: identityinfra.NewRepository(root.DB()),
	}
}
