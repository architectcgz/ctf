package composition

import (
	"context"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type practiceRuntimeServiceAdapter struct {
	cleaner     *runtimecmd.RuntimeCleanupService
	provisioner *runtimecmd.ProvisioningService
	inspector   practiceManagedContainerInspector
}

type practiceManagedContainerInspector interface {
	InspectManagedContainer(ctx context.Context, containerID string) (*runtimeports.ManagedContainerState, error)
}

func newPracticeRuntimeServiceAdapter(
	cleaner *runtimecmd.RuntimeCleanupService,
	provisioner *runtimecmd.ProvisioningService,
	inspector practiceManagedContainerInspector,
) practiceports.RuntimeInstanceService {
	if cleaner == nil && provisioner == nil && inspector == nil {
		return nil
	}
	return &practiceRuntimeServiceAdapter{
		cleaner:     cleaner,
		provisioner: provisioner,
		inspector:   inspector,
	}
}

func (a *practiceRuntimeServiceAdapter) CleanupRuntime(ctx context.Context, instance *model.Instance) error {
	if a == nil || a.cleaner == nil {
		return nil
	}
	return a.cleaner.CleanupRuntime(ctx, instance)
}

func (a *practiceRuntimeServiceAdapter) CreateTopology(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
	if a == nil || a.provisioner == nil || req == nil {
		return nil, nil
	}

	result, err := a.provisioner.CreateTopology(ctx, toRuntimeTopologyCreateRequestFromPractice(req))
	if err != nil {
		return nil, err
	}
	return fromRuntimeTopologyCreateResultForPractice(result), nil
}

func (a *practiceRuntimeServiceAdapter) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error) {
	if a == nil || a.provisioner == nil {
		return "", "", 0, 0, nil
	}
	return a.provisioner.CreateContainer(ctx, imageName, env, reservedHostPort)
}

func (a *practiceRuntimeServiceAdapter) InspectManagedContainer(ctx context.Context, containerID string) (*practiceports.ManagedContainerState, error) {
	if a == nil || a.inspector == nil {
		return nil, nil
	}
	return a.inspector.InspectManagedContainer(ctx, containerID)
}

func toRuntimeTopologyCreateRequestFromPractice(req *practiceports.TopologyCreateRequest) *runtimeports.TopologyCreateRequest {
	if req == nil {
		return nil
	}

	networks := make([]runtimeports.TopologyCreateNetwork, 0, len(req.Networks))
	for _, network := range req.Networks {
		networks = append(networks, runtimeports.TopologyCreateNetwork{
			Key:      network.Key,
			Name:     network.Name,
			Internal: network.Internal,
			Shared:   network.Shared,
		})
	}

	nodes := make([]runtimeports.TopologyCreateNode, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		nodes = append(nodes, runtimeports.TopologyCreateNode{
			Key:             node.Key,
			Image:           node.Image,
			Env:             clonePracticeRuntimeStringMap(node.Env),
			Command:         append([]string(nil), node.Command...),
			WorkingDir:      node.WorkingDir,
			ServicePort:     node.ServicePort,
			ServiceProtocol: node.ServiceProtocol,
			IsEntryPoint:    node.IsEntryPoint,
			NetworkKeys:     append([]string(nil), node.NetworkKeys...),
			NetworkAliases:  append([]string(nil), node.NetworkAliases...),
			Mounts:          append([]model.ContainerMount(nil), node.Mounts...),
			Resources:       clonePracticeRuntimeResourceLimits(node.Resources),
		})
	}

	return &runtimeports.TopologyCreateRequest{
		Networks:                   networks,
		Nodes:                      nodes,
		Policies:                   append([]model.TopologyTrafficPolicy(nil), req.Policies...),
		ReservedHostPort:           req.ReservedHostPort,
		DisableEntryPortPublishing: req.DisableEntryPortPublishing,
		ContainerName:              req.ContainerName,
	}
}

func fromRuntimeTopologyCreateResultForPractice(result *runtimeports.TopologyCreateResult) *practiceports.TopologyCreateResult {
	if result == nil {
		return nil
	}
	return &practiceports.TopologyCreateResult{
		PrimaryContainerID: result.PrimaryContainerID,
		NetworkID:          result.NetworkID,
		AccessURL:          result.AccessURL,
		RuntimeDetails:     result.RuntimeDetails,
	}
}

func clonePracticeRuntimeStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func clonePracticeRuntimeResourceLimits(input *model.ResourceLimits) *model.ResourceLimits {
	if input == nil {
		return nil
	}
	return &model.ResourceLimits{
		CPUQuota:  input.CPUQuota,
		Memory:    input.Memory,
		PidsLimit: input.PidsLimit,
	}
}
