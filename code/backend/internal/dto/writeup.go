package dto

import "time"

type UpsertChallengeWriteupReq struct {
	Title      string     `json:"title" binding:"required,max=256"`
	Content    string     `json:"content" binding:"required"`
	Visibility string     `json:"visibility" binding:"required,oneof=private public scheduled"`
	ReleaseAt  *time.Time `json:"release_at"`
}

type AdminChallengeWriteupResp struct {
	ID          int64      `json:"id"`
	ChallengeID int64      `json:"challenge_id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Visibility  string     `json:"visibility"`
	ReleaseAt   *time.Time `json:"release_at,omitempty"`
	CreatedBy   *int64     `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ChallengeWriteupResp struct {
	ID                     int64      `json:"id"`
	ChallengeID            int64      `json:"challenge_id"`
	Title                  string     `json:"title"`
	Content                string     `json:"content"`
	Visibility             string     `json:"visibility"`
	ReleaseAt              *time.Time `json:"release_at,omitempty"`
	IsReleased             bool       `json:"is_released"`
	RequiresSpoilerWarning bool       `json:"requires_spoiler_warning"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type UpsertSubmissionWriteupReq struct {
	Title            string `json:"title" binding:"required,max=256"`
	Content          string `json:"content" binding:"required"`
	SubmissionStatus string `json:"submission_status" binding:"required,oneof=draft submitted"`
}

type SubmissionWriteupResp struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	ChallengeID      int64      `json:"challenge_id"`
	ContestID        *int64     `json:"contest_id,omitempty"`
	Title            string     `json:"title"`
	Content          string     `json:"content"`
	SubmissionStatus string     `json:"submission_status"`
	ReviewStatus     string     `json:"review_status"`
	SubmittedAt      *time.Time `json:"submitted_at,omitempty"`
	ReviewedBy       *int64     `json:"reviewed_by,omitempty"`
	ReviewedAt       *time.Time `json:"reviewed_at,omitempty"`
	ReviewComment    string     `json:"review_comment,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type TeacherSubmissionWriteupQuery struct {
	StudentID        *int64 `form:"student_id" binding:"omitempty,min=1"`
	ChallengeID      *int64 `form:"challenge_id" binding:"omitempty,min=1"`
	ClassName        string `form:"class_name" binding:"omitempty,max=128"`
	SubmissionStatus string `form:"submission_status" binding:"omitempty,oneof=draft submitted"`
	ReviewStatus     string `form:"review_status" binding:"omitempty,oneof=pending reviewed excellent needs_revision"`
	Page             int    `form:"page" binding:"omitempty,min=1"`
	Size             int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type TeacherSubmissionWriteupItemResp struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	StudentUsername  string     `json:"student_username"`
	StudentName      string     `json:"student_name,omitempty"`
	ClassName        string     `json:"class_name,omitempty"`
	ChallengeID      int64      `json:"challenge_id"`
	ChallengeTitle   string     `json:"challenge_title"`
	Title            string     `json:"title"`
	ContentPreview   string     `json:"content_preview"`
	SubmissionStatus string     `json:"submission_status"`
	ReviewStatus     string     `json:"review_status"`
	SubmittedAt      *time.Time `json:"submitted_at,omitempty"`
	ReviewedAt       *time.Time `json:"reviewed_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type TeacherSubmissionWriteupDetailResp struct {
	SubmissionWriteupResp
	StudentUsername string `json:"student_username"`
	StudentName     string `json:"student_name,omitempty"`
	ClassName       string `json:"class_name,omitempty"`
	ChallengeTitle  string `json:"challenge_title"`
	ReviewerName    string `json:"reviewer_name,omitempty"`
}

type ReviewSubmissionWriteupReq struct {
	ReviewStatus  string `json:"review_status" binding:"required,oneof=pending reviewed excellent needs_revision"`
	ReviewComment string `json:"review_comment" binding:"omitempty,max=4000"`
}
