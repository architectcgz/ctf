package infrastructure

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"

	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

func (e *Engine) ListManagedContainers(ctx context.Context) ([]runtimecontracts.ManagedContainer, error) {
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", runtimecontracts.ProjectFilter()),
			filters.Arg("label", runtimecontracts.ManagedByFilter()),
		),
	})
	if err != nil {
		return nil, err
	}

	items := make([]runtimecontracts.ManagedContainer, 0, len(containers))
	for _, item := range containers {
		name := item.ID[:12]
		if len(item.Names) > 0 {
			name = item.Names[0]
		}
		items = append(items, runtimecontracts.ManagedContainer{
			ID:        item.ID,
			Name:      name,
			CreatedAt: time.Unix(item.Created, 0),
		})
	}
	return items, nil
}

func (e *Engine) InspectManagedContainer(ctx context.Context, containerID string) (*runtimecontracts.ManagedContainerState, error) {
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	if containerID == "" {
		return &runtimecontracts.ManagedContainerState{Exists: false}, nil
	}

	resp, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		if isRuntimeContainerNotFoundError(err) {
			return &runtimecontracts.ManagedContainerState{
				ID:     containerID,
				Exists: false,
			}, nil
		}
		return nil, err
	}

	state := &runtimecontracts.ManagedContainerState{
		ID:     resp.ID,
		Exists: true,
	}
	if resp.State != nil {
		state.Running = resp.State.Running
		state.Status = resp.State.Status
	}
	return state, nil
}
