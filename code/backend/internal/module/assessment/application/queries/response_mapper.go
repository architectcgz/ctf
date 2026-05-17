package queries

import (
	"time"

	assessmentdomain "ctf-platform/internal/module/assessment/domain"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :queries
type teacherAWDReviewResponseMapper interface {
	ToTeacherAWDReviewContestResp(source assessmentdomain.TeacherAWDReviewContestCard) TeacherAWDReviewContestResp
	ToTeacherAWDReviewContestResps(source []assessmentdomain.TeacherAWDReviewContestCard) []TeacherAWDReviewContestResp
	ToTeacherAWDReviewContestMetaResp(source assessmentdomain.TeacherAWDReviewContestMeta) TeacherAWDReviewContestMetaResp
	// goverter:ignore ServiceCount
	// goverter:ignore AttackCount
	// goverter:ignore TrafficCount
	ToTeacherAWDReviewRoundResp(source assessmentdomain.TeacherAWDReviewRoundSummary) TeacherAWDReviewRoundResp
	ToTeacherAWDReviewTeamResps(source []assessmentdomain.TeacherAWDReviewTeamSummary) []TeacherAWDReviewTeamResp
	ToTeacherAWDReviewServiceResps(source []assessmentdomain.TeacherAWDReviewServiceRecord) []TeacherAWDReviewServiceResp
	ToTeacherAWDReviewAttackResps(source []assessmentdomain.TeacherAWDReviewAttackRecord) []TeacherAWDReviewAttackResp
	ToTeacherAWDReviewTrafficResps(source []assessmentdomain.TeacherAWDReviewTrafficRecord) []TeacherAWDReviewTrafficResp
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
