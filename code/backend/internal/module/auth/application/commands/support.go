package commands

import (
	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

func buildAuthUser(user *identitycontracts.User) AuthUser {
	resp := authCommandResponseMapperInst.ToAuthUserBasePtr(&authUserSource{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
	resp.Name = normalizeOptionalString(user.Name)
	resp.ClassName = normalizeOptionalString(user.ClassName)
	return *resp
}
