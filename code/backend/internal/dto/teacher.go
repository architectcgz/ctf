package dto

import "time"

type TeacherClassItem struct {
	Name         string `json:"name"`
	StudentCount int64  `json:"student_count"`
}

type TeacherStudentItem struct {
	ID        int64   `json:"id"`
	Username  string  `json:"username"`
	StudentNo *string `json:"student_no,omitempty"`
	Name      *string `json:"name,omitempty"`
}

type TeacherStudentQuery struct {
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
