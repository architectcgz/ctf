package composition

import (
	authruntime "ctf-platform/internal/module/auth/runtime"
)

type AuthModule = authruntime.Module

func BuildAuthModule(root *Root, ops *OpsModule, identity *IdentityModule) (*AuthModule, error) {
	return authruntime.Build(authruntime.Deps{
		Config:          root.Config(),
		Logger:          root.Logger(),
		Users:           identity.Users,
		TokenService:    identity.TokenService,
		ProfileCommands: identity.ProfileCommands,
		ProfileQueries:  identity.ProfileQueries,
		AuditRecorder:   ops.AuditService,
	}), nil
}
