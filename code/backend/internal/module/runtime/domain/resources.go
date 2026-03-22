package domain

import (
	"time"

	"ctf-platform/internal/model"
)

// ManagedResources 表示实例占用的运行时资源集合。
type ManagedResources struct {
	ContainerIDs []string
	NetworkIDs   []string
	ACLRules     []model.InstanceRuntimeACLRule
}

// ExtractManagedResources 提取实例持有的容器、网络和 ACL 资源标识。
func ExtractManagedResources(instance *model.Instance) ManagedResources {
	if instance == nil {
		return ManagedResources{}
	}

	details, err := model.DecodeInstanceRuntimeDetails(instance.RuntimeDetails)
	if err != nil {
		return ManagedResources{
			ContainerIDs: fallbackIDs(instance.ContainerID),
			NetworkIDs:   fallbackIDs(instance.NetworkID),
		}
	}

	return ManagedResources{
		ContainerIDs: uniqueContainerIDs(details, instance.ContainerID),
		NetworkIDs:   uniqueNetworkIDs(details, instance.NetworkID),
		ACLRules:     append([]model.InstanceRuntimeACLRule(nil), details.ACLRules...),
	}
}

// RemainingExtends 计算实例剩余可续期次数。
func RemainingExtends(maxExtends int, extendCount int) int {
	remaining := maxExtends - extendCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RemainingTime 计算实例剩余有效秒数。
func RemainingTime(expiresAt, now time.Time) int64 {
	remaining := int64(expiresAt.Sub(now).Seconds())
	if remaining < 0 {
		return 0
	}
	return remaining
}

func uniqueContainerIDs(details model.InstanceRuntimeDetails, fallback string) []string {
	result := make([]string, 0, len(details.Containers))
	seen := make(map[string]struct{}, len(details.Containers))
	for _, item := range details.Containers {
		if item.ContainerID == "" {
			continue
		}
		if _, exists := seen[item.ContainerID]; exists {
			continue
		}
		seen[item.ContainerID] = struct{}{}
		result = append(result, item.ContainerID)
	}
	if len(result) == 0 {
		return fallbackIDs(fallback)
	}
	return result
}

func uniqueNetworkIDs(details model.InstanceRuntimeDetails, fallback string) []string {
	result := make([]string, 0, len(details.Networks))
	seen := make(map[string]struct{}, len(details.Networks))
	for _, item := range details.Networks {
		if item.NetworkID == "" {
			continue
		}
		if _, exists := seen[item.NetworkID]; exists {
			continue
		}
		seen[item.NetworkID] = struct{}{}
		result = append(result, item.NetworkID)
	}
	if len(result) == 0 {
		return fallbackIDs(fallback)
	}
	return result
}

func fallbackIDs(id string) []string {
	if id == "" {
		return nil
	}
	return []string{id}
}
