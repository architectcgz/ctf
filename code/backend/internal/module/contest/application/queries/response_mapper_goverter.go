package queries

import (
	"time"

	contestentity "ctf-platform/internal/module/contest/entity"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:extend CopyTime
// goverter:extend CopyTimePtr
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :queries
type contestQueryResponseMapper interface {
	ToContestResultBase(source contestentity.Contest) ContestResult
	ToContestResultBasePtr(source *contestentity.Contest) *ContestResult

	// goverter:ignore MemberCount
	ToTeamResultBase(source contestentity.Team) TeamResult
	ToTeamResultBasePtr(source *contestentity.Team) *TeamResult
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
