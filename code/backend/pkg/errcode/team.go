package errcode

import "net/http"

var (
	ErrAlreadyInTeam              = New(14101, "您已加入该竞赛的队伍", http.StatusConflict)
	ErrTeamFull                   = New(14102, "队伍人数已满", http.StatusForbidden)
	ErrTeamNotFound               = New(14103, "队伍不存在", http.StatusNotFound)
	ErrCaptainCannotLeave         = New(14104, "队长不能退出队伍，请先解散队伍", http.StatusForbidden)
	ErrNotCaptain                 = New(14105, "只有队长可以解散队伍", http.StatusForbidden)
	ErrNotInTeam                  = New(14106, "您不在该队伍中", http.StatusBadRequest)
	ErrInviteCodeGenerationFailed = New(14107, "创建队伍失败，请重试", http.StatusInternalServerError)
	ErrTeamNameExists             = New(14108, "同一竞赛下队伍名称已存在", http.StatusConflict)
	ErrContestTeamUnavailable     = New(14109, "当前竞赛状态不允许组队操作", http.StatusForbidden)
)
