package queries

import (
	"strings"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

func toAdminUserResp(user *model.User) dto.AdminUserResp {
	var name *string
	if strings.TrimSpace(user.Name) != "" {
		name = &user.Name
	}
	var email *string
	if strings.TrimSpace(user.Email) != "" {
		email = &user.Email
	}
	var studentNo *string
	if strings.TrimSpace(user.StudentNo) != "" {
		studentNo = &user.StudentNo
	}
	var teacherNo *string
	if strings.TrimSpace(user.TeacherNo) != "" {
		teacherNo = &user.TeacherNo
	}
	var className *string
	if strings.TrimSpace(user.ClassName) != "" {
		className = &user.ClassName
	}
	updatedAt := user.UpdatedAt
	return dto.AdminUserResp{
		ID:        user.ID,
		Username:  user.Username,
		Name:      name,
		Email:     email,
		StudentNo: studentNo,
		TeacherNo: teacherNo,
		ClassName: className,
		Status:    user.Status,
		Roles:     []string{user.Role},
		CreatedAt: user.CreatedAt,
		UpdatedAt: &updatedAt,
	}
}
