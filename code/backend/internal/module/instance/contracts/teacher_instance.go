package contracts

import (
	"time"
)

type TeacherInstanceListQuery struct {
	ClassName string
	Keyword   string
	StudentNo string
	Status    string
	Page      int
	PageSize  int
}

type TeacherInstanceListSummary struct {
	TotalCount        int64
	RunningCount      int64
	ExpiringSoonCount int64
	WarningCount      int64
}

type TeacherInstancePageResult struct {
	List     []TeacherInstanceItem
	Total    int64
	Page     int
	PageSize int
	Summary  TeacherInstanceListSummary
}

type TeacherInstanceItem struct {
	ID              int64
	StudentID       int64
	StudentName     string
	StudentUsername string
	StudentNo       *string
	ClassName       string
	ChallengeID     int64
	ChallengeTitle  string
	Status          string
	AccessURL       string
	Access          *InstanceAccessInfo
	ExpiresAt       time.Time
	RemainingTime   int64
	ExtendCount     int
	MaxExtends      int
	CreatedAt       time.Time
}
