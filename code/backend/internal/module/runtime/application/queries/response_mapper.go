package queries

import (
	"time"

	"ctf-platform/internal/dto"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:extend CopyTime
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :queries
type instanceResponseMapper interface {
	// goverter:ignore Status
	// goverter:ignore AccessURL
	// goverter:ignore Access
	// goverter:ignore RemainingTime
	// goverter:ignore RemainingExtends
	ToInstanceInfo(source runtimeports.UserVisibleInstanceRow) dto.InstanceInfo
	ToInstanceInfoPtr(source *runtimeports.UserVisibleInstanceRow) *dto.InstanceInfo

	// goverter:ignore Status
	// goverter:ignore Access
	// goverter:ignore RemainingTime
	ToTeacherInstanceItem(source runtimeports.TeacherInstanceRow) dto.TeacherInstanceItem
	ToTeacherInstanceItemPtr(source *runtimeports.TeacherInstanceRow) *dto.TeacherInstanceItem
}

var runtimeResponseMapper instanceResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}
