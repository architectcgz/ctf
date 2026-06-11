package queries

import (
	"time"

	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	teachinganalysiscontracts "ctf-platform/internal/module/teaching_analysis/contracts"
	queryports "ctf-platform/internal/module/teaching_analysis/ports"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:extend MapTeacherRecommendationItems
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :queries
type teachingAnalysisResponseMapper interface {
	ToClassItems(source []queryports.ClassItem) []TeacherClassItem
	ToStudentItems(source []queryports.StudentItem) []TeacherStudentItem
	ToTimelineEvents(source []queryports.TimelineEventRecord) []TimelineEvent
	ToTeacherRecommendationWeakDimension(source assessmentcontracts.RecommendationWeakDimension) TeacherRecommendationWeakDimension
	ToTeacherRecommendationWeakDimensions(source []assessmentcontracts.RecommendationWeakDimension) []TeacherRecommendationWeakDimension
	ToTeacherRecommendationResp(source assessmentcontracts.Recommendation) TeacherRecommendationResp
	ToTeacherRecommendationRespPtr(source *assessmentcontracts.Recommendation) *TeacherRecommendationResp
	// goverter:map ID ChallengeID
	ToTeacherRecommendationItem(source assessmentcontracts.ChallengeRecommendation) teachinganalysiscontracts.TeacherRecommendationItem
	ToTeacherRecommendationItemPtr(source *assessmentcontracts.ChallengeRecommendation) *teachinganalysiscontracts.TeacherRecommendationItem
}

var teachingAnalysisMapper teachingAnalysisResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}

func MapTeacherRecommendationItems(source []*assessmentcontracts.ChallengeRecommendation) []teachinganalysiscontracts.TeacherRecommendationItem {
	items := make([]teachinganalysiscontracts.TeacherRecommendationItem, 0, len(source))
	for _, item := range source {
		if item == nil {
			continue
		}
		items = append(items, teachingAnalysisMapper.ToTeacherRecommendationItem(*item))
	}
	return items
}
