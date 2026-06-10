package domain

import (
	"time"

	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
)

// ManagedResources 表示实例占用的运行时资源集合。
type ManagedResources struct {
	ContainerIDs []string
	NetworkIDs   []string
	Subnets      []string
	HostPorts    []int
	ACLHandle    *runtimecontracts.InstanceRuntimeACLHandle
}

// ExtractManagedResources 提取清理目标持有的容器、网络和 ACL 资源标识。
func ExtractManagedResources(target runtimecontracts.RuntimeCleanupTarget) ManagedResources {
	details, err := runtimecontracts.DecodeInstanceRuntimeDetails(target.RuntimeDetails)
	if err != nil {
		return ManagedResources{
			ContainerIDs: fallbackIDs(target.ContainerID),
			NetworkIDs:   fallbackIDs(target.NetworkID),
			HostPorts:    fallbackPorts(target.HostPort),
		}
	}

	return ManagedResources{
		ContainerIDs: uniqueContainerIDs(details, target.ContainerID),
		NetworkIDs:   uniqueNetworkIDs(details, target.NetworkID),
		Subnets:      uniqueNetworkSubnets(details),
		HostPorts:    uniqueHostPorts(details, target.HostPort),
		ACLHandle:    details.ACL,
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

func uniqueContainerIDs(details runtimecontracts.InstanceRuntimeDetails, fallback string) []string {
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

func uniqueNetworkIDs(details runtimecontracts.InstanceRuntimeDetails, fallback string) []string {
	result := make([]string, 0, len(details.Networks))
	seen := make(map[string]struct{}, len(details.Networks))
	for _, item := range details.Networks {
		if item.NetworkID == "" || item.Shared {
			continue
		}
		if _, exists := seen[item.NetworkID]; exists {
			continue
		}
		seen[item.NetworkID] = struct{}{}
		result = append(result, item.NetworkID)
	}
	if len(result) == 0 && len(details.Networks) == 0 {
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

func uniqueHostPorts(details runtimecontracts.InstanceRuntimeDetails, fallback int) []int {
	result := make([]int, 0, len(details.Containers)+1)
	seen := make(map[int]struct{}, len(details.Containers)+1)
	for _, item := range details.Containers {
		if item.HostPort <= 0 {
			continue
		}
		if _, exists := seen[item.HostPort]; exists {
			continue
		}
		seen[item.HostPort] = struct{}{}
		result = append(result, item.HostPort)
	}
	if fallback > 0 {
		if _, exists := seen[fallback]; !exists {
			result = append(result, fallback)
		}
	}
	return result
}

func uniqueNetworkSubnets(details runtimecontracts.InstanceRuntimeDetails) []string {
	result := make([]string, 0, len(details.Networks))
	seen := make(map[string]struct{}, len(details.Networks))
	for _, item := range details.Networks {
		if item.Shared || item.Subnet == "" {
			continue
		}
		if _, exists := seen[item.Subnet]; exists {
			continue
		}
		seen[item.Subnet] = struct{}{}
		result = append(result, item.Subnet)
	}
	return result
}

func fallbackPorts(port int) []int {
	if port <= 0 {
		return nil
	}
	return []int{port}
}
