package queries

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

func toAdminUserResp(user *model.User) dto.AdminUserResp {
	mapped := adminUserMapper.ToAdminUserResp(*user)
	mapped.Name = normalizeOptionalString(user.Name)
	mapped.Email = normalizeOptionalString(user.Email)
	mapped.StudentNo = normalizeOptionalString(user.StudentNo)
	mapped.TeacherNo = normalizeOptionalString(user.TeacherNo)
	mapped.ClassName = normalizeOptionalString(user.ClassName)
	mapped.Roles = singleRole(user.Role)
	mapped.UpdatedAt = copyTimeToPtr(user.UpdatedAt)
	return mapped
}
