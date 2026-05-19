package apperror

import "net/http"

type AppError struct {
	Code       int
	Message    string
	Cause      error
	httpStatus int
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func (e *AppError) WithCause(err error) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		Cause:      err,
		httpStatus: e.httpStatus,
	}
}

func (e *AppError) WithMessage(message string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    message,
		Cause:      e.Cause,
		httpStatus: e.httpStatus,
	}
}

func Define(code int, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		httpStatus: httpStatus,
	}
}

func HTTPStatus(err *AppError) int {
	if err == nil || err.httpStatus == 0 {
		return http.StatusInternalServerError
	}
	return err.httpStatus
}

var (
	ErrUnknown            = Define(10000, "未知错误", http.StatusInternalServerError)
	ErrInvalidParams      = Define(10001, "请求参数错误", http.StatusBadRequest)
	ErrValidationFailed   = Define(10002, "请求参数校验失败", http.StatusBadRequest)
	ErrUnauthorized       = Define(10003, "未认证，请先登录", http.StatusUnauthorized)
	ErrForbidden          = Define(10004, "无权限访问该资源", http.StatusForbidden)
	ErrNotFound           = Define(10005, "请求的资源不存在", http.StatusNotFound)
	ErrMethodNotAllowed   = Define(10006, "请求方法不允许", http.StatusMethodNotAllowed)
	ErrConflict           = Define(10007, "资源冲突", http.StatusConflict)
	ErrRateLimitExceeded  = Define(10008, "请求频率超限，请稍后重试", http.StatusTooManyRequests)
	ErrInternal           = Define(10009, "服务器内部错误", http.StatusInternalServerError)
	ErrServiceUnavailable = Define(10010, "服务暂时不可用", http.StatusServiceUnavailable)
)

var (
	ErrInvalidCredentials    = Define(11001, "用户名或密码错误", http.StatusUnauthorized)
	ErrAccessTokenExpired    = Define(11002, "会话已失效或不存在", http.StatusUnauthorized)
	ErrRefreshTokenExpired   = Define(11003, "会话已过期", http.StatusUnauthorized)
	ErrTokenInvalid          = Define(11004, "Session Cookie 格式无效", http.StatusUnauthorized)
	ErrTokenRevoked          = Define(11005, "会话已被吊销", http.StatusUnauthorized)
	ErrAccountLocked         = Define(11006, "账户已被锁定", http.StatusForbidden)
	ErrAccountDisabled       = Define(11007, "账户已被禁用", http.StatusForbidden)
	ErrUsernameExists        = Define(11008, "用户名已存在", http.StatusConflict)
	ErrEmailExists           = Define(11009, "邮箱已被注册", http.StatusConflict)
	ErrLoginTooFrequent      = Define(11010, "登录失败次数过多，账户临时锁定", http.StatusTooManyRequests)
	ErrOldPasswordInvalid    = Define(11011, "原密码错误", http.StatusBadRequest)
	ErrPasswordUnchanged     = Define(11012, "新密码不能与原密码相同", http.StatusBadRequest)
	ErrStudentNoExists       = Define(11013, "学号已存在", http.StatusConflict)
	ErrTeacherNoExists       = Define(11014, "工号已存在", http.StatusConflict)
	ErrCASDisabled           = Define(11015, "CAS 认证未启用", http.StatusServiceUnavailable)
	ErrCASNotConfigured      = Define(11016, "CAS 认证配置不完整", http.StatusServiceUnavailable)
	ErrCASNotImplemented     = Define(11017, "CAS 认证回调暂未实现", http.StatusNotImplemented)
	ErrCASTicketInvalid      = Define(11018, "CAS 票据无效或已过期", http.StatusUnauthorized)
	ErrCASUserNotProvisioned = Define(11019, "CAS 用户未在平台开通", http.StatusForbidden)
)
