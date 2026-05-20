package queries

import (
	"time"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
)

func toAdminUserResp(user *identitycontracts.User) identitycontracts.AdminUser {
	resp := adminUserMapper.ToAdminUserRespPtr(user)
	resp.Name = normalizeOptionalTrimmedString(user.Name)
	resp.Email = normalizeOptionalTrimmedString(user.Email)
	resp.StudentNo = normalizeOptionalTrimmedString(user.StudentNo)
	resp.TeacherNo = normalizeOptionalTrimmedString(user.TeacherNo)
	resp.ClassName = normalizeOptionalTrimmedString(user.ClassName)
	resp.Roles = []string{user.Role}
	resp.UpdatedAt = copyTimeToPtr(user.UpdatedAt)
	return *resp
}

func copyTimeToPtr(value time.Time) *time.Time {
	copied := value
	return &copied
}
