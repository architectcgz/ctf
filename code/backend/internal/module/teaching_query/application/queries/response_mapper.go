package queries

import (
	"time"

	"ctf-platform/internal/dto"
	queryports "ctf-platform/internal/module/teaching_query/ports"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:extend MapTeacherRecommendationItems
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :queries
type teachingQueryResponseMapper interface {
	ToClassItems(source []queryports.ClassItem) []dto.TeacherClassItem
	ToStudentItems(source []queryports.StudentItem) []dto.TeacherStudentItem
	ToClassSummary(source queryports.ClassSummary) dto.TeacherClassSummaryResp
	ToClassSummaryPtr(source *queryports.ClassSummary) *dto.TeacherClassSummaryResp
	ToClassTrendResp(source queryports.ClassTrend) dto.TeacherClassTrendResp
	ToClassTrendRespPtr(source *queryports.ClassTrend) *dto.TeacherClassTrendResp
	ToTimelineEvents(source []queryports.TimelineEventRecord) []dto.TimelineEvent
	ToReviewStudentRefs(source []dto.TeacherStudentItem) []dto.TeacherReviewStudentRef
	ToTeacherRecommendationWeakDimension(source dto.RecommendationWeakDimension) dto.TeacherRecommendationWeakDimension
	ToTeacherRecommendationWeakDimensions(source []dto.RecommendationWeakDimension) []dto.TeacherRecommendationWeakDimension
	ToTeacherRecommendationResp(source dto.RecommendationResp) dto.TeacherRecommendationResp
	ToTeacherRecommendationRespPtr(source *dto.RecommendationResp) *dto.TeacherRecommendationResp
	// goverter:map ID ChallengeID
	ToTeacherRecommendationItem(source dto.ChallengeRecommendation) dto.TeacherRecommendationItem
	ToTeacherRecommendationItemPtr(source *dto.ChallengeRecommendation) *dto.TeacherRecommendationItem
}

var teachingQueryMapper teachingQueryResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}

func MapTeacherRecommendationItems(source []*dto.ChallengeRecommendation) []dto.TeacherRecommendationItem {
	items := make([]dto.TeacherRecommendationItem, 0, len(source))
	for _, item := range source {
		if item == nil {
			continue
		}
		items = append(items, teachingQueryMapper.ToTeacherRecommendationItem(*item))
	}
	return items
}
