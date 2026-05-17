package contracts

import "ctf-platform/internal/dto"

type TeacherClassSummary struct {
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

type TeacherClassTrend struct {
	ClassName string                   `json:"class_name"`
	Points    []TeacherClassTrendPoint `json:"points"`
}

type TeacherReviewStudentRef struct {
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Name     *string `json:"name,omitempty"`
}

type TeacherClassReviewItem struct {
	Code           string                         `json:"code"`
	Severity       string                         `json:"severity"`
	Summary        string                         `json:"summary"`
	Evidence       string                         `json:"evidence,omitempty"`
	Action         string                         `json:"action,omitempty"`
	ReasonCodes    []string                       `json:"reason_codes,omitempty"`
	Dimension      string                         `json:"dimension,omitempty"`
	Students       []TeacherReviewStudentRef      `json:"students,omitempty"`
	Recommendation *dto.TeacherRecommendationItem `json:"recommendation,omitempty"`
}

type TeacherClassReview struct {
	ClassName string                   `json:"class_name"`
	Items     []TeacherClassReviewItem `json:"items"`
}
