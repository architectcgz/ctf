package commands

import (
	"time"

	"ctf-platform/internal/model"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen .

// goverter:converter
// goverter:enum:unknown @ignore
// goverter:extend CopyTime
// goverter:output:file ./response_mapper_gen.go
// goverter:output:package :commands
type instanceResponseMapper interface {
	// goverter:ignore Access
	// goverter:ignore RemainingExtends
	ToInstanceResp(source model.Instance) instancecontracts.InstanceResp
	ToInstanceRespPtr(source *model.Instance) *instancecontracts.InstanceResp
}

var runtimeResponseMapper instanceResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}
