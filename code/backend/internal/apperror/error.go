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
