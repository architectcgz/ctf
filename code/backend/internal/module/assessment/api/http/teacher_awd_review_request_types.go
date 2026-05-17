package http

type GetTeacherAWDReviewArchiveReq struct {
	RoundNumber *int   `form:"round" binding:"omitempty,min=1"`
	TeamID      *int64 `form:"team_id"`
}

type TeacherAWDReviewContestQuery struct {
	Status  string `form:"status" binding:"omitempty,max=32"`
	Keyword string `form:"keyword" binding:"omitempty,max=128"`
	Page    int    `form:"page" binding:"omitempty,min=1"`
	Size    int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}
