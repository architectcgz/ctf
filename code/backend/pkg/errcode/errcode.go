package errcode

import "net/http"

type AppError struct {
	Code       int
	Message    string
	HTTPStatus int
	Cause      error
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
		HTTPStatus: e.HTTPStatus,
		Cause:      err,
	}
}

func New(code int, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

var (
	ErrUnknown             = New(10000, "未知错误", http.StatusInternalServerError)
	ErrInvalidParams       = New(10001, "请求参数错误", http.StatusBadRequest)
	ErrValidationFailed    = New(10002, "请求参数校验失败", http.StatusBadRequest)
	ErrUnauthorized        = New(10003, "未认证，请先登录", http.StatusUnauthorized)
	ErrForbidden           = New(10004, "无权限访问该资源", http.StatusForbidden)
	ErrNotFound            = New(10005, "请求的资源不存在", http.StatusNotFound)
	ErrMethodNotAllowed    = New(10006, "请求方法不允许", http.StatusMethodNotAllowed)
	ErrConflict            = New(10007, "资源冲突", http.StatusConflict)
	ErrRateLimitExceeded   = New(10008, "请求频率超限，请稍后重试", http.StatusTooManyRequests)
	ErrInternal            = New(10009, "服务器内部错误", http.StatusInternalServerError)
	ErrServiceUnavailable  = New(10010, "服务暂时不可用", http.StatusServiceUnavailable)
	ErrInvalidCredentials  = New(11001, "用户名或密码错误", http.StatusUnauthorized)
	ErrAccessTokenExpired  = New(11002, "Access Token 已过期", http.StatusUnauthorized)
	ErrRefreshTokenExpired = New(11003, "Refresh Token 已过期", http.StatusUnauthorized)
	ErrTokenInvalid        = New(11004, "Token 格式无效", http.StatusUnauthorized)
	ErrTokenRevoked        = New(11005, "Token 已被吊销", http.StatusUnauthorized)
	ErrAccountLocked       = New(11006, "账户已被锁定", http.StatusForbidden)
	ErrAccountDisabled     = New(11007, "账户已被禁用", http.StatusForbidden)
	ErrUsernameExists      = New(11008, "用户名已存在", http.StatusConflict)
	ErrEmailExists         = New(11009, "邮箱已被注册", http.StatusConflict)
	ErrLoginTooFrequent    = New(11010, "登录失败次数过多，账户临时锁定", http.StatusTooManyRequests)
)

// 容器相关错误码 (12000-12999)
var (
	ErrInstanceNotFound      = New(12001, "实例不存在", http.StatusNotFound)
	ErrInstanceLimitExceeded = New(12002, "实例数量超限", http.StatusForbidden)
	ErrInstanceExpired       = New(12003, "实例已过期", http.StatusGone)
	ErrExtendLimitExceeded   = New(12004, "延时次数已达上限", http.StatusForbidden)
	ErrContainerCreateFailed = New(12005, "容器创建失败", http.StatusInternalServerError)
	ErrContainerStartFailed  = New(12006, "容器启动失败", http.StatusInternalServerError)
)

// Flag 提交相关错误码 (13000-13999)
var (
	ErrFlagIncorrect       = New(13001, "Flag 错误", http.StatusBadRequest)
	ErrAlreadySolved       = New(13002, "该题目已完成", http.StatusConflict)
	ErrSubmitTooFrequent   = New(13003, "提交过于频繁，请稍后再试", http.StatusTooManyRequests)
	ErrChallengeNotFound   = New(13004, "靶场不存在", http.StatusNotFound)
	ErrChallengeNotPublish = New(13005, "靶场未发布", http.StatusForbidden)
)

// 竞赛相关错误码 (14000-14999)
var (
	ErrContestNotFound           = New(14001, "竞赛不存在", http.StatusNotFound)
	ErrInvalidTimeRange          = New(14002, "结束时间必须晚于开始时间", http.StatusBadRequest)
	ErrContestAlreadyStarted     = New(14003, "竞赛已开始，无法修改", http.StatusForbidden)
	ErrInvalidStatusTransition   = New(14004, "非法的状态流转", http.StatusBadRequest)
)
