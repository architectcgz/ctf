package contracts

import (
	"net/http"

	"ctf-platform/internal/apperror"
)

var (
	ErrInstanceNotFound      = apperror.Define(12001, "实例不存在", http.StatusNotFound)
	ErrInstanceLimitExceeded = apperror.Define(12002, "实例数量超限", http.StatusForbidden)
	ErrInstanceExpired       = apperror.Define(12003, "实例已过期", http.StatusGone)
	ErrExtendLimitExceeded   = apperror.Define(12004, "延时次数已达上限", http.StatusForbidden)
	ErrContainerCreateFailed = apperror.Define(12005, "容器创建失败", http.StatusInternalServerError)
	ErrContainerStartFailed  = apperror.Define(12006, "容器启动失败", http.StatusInternalServerError)
	ErrProxyTicketInvalid    = apperror.Define(15003, "实例代理票据无效或已过期", http.StatusUnauthorized)
)
