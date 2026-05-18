package domain

import (
	"time"

	"ctf-platform/internal/model"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:extend CopyTime
// goverter:output:file ./response_mapper_goverter_gen.go
// goverter:output:package :domain
type practiceResponseMapper interface {
	// goverter:ignore Access
	// goverter:ignore RemainingExtends
	ToInstanceRespBase(source model.Instance) instancecontracts.InstanceResp
	ToInstanceRespBasePtr(source *model.Instance) *instancecontracts.InstanceResp
}

var practiceResponseMapperInst practiceResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}
