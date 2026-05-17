package http

import (
	"time"

	practiceports "ctf-platform/internal/module/practice/ports"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :http
type practiceResponseMapperContract interface {
	ToProgressResp(source practiceports.UserProgressSnapshot) ProgressResp
	ToProgressRespPtr(source *practiceports.UserProgressSnapshot) *ProgressResp
	ToCategoryStat(source practiceports.UserProgressCategorySnapshot) CategoryStat
	ToDifficultyStat(source practiceports.UserProgressDifficultySnapshot) DifficultyStat
	ToTimelineResp(source practiceports.TimelineSnapshot) TimelineResp
	ToTimelineRespPtr(source *practiceports.TimelineSnapshot) *TimelineResp
	// goverter:map Timestamp | ConvertTime
	ToTimelineEvent(source practiceports.TimelineEventSnapshot) TimelineEvent
}

var practiceResponseMapper practiceResponseMapperContract

func ConvertTime(t time.Time) time.Time {
	return t
}
