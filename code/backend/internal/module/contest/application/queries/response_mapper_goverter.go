package queries

import (
	"time"

	"ctf-platform/internal/model"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :queries
type contestQueryResponseMapper interface {
	ToContestResultBase(source model.Contest) ContestResult
	ToContestResultBasePtr(source *model.Contest) *ContestResult

	// goverter:ignore MemberCount
	ToTeamResultBase(source model.Team) TeamResult
	ToTeamResultBasePtr(source *model.Team) *TeamResult
}

var contestQueryResponseMapperInst contestQueryResponseMapper

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
