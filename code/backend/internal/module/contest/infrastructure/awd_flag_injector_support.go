package infrastructure

import (
	"context"

	"ctf-platform/internal/model"
)

func (i *dockerAWDFlagInjector) findTargetContainers(ctx context.Context, contestID, teamID, challengeID int64) ([]string, error) {
	var instances []model.Instance
	if err := i.db.WithContext(ctx).
		Table("instances AS inst").
		Select("inst.*").
		Where("inst.challenge_id = ?", challengeID).
		Where("inst.status = ?", model.InstanceStatusRunning).
		Where(
			"(inst.contest_id = ? AND inst.team_id = ?) OR ("+
				"inst.team_id IS NULL AND EXISTS ("+
				"SELECT 1 FROM team_members AS tm "+
				"WHERE tm.contest_id = ? AND tm.team_id = ? AND tm.user_id = inst.user_id))",
			contestID, teamID, contestID, teamID,
		).
		Order("inst.id ASC").
		Scan(&instances).Error; err != nil {
		return nil, err
	}

	seen := make(map[string]struct{}, len(instances))
	containerIDs := make([]string, 0, len(instances))
	for _, instance := range instances {
		for _, containerID := range collectInstanceContainerIDs(&instance) {
			if _, exists := seen[containerID]; exists || containerID == "" {
				continue
			}
			seen[containerID] = struct{}{}
			containerIDs = append(containerIDs, containerID)
		}
	}
	return containerIDs, nil
}

func collectInstanceContainerIDs(instance *model.Instance) []string {
	if instance == nil {
		return nil
	}
	details, err := model.DecodeInstanceRuntimeDetails(instance.RuntimeDetails)
	if err != nil || len(details.Containers) == 0 {
		if instance.ContainerID == "" {
			return nil
		}
		return []string{instance.ContainerID}
	}

	ids := make([]string, 0, len(details.Containers))
	for _, item := range details.Containers {
		if item.ContainerID != "" {
			ids = append(ids, item.ContainerID)
		}
	}
	if len(ids) == 0 && instance.ContainerID != "" {
		return []string{instance.ContainerID}
	}
	return ids
}
