package queries

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	identityshared "ctf-platform/internal/module/identity/application/shared"
)

func toAdminUserResp(user *model.User) dto.AdminUserResp {
	mapped := adminUserMapper.ToAdminUserResp(*user)
	mapped.Name = identityshared.NormalizeOptionalString(user.Name)
	mapped.Email = identityshared.NormalizeOptionalString(user.Email)
	mapped.StudentNo = identityshared.NormalizeOptionalString(user.StudentNo)
	mapped.TeacherNo = identityshared.NormalizeOptionalString(user.TeacherNo)
	mapped.ClassName = identityshared.NormalizeOptionalString(user.ClassName)
	mapped.Roles = identityshared.SingleRole(user.Role)
	mapped.UpdatedAt = identityshared.CopyTimeToPtr(user.UpdatedAt)
	return mapped
}
