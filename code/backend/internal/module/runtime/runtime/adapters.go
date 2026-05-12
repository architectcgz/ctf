package runtime

import (
	"context"
	"fmt"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
	opsports "ctf-platform/internal/module/ops/ports"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type runtimeOpsStatsProviderAdapter struct {
	service *runtimeapp.ContainerStatsService
}

func newRuntimeOpsStatsProvider(service *runtimeapp.ContainerStatsService) opsports.RuntimeStatsProvider {
	return &runtimeOpsStatsProviderAdapter{service: service}
}

func (p *runtimeOpsStatsProviderAdapter) ListManagedContainerStats(ctx context.Context) ([]opsports.ManagedContainerStat, error) {
	if p == nil || p.service == nil {
		return []opsports.ManagedContainerStat{}, nil
	}

	stats, err := p.service.ListManagedContainerStats(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]opsports.ManagedContainerStat, 0, len(stats))
	for _, item := range stats {
		result = append(result, opsports.ManagedContainerStat{
			ContainerID:   item.ContainerID,
			ContainerName: item.ContainerName,
			CPUPercent:    item.CPUPercent,
			MemoryPercent: item.MemoryPercent,
			MemoryUsage:   item.MemoryUsage,
			MemoryLimit:   item.MemoryLimit,
		})
	}
	return result, nil
}

func cloneRuntimeStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneRuntimeResourceLimits(input *model.ResourceLimits) *model.ResourceLimits {
	if input == nil {
		return nil
	}
	return &model.ResourceLimits{
		CPUQuota:  input.CPUQuota,
		Memory:    input.Memory,
		PidsLimit: input.PidsLimit,
	}
}

type runtimeChallengeServiceAdapter struct {
	cleaner     *runtimecmd.RuntimeCleanupService
	provisioner *runtimecmd.ProvisioningService
	publicHost  string
}

func newRuntimeChallengeServiceAdapter(cleaner *runtimecmd.RuntimeCleanupService, provisioner *runtimecmd.ProvisioningService, publicHost string) challengeports.ChallengeRuntimeProbe {
	if cleaner == nil && provisioner == nil {
		return nil
	}
	return &runtimeChallengeServiceAdapter{
		cleaner:     cleaner,
		provisioner: provisioner,
		publicHost:  publicHost,
	}
}

func (a *runtimeChallengeServiceAdapter) CreateTopology(ctx context.Context, req *challengeports.RuntimeTopologyCreateRequest) (*challengeports.RuntimeTopologyCreateResult, error) {
	if a == nil || a.provisioner == nil {
		return nil, fmt.Errorf("runtime provisioning service is not configured")
	}
	if req == nil {
		return nil, fmt.Errorf("runtime topology create request is nil")
	}
	result, err := a.provisioner.CreateTopology(ctx, toRuntimeChallengeTopologyCreateRequest(req))
	if err != nil {
		return nil, err
	}
	return &challengeports.RuntimeTopologyCreateResult{
		AccessURL:      result.AccessURL,
		RuntimeDetails: result.RuntimeDetails,
	}, nil
}

func (a *runtimeChallengeServiceAdapter) CreateContainer(ctx context.Context, imageName string, env map[string]string) (string, model.InstanceRuntimeDetails, error) {
	if a == nil || a.provisioner == nil {
		return "", model.InstanceRuntimeDetails{}, fmt.Errorf("runtime provisioning service is not configured")
	}

	containerID, networkID, hostPort, servicePort, err := a.provisioner.CreateContainer(ctx, imageName, env, 0)
	if err != nil {
		return "", model.InstanceRuntimeDetails{}, err
	}

	accessURL := fmt.Sprintf("http://%s:%d", a.publicHost, hostPort)
	return accessURL, model.InstanceRuntimeDetails{
		Networks: []model.InstanceRuntimeNetwork{
			{
				Key:       model.TopologyDefaultNetworkKey,
				Name:      model.TopologyDefaultNetworkKey,
				NetworkID: networkID,
			},
		},
		Containers: []model.InstanceRuntimeContainer{
			{
				NodeKey:         "default",
				ContainerID:     containerID,
				ServicePort:     servicePort,
				ServiceProtocol: model.ChallengeTargetProtocolHTTP,
				HostPort:        hostPort,
				IsEntryPoint:    true,
				NetworkKeys:     []string{model.TopologyDefaultNetworkKey},
			},
		},
	}, nil
}

func (a *runtimeChallengeServiceAdapter) CleanupRuntimeDetails(ctx context.Context, details model.InstanceRuntimeDetails) error {
	if a == nil || a.cleaner == nil {
		return nil
	}

	rawDetails, err := model.EncodeInstanceRuntimeDetails(details)
	if err != nil {
		return err
	}
	instance := &model.Instance{
		RuntimeDetails: rawDetails,
	}
	return a.cleaner.CleanupRuntime(ctx, instance)
}

func toRuntimeChallengeTopologyCreateRequest(req *challengeports.RuntimeTopologyCreateRequest) *runtimeports.TopologyCreateRequest {
	if req == nil {
		return nil
	}
	networks := make([]runtimeports.TopologyCreateNetwork, 0, len(req.Networks))
	for _, network := range req.Networks {
		networks = append(networks, runtimeports.TopologyCreateNetwork{
			Key:      network.Key,
			Internal: network.Internal,
		})
	}

	nodes := make([]runtimeports.TopologyCreateNode, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		nodes = append(nodes, runtimeports.TopologyCreateNode{
			Key:             node.Key,
			Image:           node.Image,
			Env:             cloneRuntimeStringMap(node.Env),
			ServicePort:     node.ServicePort,
			ServiceProtocol: node.ServiceProtocol,
			IsEntryPoint:    node.IsEntryPoint,
			NetworkKeys:     append([]string(nil), node.NetworkKeys...),
			Resources:       cloneRuntimeResourceLimits(node.Resources),
		})
	}
	return &runtimeports.TopologyCreateRequest{
		Networks:                   networks,
		Nodes:                      nodes,
		Policies:                   append([]model.TopologyTrafficPolicy(nil), req.Policies...),
		DisableEntryPortPublishing: true,
	}
}
