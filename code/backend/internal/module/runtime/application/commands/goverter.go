package commands

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

//go:generate go run github.com/jmattheis/goverter/cmd/goverter gen .

// goverter:converter
// goverter:output:file ./zz_generated.goverter.go
// goverter:skipCopySameType
type RuntimeCommandResponseMapper interface {
	// goverter:ignore Access
	// goverter:ignore RemainingExtends
	ToInstanceResp(source *model.Instance) *dto.InstanceResp
}

var runtimeCommandResponseMapperInst RuntimeCommandResponseMapper
