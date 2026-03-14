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
	SolvedCount      int     `json:"solved_count"`
	TotalScore       int     `json:"total_score"`
	RecentEventCount int     `json:"recent_event_count"`
	WeakDimension    *string `json:"weak_dimension,omitempty"`
}

type TeacherStudentQuery struct {
	Keyword   string `form:"keyword" binding:"omitempty,max=128"`
	StudentNo string `form:"student_no" binding:"omitempty,max=64"`
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

type TeacherInstanceQuery struct {
	ClassName string `form:"class_name" binding:"omitempty,max=128"`
	Keyword   string `form:"keyword" binding:"omitempty,max=128"`
	StudentNo string `form:"student_no" binding:"omitempty,max=64"`
}

type TeacherInstanceItem struct {
	ID              int64     `json:"id"`
	StudentID       int64     `json:"student_id"`
	StudentName     string    `json:"student_name"`
	StudentUsername string    `json:"student_username"`
	StudentNo       *string   `json:"student_no,omitempty"`
	ClassName       string    `json:"class_name"`
	ChallengeID     int64     `json:"challenge_id"`
	ChallengeTitle  string    `json:"challenge_title"`
	Status          string    `json:"status"`
	AccessURL       string    `json:"access_url"`
	ExpiresAt       time.Time `json:"expires_at"`
	RemainingTime   int64     `json:"remaining_time"`
	ExtendCount     int       `json:"extend_count"`
	MaxExtends      int       `json:"max_extends"`
	CreatedAt       time.Time `json:"created_at"`
}
