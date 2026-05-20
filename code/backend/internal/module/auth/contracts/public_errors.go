package contracts

import (
	"net/http"

	"ctf-platform/internal/apperror"
)

var ErrWSTicketInvalid = apperror.Define(15002, "WebSocket Ticket 无效或已过期", http.StatusUnauthorized)
var (
	ErrInvalidCredentials    = apperror.Define(11001, "用户名或密码错误", http.StatusUnauthorized)
	ErrAccessTokenExpired    = apperror.Define(11002, "会话已失效或不存在", http.StatusUnauthorized)
	ErrRefreshTokenExpired   = apperror.Define(11003, "会话已过期", http.StatusUnauthorized)
	ErrTokenInvalid          = apperror.Define(11004, "Session Cookie 格式无效", http.StatusUnauthorized)
	ErrTokenRevoked          = apperror.Define(11005, "会话已被吊销", http.StatusUnauthorized)
	ErrAccountLocked         = apperror.Define(11006, "账户已被锁定", http.StatusForbidden)
	ErrAccountDisabled       = apperror.Define(11007, "账户已被禁用", http.StatusForbidden)
	ErrLoginTooFrequent      = apperror.Define(11010, "登录失败次数过多，账户临时锁定", http.StatusTooManyRequests)
	ErrCASDisabled           = apperror.Define(11015, "CAS 认证未启用", http.StatusServiceUnavailable)
	ErrCASNotConfigured      = apperror.Define(11016, "CAS 认证配置不完整", http.StatusServiceUnavailable)
	ErrCASNotImplemented     = apperror.Define(11017, "CAS 认证回调暂未实现", http.StatusNotImplemented)
	ErrCASTicketInvalid      = apperror.Define(11018, "CAS 票据无效或已过期", http.StatusUnauthorized)
	ErrCASUserNotProvisioned = apperror.Define(11019, "CAS 用户未在平台开通", http.StatusForbidden)
)
