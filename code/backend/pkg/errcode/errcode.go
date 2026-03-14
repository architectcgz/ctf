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
	ErrUnknown               = New(10000, "未知错误", http.StatusInternalServerError)
	ErrInvalidParams         = New(10001, "请求参数错误", http.StatusBadRequest)
	ErrValidationFailed      = New(10002, "请求参数校验失败", http.StatusBadRequest)
	ErrUnauthorized          = New(10003, "未认证，请先登录", http.StatusUnauthorized)
	ErrForbidden             = New(10004, "无权限访问该资源", http.StatusForbidden)
	ErrNotFound              = New(10005, "请求的资源不存在", http.StatusNotFound)
	ErrMethodNotAllowed      = New(10006, "请求方法不允许", http.StatusMethodNotAllowed)
	ErrConflict              = New(10007, "资源冲突", http.StatusConflict)
	ErrRateLimitExceeded     = New(10008, "请求频率超限，请稍后重试", http.StatusTooManyRequests)
	ErrInternal              = New(10009, "服务器内部错误", http.StatusInternalServerError)
	ErrServiceUnavailable    = New(10010, "服务暂时不可用", http.StatusServiceUnavailable)
	ErrInvalidCredentials    = New(11001, "用户名或密码错误", http.StatusUnauthorized)
	ErrAccessTokenExpired    = New(11002, "Access Token 已过期", http.StatusUnauthorized)
	ErrRefreshTokenExpired   = New(11003, "Refresh Token 已过期", http.StatusUnauthorized)
	ErrTokenInvalid          = New(11004, "Token 格式无效", http.StatusUnauthorized)
	ErrTokenRevoked          = New(11005, "Token 已被吊销", http.StatusUnauthorized)
	ErrAccountLocked         = New(11006, "账户已被锁定", http.StatusForbidden)
	ErrAccountDisabled       = New(11007, "账户已被禁用", http.StatusForbidden)
	ErrUsernameExists        = New(11008, "用户名已存在", http.StatusConflict)
	ErrEmailExists           = New(11009, "邮箱已被注册", http.StatusConflict)
	ErrLoginTooFrequent      = New(11010, "登录失败次数过多，账户临时锁定", http.StatusTooManyRequests)
	ErrOldPasswordInvalid    = New(11011, "原密码错误", http.StatusBadRequest)
	ErrPasswordUnchanged     = New(11012, "新密码不能与原密码相同", http.StatusBadRequest)
	ErrStudentNoExists       = New(11013, "学号已存在", http.StatusConflict)
	ErrTeacherNoExists       = New(11014, "工号已存在", http.StatusConflict)
	ErrCASDisabled           = New(11015, "CAS 认证未启用", http.StatusServiceUnavailable)
	ErrCASNotConfigured      = New(11016, "CAS 认证配置不完整", http.StatusServiceUnavailable)
	ErrCASNotImplemented     = New(11017, "CAS 认证回调暂未实现", http.StatusNotImplemented)
	ErrCASTicketInvalid      = New(11018, "CAS 票据无效或已过期", http.StatusUnauthorized)
	ErrCASUserNotProvisioned = New(11019, "CAS 用户未在平台开通", http.StatusForbidden)
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
	ErrContestNotFound             = New(14001, "竞赛不存在", http.StatusNotFound)
	ErrInvalidTimeRange            = New(14002, "结束时间必须晚于开始时间", http.StatusBadRequest)
	ErrContestAlreadyStarted       = New(14003, "竞赛已开始，无法修改", http.StatusForbidden)
	ErrInvalidStatusTransition     = New(14004, "非法的状态流转", http.StatusBadRequest)
	ErrCannotModifyAfterDraft      = New(14005, "非草稿状态下无法修改该字段", http.StatusForbidden)
	ErrContestImmutable            = New(14006, "竞赛已开始或已结束，无法修改题目配置", http.StatusForbidden)
	ErrChallengeAlreadyAdded       = New(14007, "题目已添加到竞赛", http.StatusConflict)
	ErrChallengeNotInContest       = New(14008, "题目不在竞赛中", http.StatusNotFound)
	ErrChallengeNotPublished       = New(14009, "只能添加已发布的题目", http.StatusBadRequest)
	ErrContestChallengeVisible     = New(14010, "当前竞赛状态下不可查看题目", http.StatusForbidden)
	ErrContestChallengeHasSubs     = New(14011, "该题目已有竞赛提交记录，无法移除", http.StatusConflict)
	ErrContestEnded                = New(14012, "竞赛已结束", http.StatusForbidden)
	ErrScoreboardNotFrozen         = New(14013, "排行榜未冻结", http.StatusBadRequest)
	ErrContestNotRunning           = New(14014, "竞赛未在进行中", http.StatusForbidden)
	ErrRegistrationNotApproved     = New(14015, "报名未通过审核", http.StatusForbidden)
	ErrNotRegistered               = New(14016, "未报名该竞赛", http.StatusForbidden)
	ErrContestChallengeSolved      = New(14017, "该题目已在本场竞赛中解出", http.StatusConflict)
	ErrContestRegistrationClosed   = New(14018, "当前竞赛状态不允许报名", http.StatusForbidden)
	ErrContestAnnouncementNotFound = New(14019, "竞赛公告不存在", http.StatusNotFound)
	ErrContestRegistrationPending  = New(14020, "报名待审核", http.StatusForbidden)
	ErrContestRegistrationNotFound = New(14021, "竞赛报名记录不存在", http.StatusNotFound)
	ErrAWDTeamRequired             = New(14022, "AWD 竞赛要求以队伍身份参赛", http.StatusForbidden)
	ErrAWDRoundNotActive           = New(14023, "当前没有可用的 AWD 轮次", http.StatusConflict)
	ErrAWDFlagUnavailable          = New(14024, "当前轮 Flag 暂不可用", http.StatusServiceUnavailable)
)

// 通知与 WebSocket 相关错误码 (15000-15999)
var (
	ErrNotificationNotFound = New(15001, "通知不存在", http.StatusNotFound)
	ErrWSTicketInvalid      = New(15002, "WebSocket Ticket 无效或已过期", http.StatusUnauthorized)
	ErrProxyTicketInvalid   = New(15003, "实例代理票据无效或已过期", http.StatusUnauthorized)
)
