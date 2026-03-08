package dto

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
