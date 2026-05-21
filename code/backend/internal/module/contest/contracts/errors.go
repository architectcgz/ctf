package contracts

import (
	"net/http"

	"ctf-platform/internal/apperror"
)

var (
	ErrContestNotFound                 = apperror.Define(14001, "竞赛不存在", http.StatusNotFound)
	ErrInvalidTimeRange                = apperror.Define(14002, "结束时间必须晚于开始时间", http.StatusBadRequest)
	ErrContestAlreadyStarted           = apperror.Define(14003, "竞赛已开始，无法修改", http.StatusForbidden)
	ErrInvalidStatusTransition         = apperror.Define(14004, "非法的状态流转", http.StatusBadRequest)
	ErrCannotModifyAfterDraft          = apperror.Define(14005, "非草稿状态下无法修改该字段", http.StatusForbidden)
	ErrContestImmutable                = apperror.Define(14006, "竞赛已开始或已结束，无法修改题目配置", http.StatusForbidden)
	ErrChallengeAlreadyAdded           = apperror.Define(14007, "题目已添加到竞赛", http.StatusConflict)
	ErrChallengeNotInContest           = apperror.Define(14008, "题目不在竞赛中", http.StatusNotFound)
	ErrChallengeNotPublished           = apperror.Define(14009, "只能添加已发布的题目", http.StatusBadRequest)
	ErrContestChallengeVisible         = apperror.Define(14010, "当前竞赛状态下不可查看题目", http.StatusForbidden)
	ErrContestChallengeHasSubs         = apperror.Define(14011, "该题目已有竞赛提交记录，无法移除", http.StatusConflict)
	ErrContestEnded                    = apperror.Define(14012, "竞赛已结束", http.StatusForbidden)
	ErrScoreboardNotFrozen             = apperror.Define(14013, "排行榜未冻结", http.StatusBadRequest)
	ErrContestNotRunning               = apperror.Define(14014, "竞赛未在进行中", http.StatusForbidden)
	ErrRegistrationNotApproved         = apperror.Define(14015, "报名未通过审核", http.StatusForbidden)
	ErrNotRegistered                   = apperror.Define(14016, "未报名该竞赛", http.StatusForbidden)
	ErrContestChallengeSolved          = apperror.Define(14017, "该题目已在本场竞赛中解出", http.StatusConflict)
	ErrContestRegistrationClosed       = apperror.Define(14018, "当前竞赛状态不允许报名", http.StatusForbidden)
	ErrContestAnnouncementNotFound     = apperror.Define(14019, "竞赛公告不存在", http.StatusNotFound)
	ErrContestRegistrationPending      = apperror.Define(14020, "报名待审核", http.StatusForbidden)
	ErrContestRegistrationNotFound     = apperror.Define(14021, "竞赛报名记录不存在", http.StatusNotFound)
	ErrAWDTeamRequired                 = apperror.Define(14022, "AWD 竞赛要求以队伍身份参赛", http.StatusForbidden)
	ErrAWDRoundNotActive               = apperror.Define(14023, "当前没有可用的 AWD 轮次", http.StatusConflict)
	ErrAWDFlagUnavailable              = apperror.Define(14024, "当前轮 Flag 暂不可用", http.StatusServiceUnavailable)
	ErrAWDReadinessBlocked             = apperror.Define(14025, "AWD 开赛就绪检查未通过", http.StatusConflict)
	ErrAWDCheckerPreviewExpired        = apperror.Define(14026, "试跑结果已失效，请重新试跑后再保存", http.StatusConflict)
	ErrAWDCheckerPreviewUnavailable    = apperror.Define(14027, "当前未配置可保存试跑结果的 Redis，请先完成 Redis 配置后再试跑", http.StatusServiceUnavailable)
	ErrAWDDefenseSSHUnavailable        = apperror.Define(14028, "AWD SSH 防守入口未启用", http.StatusConflict)
	ErrAWDTeamRetired                  = apperror.Define(14029, "当前队伍已退赛", http.StatusForbidden)
	ErrAWDServiceDisabled              = apperror.Define(14030, "当前队伍服务已被停用", http.StatusForbidden)
	ErrContestEarlyEndRequiresOverride = apperror.Define(14031, "比赛尚未到结束时间，提前结束需要显式强制确认", http.StatusConflict)
)

var (
	ErrAlreadyInTeam              = apperror.Define(14101, "您已加入该竞赛的队伍", http.StatusConflict)
	ErrTeamFull                   = apperror.Define(14102, "队伍人数已满", http.StatusForbidden)
	ErrTeamNotFound               = apperror.Define(14103, "队伍不存在", http.StatusNotFound)
	ErrCaptainCannotLeave         = apperror.Define(14104, "队长不能退出队伍，请先解散队伍", http.StatusForbidden)
	ErrNotCaptain                 = apperror.Define(14105, "只有队长可以解散队伍", http.StatusForbidden)
	ErrNotInTeam                  = apperror.Define(14106, "您不在该队伍中", http.StatusBadRequest)
	ErrInviteCodeGenerationFailed = apperror.Define(14107, "创建队伍失败，请重试", http.StatusInternalServerError)
	ErrTeamNameExists             = apperror.Define(14108, "同一竞赛下队伍名称已存在", http.StatusConflict)
	ErrContestTeamUnavailable     = apperror.Define(14109, "当前竞赛状态不允许组队操作", http.StatusForbidden)
)
