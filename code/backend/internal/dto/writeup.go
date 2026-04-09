package dto

import "time"

type UpsertChallengeWriteupReq struct {
	Title      string `json:"title" binding:"required,max=256"`
	Content    string `json:"content" binding:"required"`
	Visibility string `json:"visibility" binding:"required,oneof=private public"`
}

type AdminChallengeWriteupResp struct {
	ID            int64      `json:"id"`
	ChallengeID   int64      `json:"challenge_id"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	Visibility    string     `json:"visibility"`
	CreatedBy     *int64     `json:"created_by,omitempty"`
	IsRecommended bool       `json:"is_recommended"`
	RecommendedAt *time.Time `json:"recommended_at,omitempty"`
	RecommendedBy *int64     `json:"recommended_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type ChallengeWriteupResp struct {
	ID                     int64      `json:"id"`
	ChallengeID            int64      `json:"challenge_id"`
	Title                  string     `json:"title"`
	Content                string     `json:"content"`
	Visibility             string     `json:"visibility"`
	RequiresSpoilerWarning bool       `json:"requires_spoiler_warning"`
	IsRecommended          bool       `json:"is_recommended"`
	RecommendedAt          *time.Time `json:"recommended_at,omitempty"`
	RecommendedBy          *int64     `json:"recommended_by,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

type UpsertSubmissionWriteupReq struct {
	Title            string `json:"title" binding:"required,max=256"`
	Content          string `json:"content" binding:"required"`
	SubmissionStatus string `json:"submission_status" binding:"required,oneof=draft published"`
}

type SubmissionWriteupResp struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	ChallengeID      int64      `json:"challenge_id"`
	ContestID        *int64     `json:"contest_id,omitempty"`
	Title            string     `json:"title"`
	Content          string     `json:"content"`
	SubmissionStatus string     `json:"submission_status"`
	VisibilityStatus string     `json:"visibility_status"`
	IsRecommended    bool       `json:"is_recommended"`
	RecommendedAt    *time.Time `json:"recommended_at,omitempty"`
	RecommendedBy    *int64     `json:"recommended_by,omitempty"`
	PublishedAt      *time.Time `json:"published_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type TeacherSubmissionWriteupQuery struct {
	StudentID        *int64 `form:"student_id" binding:"omitempty,min=1"`
	ChallengeID      *int64 `form:"challenge_id" binding:"omitempty,min=1"`
	ClassName        string `form:"class_name" binding:"omitempty,max=128"`
	SubmissionStatus string `form:"submission_status" binding:"omitempty,oneof=draft published"`
	VisibilityStatus string `form:"visibility_status" binding:"omitempty,oneof=visible hidden"`
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
	VisibilityStatus string     `json:"visibility_status"`
	IsRecommended    bool       `json:"is_recommended"`
	PublishedAt      *time.Time `json:"published_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type TeacherSubmissionWriteupDetailResp struct {
	SubmissionWriteupResp
	StudentUsername string `json:"student_username"`
	StudentName     string `json:"student_name,omitempty"`
	ClassName       string `json:"class_name,omitempty"`
	ChallengeTitle  string `json:"challenge_title"`
}

type CommunityChallengeSolutionQuery struct {
	Q    string `form:"q" binding:"omitempty,max=128"`
	Sort string `form:"sort" binding:"omitempty,oneof=newest oldest"`
	Page int    `form:"page" binding:"omitempty,min=1"`
	Size int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type RecommendedChallengeSolutionResp struct {
	ID            string     `json:"id"`
	SourceType    string     `json:"source_type"`
	SourceID      int64      `json:"source_id"`
	ChallengeID   int64      `json:"challenge_id"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	AuthorName    string     `json:"author_name"`
	IsRecommended bool       `json:"is_recommended"`
	RecommendedAt *time.Time `json:"recommended_at,omitempty"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CommunityChallengeSolutionResp struct {
	ID               int64      `json:"id"`
	ChallengeID      int64      `json:"challenge_id"`
	UserID           int64      `json:"user_id"`
	Title            string     `json:"title"`
	Content          string     `json:"content"`
	ContentPreview   string     `json:"content_preview"`
	AuthorName       string     `json:"author_name"`
	SubmissionStatus string     `json:"submission_status"`
	VisibilityStatus string     `json:"visibility_status"`
	IsRecommended    bool       `json:"is_recommended"`
	PublishedAt      *time.Time `json:"published_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
