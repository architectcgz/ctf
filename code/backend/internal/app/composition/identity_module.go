package composition

import (
	authinfra "ctf-platform/internal/module/auth/infrastructure"
	identityruntime "ctf-platform/internal/module/identity/runtime"
)

type IdentityModule = identityruntime.Module

func BuildIdentityModule(root *Root) (*IdentityModule, error) {
	return identityruntime.Build(identityruntime.Deps{
		Config:       root.Config(),
		Logger:       root.Logger(),
		DB:           root.DB(),
		Cache:        root.Cache(),
		TokenService: authinfra.NewTokenService(root.Config().Auth, root.Config().WebSocket, root.Cache()),
	}), nil
}
