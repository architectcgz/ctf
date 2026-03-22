package dto

import "time"

type CreateContestReq struct {
	Title       string    `json:"title" binding:"required,min=1,max=200"`
	Description string    `json:"description" binding:"max=5000"`
	Mode        string    `json:"mode" binding:"required,oneof=jeopardy awd"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required,gtfield=StartTime"`
}

type UpdateContestReq struct {
	Title       *string    `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string    `json:"description" binding:"omitempty,max=5000"`
	Mode        *string    `json:"mode" binding:"omitempty,oneof=jeopardy awd"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
	Status      *string    `json:"status" binding:"omitempty,oneof=draft registration running frozen ended"`
}

type ContestResp struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Mode        string     `json:"mode"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time"`
	FreezeTime  *time.Time `json:"freeze_time,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ListContestsReq struct {
	Status *string `form:"status" binding:"omitempty,oneof=draft registration running frozen ended"`
	Page   int     `form:"page" binding:"omitempty,min=1"`
	Size   int     `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type ScoreboardContestInfo struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
	EndsAt    time.Time `json:"ends_at"`
}

type ScoreboardItem struct {
	Rank             int        `json:"rank"`
	TeamID           int64      `json:"team_id"`
	TeamName         string     `json:"team_name"`
	Score            float64    `json:"score"`
	SolvedCount      int        `json:"solved_count"`
	LastSubmissionAt *time.Time `json:"last_submission_at,omitempty"`
}

type ScoreboardPage struct {
	List     []*ScoreboardItem `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

type ScoreboardResp struct {
	Contest    *ScoreboardContestInfo `json:"contest"`
	Scoreboard *ScoreboardPage        `json:"scoreboard"`
	Frozen     bool                   `json:"frozen"`
}

type TeamRankResp struct {
	TeamID int64   `json:"team_id"`
	Rank   int     `json:"rank"`
	Score  float64 `json:"score"`
}

type FreezeReq struct {
	MinutesBeforeEnd int `json:"minutes_before_end" binding:"required,min=1"`
}

type CreateContestAnnouncementReq struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"max=5000"`
}

type ContestRegistrationQuery struct {
	Status *string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
	Page   int     `form:"page" binding:"omitempty,min=1"`
	Size   int     `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type ReviewContestRegistrationReq struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
}

type ContestRegistrationResp struct {
	ID         int64      `json:"id"`
	ContestID  int64      `json:"contest_id"`
	UserID     int64      `json:"user_id"`
	Username   string     `json:"username"`
	TeamID     *int64     `json:"team_id,omitempty"`
	Status     string     `json:"status"`
	ReviewedBy *int64     `json:"reviewed_by,omitempty"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type ContestAnnouncementResp struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ContestSolvedProgressItem struct {
	ContestChallengeID int64     `json:"contest_challenge_id"`
	SolvedAt           time.Time `json:"solved_at"`
	PointsEarned       int       `json:"points_earned"`
}

type ContestMyProgressResp struct {
	ContestID int64                        `json:"contest_id"`
	TeamID    *int64                       `json:"team_id,omitempty"`
	Solved    []*ContestSolvedProgressItem `json:"solved"`
}
