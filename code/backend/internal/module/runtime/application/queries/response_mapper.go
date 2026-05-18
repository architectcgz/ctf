package queries

import (
	"time"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

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
