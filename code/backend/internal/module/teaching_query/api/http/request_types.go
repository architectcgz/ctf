package http

import "time"

type TeacherClassQuery struct {
	Page int `form:"page" binding:"omitempty,min=1"`
	Size int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type TeacherClassInsightQuery struct {
	FromDate string `form:"from_date"`
	ToDate   string `form:"to_date"`
}

type TeacherStudentQuery struct {
	Keyword   string `form:"keyword" binding:"omitempty,max=128"`
	StudentNo string `form:"student_no" binding:"omitempty,max=64"`
}

type TeacherStudentDirectoryQuery struct {
	ClassName string `form:"class_name" binding:"omitempty,max=128"`
	Keyword   string `form:"keyword" binding:"omitempty,max=128"`
	StudentNo string `form:"student_no" binding:"omitempty,max=64"`
	Page      int    `form:"page" binding:"omitempty,min=1"`
	Size      int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortKey   string `form:"sort_key" binding:"omitempty,oneof=name student_no total_score solved_count"`
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}

type TeacherEvidenceQuery struct {
	ChallengeID *int64     `form:"challenge_id" binding:"omitempty,min=1"`
	ContestID   *int64     `form:"contest_id" binding:"omitempty,min=1"`
	RoundID     *int64     `form:"round_id" binding:"omitempty,min=1"`
	EventType   string     `form:"event_type" binding:"omitempty,max=64"`
	From        *time.Time `form:"from" time_format:"2006-01-02T15:04:05Z07:00"`
	To          *time.Time `form:"to" time_format:"2006-01-02T15:04:05Z07:00"`
	Limit       int        `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset      int        `form:"offset" binding:"omitempty,min=0"`
}

type TeacherAttackSessionQuery struct {
	Mode        string `form:"mode" binding:"omitempty,oneof=practice jeopardy awd"`
	ChallengeID *int64 `form:"challenge_id" binding:"omitempty,min=1"`
	ContestID   *int64 `form:"contest_id" binding:"omitempty,min=1"`
	RoundID     *int64 `form:"round_id" binding:"omitempty,min=1"`
	Result      string `form:"result" binding:"omitempty,oneof=success failed in_progress unknown"`
	WithEvents  *bool  `form:"with_events"`
	Limit       int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset      int    `form:"offset" binding:"omitempty,min=0"`
}
