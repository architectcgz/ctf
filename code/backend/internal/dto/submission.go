package dto

import "time"

const (
	SubmissionStatusCorrect       = "correct"
	SubmissionStatusIncorrect     = "incorrect"
	SubmissionStatusPendingReview = "pending_review"
)

// SubmitFlagReq Flag 提交请求
type SubmitFlagReq struct {
	Flag string `json:"flag" binding:"required"`
}

// SubmissionResp Flag 提交响应
type SubmissionResp struct {
	IsCorrect          bool       `json:"is_correct"`
	Status             string     `json:"status"`
	Message            string     `json:"message"`
	Points             int        `json:"points,omitempty"`
	SubmittedAt        time.Time  `json:"submitted_at"`
	InstanceShutdownAt *time.Time `json:"instance_shutdown_at,omitempty"`
}

type TeacherManualReviewSubmissionQuery struct {
	StudentID    *int64 `form:"student_id" binding:"omitempty,min=1"`
	ChallengeID  *int64 `form:"challenge_id" binding:"omitempty,min=1"`
	ClassName    string `form:"class_name" binding:"omitempty,max=128"`
	ReviewStatus string `form:"review_status" binding:"omitempty,oneof=pending approved rejected"`
	Page         int    `form:"page" binding:"omitempty,min=1"`
	Size         int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type TeacherManualReviewSubmissionItemResp struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	StudentUsername string     `json:"student_username"`
	StudentName     string     `json:"student_name,omitempty"`
	ClassName       string     `json:"class_name,omitempty"`
	ChallengeID     int64      `json:"challenge_id"`
	ChallengeTitle  string     `json:"challenge_title"`
	AnswerPreview   string     `json:"answer_preview"`
	ReviewStatus    string     `json:"review_status"`
	SubmittedAt     time.Time  `json:"submitted_at"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type TeacherManualReviewSubmissionDetailResp struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	StudentUsername string     `json:"student_username"`
	StudentName     string     `json:"student_name,omitempty"`
	ClassName       string     `json:"class_name,omitempty"`
	ChallengeID     int64      `json:"challenge_id"`
	ChallengeTitle  string     `json:"challenge_title"`
	Answer          string     `json:"answer"`
	IsCorrect       bool       `json:"is_correct"`
	Score           int        `json:"score"`
	ReviewStatus    string     `json:"review_status"`
	ReviewedBy      *int64     `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	ReviewComment   string     `json:"review_comment,omitempty"`
	SubmittedAt     time.Time  `json:"submitted_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ReviewerName    string     `json:"reviewer_name,omitempty"`
}

type ReviewManualReviewSubmissionReq struct {
	ReviewStatus  string `json:"review_status" binding:"required,oneof=approved rejected"`
	ReviewComment string `json:"review_comment" binding:"omitempty,max=4000"`
}
