package queries

import (
	"time"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
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
	ToInstanceInfo(source runtimeports.UserVisibleInstanceRow) instancecontracts.InstanceInfo
	ToInstanceInfoPtr(source *runtimeports.UserVisibleInstanceRow) *instancecontracts.InstanceInfo
}

var runtimeResponseMapper instanceResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}
