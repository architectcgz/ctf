package commands

import (
	"time"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

type instanceResponseMapper interface {
	// goverter:ignore Access
	// goverter:ignore RemainingExtends
	ToInstanceResp(source instancecontracts.Instance) instancecontracts.InstanceResp
	ToInstanceRespPtr(source *instancecontracts.Instance) *instancecontracts.InstanceResp
}

var runtimeResponseMapper instanceResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}
