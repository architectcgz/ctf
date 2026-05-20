package commands

import (
	"errors"
	"strings"
	"time"

	"ctf-platform/internal/apperror"
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

func defaultUserStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return identitycontracts.UserStatusActive
	}
	return strings.TrimSpace(status)
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, identitycontracts.ErrUserNotFound):
		return apperror.ErrNotFound
	case errors.Is(err, identitycontracts.ErrUsernameExists):
		return identitycontracts.ErrDuplicateUsername
	case errors.Is(err, identitycontracts.ErrEmailExists):
		return identitycontracts.ErrDuplicateEmail
	case errors.Is(err, identitycontracts.ErrStudentNoExists):
		return identitycontracts.ErrDuplicateStudentNo
	case errors.Is(err, identitycontracts.ErrTeacherNoExists):
		return identitycontracts.ErrDuplicateTeacherNo
	case errors.Is(err, identitycontracts.ErrRoleNotFound):
		return apperror.ErrInternal.WithCause(err)
	default:
		return apperror.ErrInternal.WithCause(err)
	}
}

func looksLikeHeader(record []string) bool {
	return strings.EqualFold(strings.TrimSpace(getCSVValue(record, 0)), "username")
}

func isBlankRecord(record []string) bool {
	for _, item := range record {
		if strings.TrimSpace(item) != "" {
			return false
		}
	}
	return true
}

func getCSVValue(record []string, index int) string {
	if index < 0 || index >= len(record) {
		return ""
	}
	return record[index]
}

type identityNumbers struct {
	StudentNo string
	TeacherNo string
}

func normalizeIdentityNumbers(role, studentNo, teacherNo string) identityNumbers {
	normalized := identityNumbers{
		StudentNo: strings.TrimSpace(studentNo),
		TeacherNo: strings.TrimSpace(teacherNo),
	}

	switch strings.TrimSpace(role) {
	case identitycontracts.RoleStudent:
		normalized.TeacherNo = ""
	case identitycontracts.RoleTeacher:
		normalized.StudentNo = ""
	default:
		normalized.StudentNo = ""
		normalized.TeacherNo = ""
	}

	return normalized
}
