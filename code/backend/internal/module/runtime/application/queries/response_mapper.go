package queries

import (
	"time"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceports "ctf-platform/internal/module/instance/ports"
)

type instanceResponseMapper interface {
	// goverter:ignore Status
	// goverter:ignore AccessURL
	// goverter:ignore Access
	// goverter:ignore RemainingTime
	// goverter:ignore RemainingExtends
	ToInstanceInfo(source instanceports.UserVisibleInstanceRow) instancecontracts.InstanceInfo
	ToInstanceInfoPtr(source *instanceports.UserVisibleInstanceRow) *instancecontracts.InstanceInfo
}

var runtimeResponseMapper instanceResponseMapper

func CopyTime(value time.Time) time.Time {
	return value
}
