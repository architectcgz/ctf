package queries

import (
	"time"

	"ctf-platform/internal/dto"
	assessmentdomain "ctf-platform/internal/module/assessment/domain"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :queries
type teacherAWDReviewResponseMapper interface {
	ToTeacherAWDReviewContestResp(source assessmentdomain.TeacherAWDReviewContestCard) dto.TeacherAWDReviewContestResp
	ToTeacherAWDReviewContestResps(source []assessmentdomain.TeacherAWDReviewContestCard) []dto.TeacherAWDReviewContestResp
	ToTeacherAWDReviewContestMetaResp(source assessmentdomain.TeacherAWDReviewContestMeta) dto.TeacherAWDReviewContestMetaResp
	// goverter:ignore ServiceCount
	// goverter:ignore AttackCount
	// goverter:ignore TrafficCount
	ToTeacherAWDReviewRoundResp(source assessmentdomain.TeacherAWDReviewRoundSummary) dto.TeacherAWDReviewRoundResp
	ToTeacherAWDReviewTeamResps(source []assessmentdomain.TeacherAWDReviewTeamSummary) []dto.TeacherAWDReviewTeamResp
	ToTeacherAWDReviewServiceResps(source []assessmentdomain.TeacherAWDReviewServiceRecord) []dto.TeacherAWDReviewServiceResp
	ToTeacherAWDReviewAttackResps(source []assessmentdomain.TeacherAWDReviewAttackRecord) []dto.TeacherAWDReviewAttackResp
	ToTeacherAWDReviewTrafficResps(source []assessmentdomain.TeacherAWDReviewTrafficRecord) []dto.TeacherAWDReviewTrafficResp
}

var teacherAWDReviewMapper teacherAWDReviewResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}

func CopyTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
