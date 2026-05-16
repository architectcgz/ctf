package infrastructure

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"

	runtimedomain "ctf-platform/internal/module/runtime/domain"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

func (e *Engine) ListManagedContainers(ctx context.Context) ([]runtimeports.ManagedContainer, error) {
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", runtimedomain.ProjectFilter()),
			filters.Arg("label", runtimedomain.ManagedByFilter()),
		),
	})
	if err != nil {
		return nil, err
	}

	items := make([]runtimeports.ManagedContainer, 0, len(containers))
	for _, item := range containers {
		name := item.ID[:12]
		if len(item.Names) > 0 {
			name = item.Names[0]
		}
		items = append(items, runtimeports.ManagedContainer{
			ID:        item.ID,
			Name:      name,
			CreatedAt: time.Unix(item.Created, 0),
		})
	}
	return items, nil
}

func (e *Engine) InspectManagedContainer(ctx context.Context, containerID string) (*runtimeports.ManagedContainerState, error) {
	cli, err := e.requireClient()
	if err != nil {
		return nil, err
	}
	if containerID == "" {
		return &runtimeports.ManagedContainerState{Exists: false}, nil
	}

	resp, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		if isRuntimeContainerNotFoundError(err) {
			return &runtimeports.ManagedContainerState{
				ID:     containerID,
				Exists: false,
			}, nil
		}
		return nil, err
	}

	state := &runtimeports.ManagedContainerState{
		ID:     resp.ID,
		Exists: true,
	}
	if resp.State != nil {
		state.Running = resp.State.Running
		state.Status = resp.State.Status
	}
	return state, nil
}
