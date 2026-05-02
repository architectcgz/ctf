package queries

import (
	"time"

	"ctf-platform/internal/dto"
	readmodelports "ctf-platform/internal/module/teaching_readmodel/ports"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :queries
type teachingReadmodelResponseMapper interface {
	ToClassItems(source []readmodelports.ClassItem) []dto.TeacherClassItem
	ToStudentItems(source []readmodelports.StudentItem) []dto.TeacherStudentItem
	ToClassSummary(source readmodelports.ClassSummary) dto.TeacherClassSummaryResp
	ToClassTrendResp(source readmodelports.ClassTrend) dto.TeacherClassTrendResp
	ToTimelineEvents(source []readmodelports.TimelineEventRecord) []dto.TimelineEvent
	ToReviewStudentRefs(source []dto.TeacherStudentItem) []dto.TeacherReviewStudentRef
}

var teachingReadmodelMapper teachingReadmodelResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}
