package errcode

var (
	ErrAlreadyInTeam             = New(14001, "您已加入该竞赛的队伍", 409)
	ErrInvalidInviteCode         = New(14002, "邀请码无效", 404)
	ErrAlreadyInTeamDup          = New(14003, "您已加入该竞赛的队伍", 409)
	ErrTeamFull                  = New(14004, "队伍人数已满", 403)
	ErrTeamNotFound              = New(14005, "队伍不存在", 404)
	ErrCaptainCannotLeave        = New(14006, "队长不能退出队伍，请先解散队伍", 403)
	ErrNotCaptain                = New(14007, "只有队长可以解散队伍", 403)
	ErrNotInTeam                 = New(14008, "您不在该队伍中", 400)
	ErrInviteCodeGenerationFailed = New(14009, "创建队伍失败，请重试", 500)
)
