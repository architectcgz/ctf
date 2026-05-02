package queries

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	commonmapper "ctf-platform/internal/shared/mapperhelper"
)

func toAdminUserResp(user *model.User) dto.AdminUserResp {
	mapped := adminUserMapper.ToAdminUserResp(*user)
	mapped.Name = commonmapper.NormalizeOptionalTrimmedString(user.Name)
	mapped.Email = commonmapper.NormalizeOptionalTrimmedString(user.Email)
	mapped.StudentNo = commonmapper.NormalizeOptionalTrimmedString(user.StudentNo)
	mapped.TeacherNo = commonmapper.NormalizeOptionalTrimmedString(user.TeacherNo)
	mapped.ClassName = commonmapper.NormalizeOptionalTrimmedString(user.ClassName)
	mapped.Roles = commonmapper.SingleString(user.Role)
	mapped.UpdatedAt = commonmapper.CopyTimeToPtr(user.UpdatedAt)
	return mapped
}
