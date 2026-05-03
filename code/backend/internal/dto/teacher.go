package dto

import "time"

type TeacherClassItem struct {
	Name         string `json:"name"`
	StudentCount int64  `json:"student_count"`
}

type TeacherClassQuery struct {
	Page int `form:"page" binding:"omitempty,min=1"`
	Size int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type TeacherClassSummaryResp struct {
	ClassName          string  `json:"class_name"`
	StudentCount       int64   `json:"student_count"`
	AverageSolved      float64 `json:"average_solved"`
	ActiveStudentCount int64   `json:"active_student_count"`
	ActiveRate         float64 `json:"active_rate"`
	RecentEventCount   int64   `json:"recent_event_count"`
}

type TeacherClassTrendPoint struct {
	Date               string `json:"date"`
	ActiveStudentCount int64  `json:"active_student_count"`
	EventCount         int64  `json:"event_count"`
	SolveCount         int64  `json:"solve_count"`
}

type TeacherClassTrendResp struct {
	ClassName string                   `json:"class_name"`
	Points    []TeacherClassTrendPoint `json:"points"`
}

type TeacherReviewStudentRef struct {
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Name     *string `json:"name,omitempty"`
}

type TeacherClassReviewItem struct {
	Key            string                     `json:"key"`
	Title          string                     `json:"title"`
	Detail         string                     `json:"detail"`
	Accent         string                     `json:"accent"`
	Students       []TeacherReviewStudentRef  `json:"students,omitempty"`
	Recommendation *TeacherRecommendationItem `json:"recommendation,omitempty"`
}

type TeacherClassReviewResp struct {
	ClassName string                   `json:"class_name"`
	Items     []TeacherClassReviewItem `json:"items"`
}

type TeacherStudentItem struct {
	ID               int64   `json:"id"`
	Username         string  `json:"username"`
	StudentNo        *string `json:"student_no,omitempty"`
	Name             *string `json:"name,omitempty"`
	ClassName        *string `json:"class_name,omitempty"`
	SolvedCount      int     `json:"solved_count"`
	TotalScore       int     `json:"total_score"`
	RecentEventCount int     `json:"recent_event_count"`
	WeakDimension    *string `json:"weak_dimension,omitempty"`
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

type ProgressBreakdown struct {
	Total  int `json:"total"`
	Solved int `json:"solved"`
}

type TeacherProgressResp struct {
	TotalChallenges  int                          `json:"total_challenges"`
	SolvedChallenges int                          `json:"solved_challenges"`
	ByCategory       map[string]ProgressBreakdown `json:"by_category,omitempty"`
	ByDifficulty     map[string]ProgressBreakdown `json:"by_difficulty,omitempty"`
}

type TeacherRecommendationItem struct {
	ChallengeID int64  `json:"challenge_id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Difficulty  string `json:"difficulty"`
	Reason      string `json:"reason"`
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

type TeacherEvidenceSummary struct {
	TotalEvents       int   `json:"total_events"`
	ProxyRequestCount int   `json:"proxy_request_count"`
	SubmitCount       int   `json:"submit_count"`
	SuccessCount      int   `json:"success_count"`
	ChallengeID       int64 `json:"challenge_id"`
}

type TeacherEvidenceEvent struct {
	Type        string         `json:"type"`
	ChallengeID int64          `json:"challenge_id"`
	Title       string         `json:"title"`
	Detail      string         `json:"detail"`
	Timestamp   time.Time      `json:"timestamp"`
	Meta        map[string]any `json:"meta,omitempty"`
}

type TeacherEvidenceResp struct {
	Summary TeacherEvidenceSummary `json:"summary"`
	Events  []TeacherEvidenceEvent `json:"events"`
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

type TeacherAttackActor struct {
	UserID int64  `json:"user_id"`
	TeamID *int64 `json:"team_id,omitempty"`
}

type TeacherAttackTarget struct {
	ChallengeID  *int64 `json:"challenge_id,omitempty"`
	ContestID    *int64 `json:"contest_id,omitempty"`
	RoundID      *int64 `json:"round_id,omitempty"`
	ServiceID    *int64 `json:"service_id,omitempty"`
	VictimTeamID *int64 `json:"victim_team_id,omitempty"`
}

type TeacherAttackEvent struct {
	ID               string                 `json:"id"`
	SessionID        string                 `json:"session_id,omitempty"`
	Type             string                 `json:"type"`
	Stage            string                 `json:"stage"`
	Source           string                 `json:"source"`
	OccurredAt       time.Time              `json:"occurred_at"`
	Actor            TeacherAttackActor     `json:"actor"`
	Target           TeacherAttackTarget    `json:"target"`
	Summary          string                 `json:"summary"`
	Meta             map[string]any         `json:"meta,omitempty"`
	CaptureAvailable bool                   `json:"capture_available"`
	CaptureRef       map[string]interface{} `json:"capture_ref,omitempty"`
}

type TeacherAttackSession struct {
	ID           string               `json:"id"`
	Mode         string               `json:"mode"`
	StudentID    int64                `json:"student_id"`
	TeamID       *int64               `json:"team_id,omitempty"`
	ChallengeID  *int64               `json:"challenge_id,omitempty"`
	ContestID    *int64               `json:"contest_id,omitempty"`
	RoundID      *int64               `json:"round_id,omitempty"`
	ServiceID    *int64               `json:"service_id,omitempty"`
	VictimTeamID *int64               `json:"victim_team_id,omitempty"`
	Title        string               `json:"title"`
	StartedAt    time.Time            `json:"started_at"`
	EndedAt      time.Time            `json:"ended_at"`
	Result       string               `json:"result"`
	EventCount   int                  `json:"event_count"`
	CaptureCount int                  `json:"capture_count"`
	Events       []TeacherAttackEvent `json:"events,omitempty"`
}

type TeacherAttackSessionSummary struct {
	TotalSessions         int `json:"total_sessions"`
	SuccessCount          int `json:"success_count"`
	FailedCount           int `json:"failed_count"`
	InProgressCount       int `json:"in_progress_count"`
	UnknownCount          int `json:"unknown_count"`
	EventCount            int `json:"event_count"`
	CaptureAvailableCount int `json:"capture_available_count"`
}

type TeacherAttackSessionResp struct {
	Summary  TeacherAttackSessionSummary `json:"summary"`
	Sessions []TeacherAttackSession      `json:"sessions"`
}

type TeacherInstanceQuery struct {
	ClassName string `form:"class_name" binding:"omitempty,max=128"`
	Keyword   string `form:"keyword" binding:"omitempty,max=128"`
	StudentNo string `form:"student_no" binding:"omitempty,max=64"`
}

type TeacherInstanceItem struct {
	ID              int64               `json:"id"`
	StudentID       int64               `json:"student_id"`
	StudentName     string              `json:"student_name"`
	StudentUsername string              `json:"student_username"`
	StudentNo       *string             `json:"student_no,omitempty"`
	ClassName       string              `json:"class_name"`
	ChallengeID     int64               `json:"challenge_id"`
	ChallengeTitle  string              `json:"challenge_title"`
	Status          string              `json:"status"`
	AccessURL       string              `json:"access_url"`
	Access          *InstanceAccessInfo `json:"access,omitempty"`
	ExpiresAt       time.Time           `json:"expires_at"`
	RemainingTime   int64               `json:"remaining_time"`
	ExtendCount     int                 `json:"extend_count"`
	MaxExtends      int                 `json:"max_extends"`
	CreatedAt       time.Time           `json:"created_at"`
}
