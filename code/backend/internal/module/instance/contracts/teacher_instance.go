package contracts

import (
	"time"

	"ctf-platform/internal/dto"
)

type TeacherInstanceListQuery struct {
	ClassName string
	Keyword   string
	StudentNo string
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
	Access          *dto.InstanceAccessInfo
	ExpiresAt       time.Time
	RemainingTime   int64
	ExtendCount     int
	MaxExtends      int
	CreatedAt       time.Time
}
