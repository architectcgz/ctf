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
