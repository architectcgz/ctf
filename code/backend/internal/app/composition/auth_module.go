package composition

import (
	"ctf-platform/internal/auditlog"
	authruntime "ctf-platform/internal/module/auth/runtime"
	authcontracts "ctf-platform/internal/module/auth/contracts"
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

func BuildAuthModule(root *Root, ops *OpsModule, identity *IdentityModule) (*authruntime.Module, error) {
	deps := buildAuthModuleDeps(ops, identity)
	return authruntime.Build(authruntime.Deps{
		Config:          root.Config(),
		Logger:          root.Logger(),
		Users:           deps.users,
		TokenService:    deps.tokenService,
		ProfileCommands: deps.profileCommands,
		ProfileQueries:  deps.profileQueries,
		AuditRecorder:   deps.auditRecorder,
	}), nil
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
