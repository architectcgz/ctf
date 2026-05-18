package commands

import (
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	commonmapper "ctf-platform/internal/shared/mapperhelper"
)

func buildAuthUser(user *identitycontracts.User) AuthUser {
	resp := authCommandResponseMapperInst.ToAuthUserBasePtr(&authUserSource{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
	})
	resp.Name = commonmapper.NormalizeOptionalString(user.Name)
	resp.ClassName = commonmapper.NormalizeOptionalString(user.ClassName)
	return *resp
}
