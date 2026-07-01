package composition

import (
	"ctf-platform/internal/auditlog"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	authinfra "ctf-platform/internal/module/auth/infrastructure"
	authruntime "ctf-platform/internal/module/auth/runtime"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	identityinfra "ctf-platform/internal/module/identity/infrastructure"
)

type AuthModule struct {
	Handler *authruntime.Module
}

type authModuleDeps struct {
	users           *identityinfra.Repository
	tokenService    authcontracts.TokenService
	profileCommands identitycontracts.ProfileCommandService
	profileQueries  identitycontracts.ProfileQueryService
	auditRecorder   auditlog.Recorder
}

func BuildAuthModule(root *Root, ops *OpsModule, identity *IdentityModule, tokenService authcontracts.TokenService) (*authruntime.Module, error) {
	deps := buildAuthModuleDeps(ops, identity, tokenService)
	oauthStore := authinfra.NewOAuthStore(root.DB(), root.Cache(), root.Config().Auth.OAuth)
	return authruntime.Build(authruntime.Deps{
		Config:          root.Config(),
		Logger:          root.Logger(),
		Users:           deps.users,
		TokenService:    deps.tokenService,
		OAuthStore:      oauthStore,
		ProfileCommands: deps.profileCommands,
		ProfileQueries:  deps.profileQueries,
		AuditRecorder:   deps.auditRecorder,
	}), nil
}

func buildAuthModuleDeps(ops *OpsModule, identity *IdentityModule, tokenService authcontracts.TokenService) authModuleDeps {
	return authModuleDeps{
		users:           identity.userRepo,
		tokenService:    tokenService,
		profileCommands: identity.ProfileCommands,
		profileQueries:  identity.ProfileQueries,
		auditRecorder:   ops.AuditService,
	}
}
