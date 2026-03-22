package runtime

import (
	"ctf-platform/internal/model"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimeinfra "ctf-platform/internal/module/runtimeinfra"
)

type ManagedContainer = runtimeinfra.ManagedContainer
type ManagedContainerStat = runtimeinfra.ManagedContainerStat
type ProxyTicketClaims = runtimeapp.ProxyTicketClaims

type InstanceRepository interface {
	UpdateRuntime(instance *model.Instance) error
	UpdateStatusAndReleasePort(id int64, status string) error
	FindByUserAndChallenge(userID, challengeID int64) (*model.Instance, error)
}
