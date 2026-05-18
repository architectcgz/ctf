package domain

import (
	"time"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

type practiceResponseMapper interface {
	// goverter:ignore Access
	// goverter:ignore RemainingExtends
	ToInstanceRespBase(source instancecontracts.Instance) instancecontracts.InstanceResp
	ToInstanceRespBasePtr(source *instancecontracts.Instance) *instancecontracts.InstanceResp
}

var practiceResponseMapperInst practiceResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}
