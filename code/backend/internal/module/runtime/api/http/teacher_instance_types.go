package http

import (
	"time"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

type TeacherInstanceQuery struct {
	ClassName string `form:"class_name" binding:"omitempty,max=128"`
	Keyword   string `form:"keyword" binding:"omitempty,max=128"`
	StudentNo string `form:"student_no" binding:"omitempty,max=64"`
}

type TeacherInstanceItem struct {
	ID              int64                   `json:"id"`
	StudentID       int64                   `json:"student_id"`
	StudentName     string                  `json:"student_name"`
	StudentUsername string                  `json:"student_username"`
	StudentNo       *string                 `json:"student_no,omitempty"`
	ClassName       string                  `json:"class_name"`
	ChallengeID     int64                   `json:"challenge_id"`
	ChallengeTitle  string                  `json:"challenge_title"`
	Status          string                  `json:"status"`
	AccessURL       string                  `json:"access_url"`
	Access          *instancecontracts.InstanceAccessInfo `json:"access,omitempty"`
	ExpiresAt       time.Time                             `json:"expires_at"`
	RemainingTime   int64                                 `json:"remaining_time"`
	ExtendCount     int                                   `json:"extend_count"`
	MaxExtends      int                                   `json:"max_extends"`
	CreatedAt       time.Time                             `json:"created_at"`
}
