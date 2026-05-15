package infrastructure

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type contestEndedRuntimeServiceStore interface {
	ListContestAWDServicesByContest(ctx context.Context, contestID int64) ([]model.ContestAWDService, error)
}

type contestEndedRuntimeCleanupService interface {
	CleanupRuntime(ctx context.Context, instance *model.Instance) error
}

type contestEndedRuntimeStateStore interface {
	ExpireInstanceRuntime(ctx context.Context, id int64) error
	FindAWDDefenseWorkspace(ctx context.Context, contestID, teamID, serviceID int64) (*model.AWDDefenseWorkspace, error)
	UpsertAWDDefenseWorkspace(ctx context.Context, workspace *model.AWDDefenseWorkspace) error
	FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error
}

type ContestEndedRuntimeCleaner struct {
	serviceStore contestEndedRuntimeServiceStore
	instanceRepo contestports.AWDServiceInstanceQuery
	runtime      contestEndedRuntimeCleanupService
	stateStore   contestEndedRuntimeStateStore
}

var _ contestports.ContestEndedRuntimeCleaner = (*ContestEndedRuntimeCleaner)(nil)

func NewContestEndedRuntimeCleaner(
	serviceStore contestEndedRuntimeServiceStore,
	instanceRepo contestports.AWDServiceInstanceQuery,
	runtime contestEndedRuntimeCleanupService,
	stateStore contestEndedRuntimeStateStore,
) *ContestEndedRuntimeCleaner {
	if serviceStore == nil || instanceRepo == nil || runtime == nil || stateStore == nil {
		return nil
	}
	return &ContestEndedRuntimeCleaner{
		serviceStore: serviceStore,
		instanceRepo: instanceRepo,
		runtime:      runtime,
		stateStore:   stateStore,
	}
}

func (c *ContestEndedRuntimeCleaner) CleanupEndedContestAWDInstances(ctx context.Context, contestID int64) error {
	if c == nil || contestID <= 0 {
		return nil
	}

	services, err := c.serviceStore.ListContestAWDServicesByContest(ctx, contestID)
	if err != nil {
		return err
	}
	if len(services) == 0 {
		return nil
	}

	serviceIDs := make([]int64, 0, len(services))
	for _, item := range services {
		serviceIDs = append(serviceIDs, item.ID)
	}

	instances, err := c.instanceRepo.ListServiceInstancesByContest(ctx, contestID, serviceIDs)
	if err != nil {
		return err
	}
	for _, item := range instances {
		workspace, err := c.loadDefenseWorkspace(ctx, contestID, item)
		if err != nil {
			return err
		}
		if err := c.cleanupDefenseWorkspaceRuntime(ctx, item.InstanceID, workspace); err != nil {
			return err
		}
		if err := c.runtime.CleanupRuntime(ctx, endedContestRuntimeView(item)); err != nil {
			return err
		}
		if err := c.stateStore.ExpireInstanceRuntime(ctx, item.InstanceID); err != nil {
			return err
		}
		if err := c.clearDefenseWorkspaceRuntimeState(ctx, item.InstanceID, workspace); err != nil {
			return err
		}
		if err := c.stateStore.FinishActiveAWDServiceOperationForInstance(
			ctx,
			item.InstanceID,
			model.AWDServiceOperationStatusFailed,
			"contest_ended",
			time.Now().UTC(),
		); err != nil {
			return err
		}
	}
	return nil
}

func (c *ContestEndedRuntimeCleaner) loadDefenseWorkspace(ctx context.Context, contestID int64, item contestports.AWDServiceInstance) (*model.AWDDefenseWorkspace, error) {
	if c == nil || c.stateStore == nil || contestID <= 0 || item.TeamID <= 0 || item.ServiceID <= 0 {
		return nil, nil
	}
	return c.stateStore.FindAWDDefenseWorkspace(ctx, contestID, item.TeamID, item.ServiceID)
}

func (c *ContestEndedRuntimeCleaner) cleanupDefenseWorkspaceRuntime(ctx context.Context, instanceID int64, workspace *model.AWDDefenseWorkspace) error {
	if c == nil || c.runtime == nil || workspace == nil {
		return nil
	}
	containerID := strings.TrimSpace(workspace.ContainerID)
	if containerID == "" {
		return nil
	}
	return c.runtime.CleanupRuntime(ctx, endedContestDefenseWorkspaceRuntimeView(instanceID, containerID))
}

func (c *ContestEndedRuntimeCleaner) clearDefenseWorkspaceRuntimeState(ctx context.Context, instanceID int64, workspace *model.AWDDefenseWorkspace) error {
	if c == nil || c.stateStore == nil || workspace == nil {
		return nil
	}
	updated := *workspace
	updated.InstanceID = instanceID
	updated.Status = model.AWDDefenseWorkspaceStatusFailed
	updated.ContainerID = ""
	return c.stateStore.UpsertAWDDefenseWorkspace(ctx, &updated)
}

func endedContestRuntimeView(item contestports.AWDServiceInstance) *model.Instance {
	serviceID := item.ServiceID
	teamID := item.TeamID
	return &model.Instance{
		ID:             item.InstanceID,
		TeamID:         &teamID,
		ServiceID:      &serviceID,
		HostPort:       item.HostPort,
		ContainerID:    item.ContainerID,
		NetworkID:      item.NetworkID,
		Status:         item.Status,
		AccessURL:      item.AccessURL,
		RuntimeDetails: item.RuntimeDetails,
	}
}

func endedContestDefenseWorkspaceRuntimeView(instanceID int64, containerID string) *model.Instance {
	return &model.Instance{
		ID:          instanceID,
		ContainerID: strings.TrimSpace(containerID),
	}
}
