package apperror

import "net/http"

type Error struct {
	Code       int
	Message    string
	HTTPStatus int
}

func (e *Error) Error() string {
	return e.Message
}

func New(code int, message string, httpStatus int) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

var (
	ErrInvalidParams = New(10001, "请求参数错误", http.StatusBadRequest)
	ErrUnauthorized  = New(10003, "未认证，请先登录", http.StatusUnauthorized)
	ErrForbidden     = New(10004, "无权限访问该资源", http.StatusForbidden)
	ErrNotFound      = New(10005, "请求的资源不存在", http.StatusNotFound)
	ErrInternal      = New(10009, "服务器内部错误", http.StatusInternalServerError)
)
