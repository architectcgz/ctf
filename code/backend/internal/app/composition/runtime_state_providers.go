package composition

import (
	"context"
	"strings"

	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

type activeContainerInventoryProvider interface {
	ListActiveContainerIDs(ctx context.Context) ([]string, error)
}

type instanceRuntimeInventorySource interface {
	ListActiveContainerInventory(ctx context.Context) ([]instancecontracts.Instance, error)
	ListContainerNodeLookupCandidates(ctx context.Context, containerID string) ([]instancecontracts.Instance, error)
}

type instanceRuntimeInventoryProvider struct {
	source instanceRuntimeInventorySource
}

func newInstanceRuntimeInventoryProvider(source instanceRuntimeInventorySource) *instanceRuntimeInventoryProvider {
	if source == nil {
		return nil
	}
	return &instanceRuntimeInventoryProvider{source: source}
}

func (p *instanceRuntimeInventoryProvider) ListActiveContainerIDs(ctx context.Context) ([]string, error) {
	if p == nil || p.source == nil {
		return nil, nil
	}

	items, err := p.source.ListActiveContainerInventory(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		ids := []string{item.ContainerID}
		details, err := runtimecontracts.DecodeInstanceRuntimeDetails(item.RuntimeDetails)
		if err == nil {
			for _, container := range details.Containers {
				ids = append(ids, container.ContainerID)
			}
		}
		for _, containerID := range ids {
			containerID = strings.TrimSpace(containerID)
			if containerID == "" {
				continue
			}
			if _, exists := seen[containerID]; exists {
				continue
			}
			seen[containerID] = struct{}{}
			result = append(result, containerID)
		}
	}
	return result, nil
}

func (p *instanceRuntimeInventoryProvider) FindRuntimeNodeIDByContainerID(ctx context.Context, containerID string) (*int64, error) {
	if p == nil || p.source == nil {
		return nil, nil
	}

	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return nil, nil
	}

	rows, err := p.source.ListContainerNodeLookupCandidates(ctx, containerID)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		if strings.TrimSpace(row.ContainerID) == containerID {
			return row.RuntimeNodeID, nil
		}
		details, err := runtimecontracts.DecodeInstanceRuntimeDetails(row.RuntimeDetails)
		if err != nil {
			continue
		}
		for _, item := range details.Containers {
			if strings.TrimSpace(item.ContainerID) == containerID {
				return row.RuntimeNodeID, nil
			}
		}
	}
	return nil, nil
}

type compositeActiveContainerInventory struct {
	providers []activeContainerInventoryProvider
}

func newCompositeActiveContainerInventory(providers ...activeContainerInventoryProvider) *compositeActiveContainerInventory {
	items := make([]activeContainerInventoryProvider, 0, len(providers))
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		items = append(items, provider)
	}
	if len(items) == 0 {
		return nil
	}
	return &compositeActiveContainerInventory{providers: items}
}

func (c *compositeActiveContainerInventory) ListActiveContainerIDs(ctx context.Context) ([]string, error) {
	if c == nil {
		return nil, nil
	}

	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, provider := range c.providers {
		ids, err := provider.ListActiveContainerIDs(ctx)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			id = strings.TrimSpace(id)
			if id == "" {
				continue
			}
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result, nil
}

type runtimeNodeContainerIndexProvider interface {
	FindRuntimeNodeIDByContainerID(ctx context.Context, containerID string) (*int64, error)
}

type compositeRuntimeNodeContainerIndex struct {
	providers []runtimeNodeContainerIndexProvider
}

func newCompositeRuntimeNodeContainerIndex(providers ...runtimeNodeContainerIndexProvider) *compositeRuntimeNodeContainerIndex {
	items := make([]runtimeNodeContainerIndexProvider, 0, len(providers))
	for _, provider := range providers {
		if provider == nil {
			continue
		}
		items = append(items, provider)
	}
	if len(items) == 0 {
		return nil
	}
	return &compositeRuntimeNodeContainerIndex{providers: items}
}

func (c *compositeRuntimeNodeContainerIndex) FindRuntimeNodeIDByContainerID(ctx context.Context, containerID string) (*int64, error) {
	if c == nil {
		return nil, nil
	}

	for _, provider := range c.providers {
		nodeID, err := provider.FindRuntimeNodeIDByContainerID(ctx, containerID)
		if err != nil {
			return nil, err
		}
		if nodeID != nil && *nodeID > 0 {
			return nodeID, nil
		}
	}
	return nil, nil
}
