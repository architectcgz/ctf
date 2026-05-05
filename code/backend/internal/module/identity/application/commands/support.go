package commands

import (
	"errors"
	"strings"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	commonmapper "ctf-platform/internal/shared/mapperhelper"
	"ctf-platform/pkg/errcode"
)

func toAdminUserResp(user *model.User) dto.AdminUserResp {
	resp := adminUserMapper.ToAdminUserRespPtr(user)
	resp.Name = commonmapper.NormalizeOptionalTrimmedString(user.Name)
	resp.Email = commonmapper.NormalizeOptionalTrimmedString(user.Email)
	resp.StudentNo = commonmapper.NormalizeOptionalTrimmedString(user.StudentNo)
	resp.TeacherNo = commonmapper.NormalizeOptionalTrimmedString(user.TeacherNo)
	resp.ClassName = commonmapper.NormalizeOptionalTrimmedString(user.ClassName)
	resp.Roles = commonmapper.SingleString(user.Role)
	resp.UpdatedAt = commonmapper.CopyTimeToPtr(user.UpdatedAt)
	return *resp
}

func defaultUserStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return model.UserStatusActive
	}
	return strings.TrimSpace(status)
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, identitycontracts.ErrUserNotFound):
		return errcode.ErrNotFound
	case errors.Is(err, identitycontracts.ErrUsernameExists):
		return errcode.ErrUsernameExists
	case errors.Is(err, identitycontracts.ErrEmailExists):
		return errcode.ErrEmailExists
	case errors.Is(err, identitycontracts.ErrStudentNoExists):
		return errcode.ErrStudentNoExists
	case errors.Is(err, identitycontracts.ErrTeacherNoExists):
		return errcode.ErrTeacherNoExists
	case errors.Is(err, identitycontracts.ErrRoleNotFound):
		return errcode.ErrInternal.WithCause(err)
	default:
		return errcode.ErrInternal.WithCause(err)
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
	case model.RoleStudent:
		normalized.TeacherNo = ""
	case model.RoleTeacher:
		normalized.StudentNo = ""
	default:
		normalized.StudentNo = ""
		normalized.TeacherNo = ""
	}

	return normalized
}
