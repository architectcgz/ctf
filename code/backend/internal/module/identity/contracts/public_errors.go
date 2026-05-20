package contracts

import (
	"net/http"

	"ctf-platform/internal/apperror"
)

var (
	ErrDuplicateUsername  = apperror.Define(11008, "用户名已存在", http.StatusConflict)
	ErrDuplicateEmail     = apperror.Define(11009, "邮箱已被注册", http.StatusConflict)
	ErrInvalidOldPassword = apperror.Define(11011, "原密码错误", http.StatusBadRequest)
	ErrPasswordReuse      = apperror.Define(11012, "新密码不能与原密码相同", http.StatusBadRequest)
	ErrDuplicateStudentNo = apperror.Define(11013, "学号已存在", http.StatusConflict)
	ErrDuplicateTeacherNo = apperror.Define(11014, "工号已存在", http.StatusConflict)
)
