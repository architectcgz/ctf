package http

import (
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	"time"
)

type TeacherInstanceQuery struct {
	ClassName string `form:"class_name" binding:"omitempty,max=128"`
	Keyword   string `form:"keyword" binding:"omitempty,max=128"`
	StudentNo string `form:"student_no" binding:"omitempty,max=64"`
	Status    string `form:"status" binding:"omitempty,oneof=running creating expired failed inactive"`
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type TeacherInstanceListSummaryResp struct {
	TotalCount        int64 `json:"total_count"`
	RunningCount      int64 `json:"running_count"`
	ExpiringSoonCount int64 `json:"expiring_soon_count"`
	WarningCount      int64 `json:"warning_count"`
}

type TeacherInstancePageResp struct {
	List     []TeacherInstanceItem          `json:"list"`
	Total    int64                          `json:"total"`
	Page     int                            `json:"page"`
	PageSize int                            `json:"page_size"`
	Summary  TeacherInstanceListSummaryResp `json:"summary"`
}

type TeacherInstanceItem struct {
	ID              int64                                 `json:"id"`
	StudentID       int64                                 `json:"student_id"`
	StudentName     string                                `json:"student_name"`
	StudentUsername string                                `json:"student_username"`
	StudentNo       *string                               `json:"student_no,omitempty"`
	ClassName       string                                `json:"class_name"`
	ChallengeID     int64                                 `json:"challenge_id"`
	ChallengeTitle  string                                `json:"challenge_title"`
	Status          string                                `json:"status"`
	AccessURL       string                                `json:"access_url"`
	Access          *instancecontracts.InstanceAccessInfo `json:"access,omitempty"`
	ExpiresAt       time.Time                             `json:"expires_at"`
	RemainingTime   int64                                 `json:"remaining_time"`
	ExtendCount     int                                   `json:"extend_count"`
	MaxExtends      int                                   `json:"max_extends"`
	CreatedAt       time.Time                             `json:"created_at"`
}
