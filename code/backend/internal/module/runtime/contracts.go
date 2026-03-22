package runtime

import (
	"context"

	"ctf-platform/internal/model"
	"ctf-platform/internal/module/container"
)

type ManagedContainerStat = container.ManagedContainerStat
type TopologyCreateRequest = container.TopologyCreateRequest
type TopologyCreateResult = container.TopologyCreateResult
type TopologyCreateNetwork = container.TopologyCreateNetwork
type TopologyCreateNode = container.TopologyCreateNode

type RuntimeStatsProvider interface {
	ListManagedContainerStats(ctx context.Context) ([]ManagedContainerStat, error)
}

type InstanceRepository interface {
	UpdateRuntime(instance *model.Instance) error
	UpdateStatusAndReleasePort(id int64, status string) error
	FindByUserAndChallenge(userID, challengeID int64) (*model.Instance, error)
}

type RuntimeFacade interface {
	RuntimeStatsProvider
	InspectImageSize(ctx context.Context, imageRef string) (int64, error)
	RemoveImage(ctx context.Context, imageRef string) error
	CleanupRuntime(instance *model.Instance) error
	CreateTopology(ctx context.Context, req *TopologyCreateRequest) (*TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
	CleanExpiredInstances(ctx context.Context) error
	CleanupOrphans(ctx context.Context) error
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
}
