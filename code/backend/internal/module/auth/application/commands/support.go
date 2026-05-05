package commands

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	commonmapper "ctf-platform/internal/shared/mapperhelper"
)

func buildAuthUser(user *model.User) dto.AuthUser {
	resp := authCommandResponseMapperInst.ToAuthUserBasePtr(user)
	resp.Name = commonmapper.NormalizeOptionalString(user.Name)
	resp.ClassName = commonmapper.NormalizeOptionalString(user.ClassName)
	return *resp
}
