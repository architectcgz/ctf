package dto

import "time"

type TeacherClassItem struct {
	Name         string `json:"name"`
	StudentCount int64  `json:"student_count"`
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
	Code           string                     `json:"code"`
	Severity       string                     `json:"severity"`
	Summary        string                     `json:"summary"`
	Evidence       string                     `json:"evidence,omitempty"`
	Action         string                     `json:"action,omitempty"`
	ReasonCodes    []string                   `json:"reason_codes,omitempty"`
	Dimension      string                     `json:"dimension,omitempty"`
	Students       []TeacherReviewStudentRef  `json:"students,omitempty"`
	Recommendation *TeacherRecommendationItem `json:"recommendation,omitempty"`
}

type TeacherClassReviewResp struct {
	ClassName string                   `json:"class_name"`
	Items     []TeacherClassReviewItem `json:"items"`
}

type TeacherOverviewSummaryResp struct {
	ClassCount         int64   `json:"class_count"`
	StudentCount       int64   `json:"student_count"`
	ActiveStudentCount int64   `json:"active_student_count"`
	ActiveRate         float64 `json:"active_rate"`
	AverageSolved      float64 `json:"average_solved"`
	RecentEventCount   int64   `json:"recent_event_count"`
	RiskStudentCount   int64   `json:"risk_student_count"`
}

type TeacherOverviewTrendPoint struct {
	Date               string `json:"date"`
	ActiveStudentCount int64  `json:"active_student_count"`
	EventCount         int64  `json:"event_count"`
	SolveCount         int64  `json:"solve_count"`
}

type TeacherOverviewTrendResp struct {
	Points []TeacherOverviewTrendPoint `json:"points"`
}

type TeacherOverviewWeakDimensionResp struct {
	Dimension    string `json:"dimension"`
	StudentCount int64  `json:"student_count"`
}

type TeacherOverviewClassFocusResp struct {
	ClassName             string  `json:"class_name"`
	StudentCount          int64   `json:"student_count"`
	ActiveRate            float64 `json:"active_rate"`
	RecentEventCount      int64   `json:"recent_event_count"`
	RiskStudentCount      int64   `json:"risk_student_count"`
	DominantWeakDimension string  `json:"dominant_weak_dimension,omitempty"`
}

type TeacherOverviewResp struct {
	Summary          TeacherOverviewSummaryResp         `json:"summary"`
	Trend            TeacherOverviewTrendResp           `json:"trend"`
	FocusClasses     []TeacherOverviewClassFocusResp    `json:"focus_classes"`
	FocusStudents    []TeacherStudentItem               `json:"focus_students"`
	SpotlightStudent *TeacherStudentItem                `json:"spotlight_student,omitempty"`
	WeakDimensions   []TeacherOverviewWeakDimensionResp `json:"weak_dimensions"`
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

type TeacherRecommendationWeakDimension struct {
	Dimension  string  `json:"dimension"`
	Severity   string  `json:"severity"`
	Confidence float64 `json:"confidence"`
	Evidence   string  `json:"evidence,omitempty"`
}

type TeacherRecommendationItem struct {
	ChallengeID    int64    `json:"challenge_id"`
	Title          string   `json:"title"`
	Category       string   `json:"category"`
	Difficulty     string   `json:"difficulty"`
	Dimension      string   `json:"dimension,omitempty"`
	DifficultyBand string   `json:"difficulty_band,omitempty"`
	Severity       string   `json:"severity,omitempty"`
	ReasonCodes    []string `json:"reason_codes,omitempty"`
	Summary        string   `json:"summary"`
	Evidence       string   `json:"evidence,omitempty"`
}

type TeacherRecommendationResp struct {
	WeakDimensions []TeacherRecommendationWeakDimension `json:"weak_dimensions"`
	Challenges     []TeacherRecommendationItem          `json:"challenges"`
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
