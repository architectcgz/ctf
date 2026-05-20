package queries

import (
	"time"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
	commonmapper "ctf-platform/internal/shared/mapperhelper"
)

func toAdminUserResp(user *identitycontracts.User) identitycontracts.AdminUser {
	resp := adminUserMapper.ToAdminUserRespPtr(user)
	resp.Name = commonmapper.NormalizeOptionalTrimmedString(user.Name)
	resp.Email = commonmapper.NormalizeOptionalTrimmedString(user.Email)
	resp.StudentNo = commonmapper.NormalizeOptionalTrimmedString(user.StudentNo)
	resp.TeacherNo = commonmapper.NormalizeOptionalTrimmedString(user.TeacherNo)
	resp.ClassName = commonmapper.NormalizeOptionalTrimmedString(user.ClassName)
	resp.Roles = []string{user.Role}
	resp.UpdatedAt = copyTimeToPtr(user.UpdatedAt)
	return *resp
}

func copyTimeToPtr(value time.Time) *time.Time {
	copied := value
	return &copied
}
