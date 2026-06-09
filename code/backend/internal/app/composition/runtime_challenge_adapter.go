package composition

import (
	"context"
	"fmt"
	"strings"

	challengeports "ctf-platform/internal/module/challenge/ports"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type challengeRuntimeProbeAdapter struct {
	cleaner     *runtimecmd.RuntimeCleanupService
	provisioner *runtimecmd.ProvisioningService
	accessHost  string
}

func newChallengeRuntimeProbeAdapter(cleaner *runtimecmd.RuntimeCleanupService, provisioner *runtimecmd.ProvisioningService, accessHost string) challengeports.ChallengeRuntimeProbe {
	if cleaner == nil && provisioner == nil {
		return nil
	}
	return &challengeRuntimeProbeAdapter{
		cleaner:     cleaner,
		provisioner: provisioner,
		accessHost:  accessHost,
	}
}

func (a *challengeRuntimeProbeAdapter) CreateTopology(ctx context.Context, req *challengeports.RuntimeTopologyCreateRequest) (*challengeports.RuntimeTopologyCreateResult, error) {
	if a == nil || a.provisioner == nil {
		return nil, fmt.Errorf("runtime provisioning service is not configured")
	}
	if req == nil {
		return nil, fmt.Errorf("runtime topology create request is nil")
	}
	result, err := a.provisioner.CreateTopology(ctx, toRuntimeTopologyCreateRequestFromChallenge(req, a.accessHost))
	if err != nil {
		return nil, err
	}
	return &challengeports.RuntimeTopologyCreateResult{
		AccessURL:      result.AccessURL,
		RuntimeDetails: result.RuntimeDetails,
	}, nil
}

func (a *challengeRuntimeProbeAdapter) CreateContainer(ctx context.Context, imageName string, env map[string]string) (string, runtimecontracts.InstanceRuntimeDetails, error) {
	if a == nil || a.provisioner == nil {
		return "", runtimecontracts.InstanceRuntimeDetails{}, fmt.Errorf("runtime provisioning service is not configured")
	}

	result, err := a.provisioner.CreateTopology(ctx, buildRuntimeSingleContainerCreateRequestForChallenge(imageName, env))
	if err != nil {
		return "", runtimecontracts.InstanceRuntimeDetails{}, err
	}
	return result.AccessURL, result.RuntimeDetails, nil
}

func (a *challengeRuntimeProbeAdapter) CleanupRuntimeDetails(ctx context.Context, details runtimecontracts.InstanceRuntimeDetails) error {
	if a == nil || a.cleaner == nil {
		return nil
	}

	rawDetails, err := runtimecontracts.EncodeInstanceRuntimeDetails(details)
	if err != nil {
		return err
	}
	instance := &instancecontracts.Instance{
		RuntimeDetails: rawDetails,
	}
	return a.cleaner.CleanupRuntime(ctx, instance)
}

func toRuntimeTopologyCreateRequestFromChallenge(req *challengeports.RuntimeTopologyCreateRequest, publishedAccessHost string) *runtimeports.TopologyCreateRequest {
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
			Env:             cloneCompositionStringMap(node.Env),
			ServicePort:     node.ServicePort,
			ServiceProtocol: node.ServiceProtocol,
			IsEntryPoint:    node.IsEntryPoint,
			NetworkKeys:     append([]string(nil), node.NetworkKeys...),
			Resources:       cloneCompositionResourceLimits(node.Resources),
		})
	}
	return &runtimeports.TopologyCreateRequest{
		SubnetPool:                 runtimeports.SubnetPoolTopology,
		Networks:                   networks,
		Nodes:                      nodes,
		Policies:                   cloneCompositionTrafficPolicies(req.Policies),
		DisableEntryPortPublishing: strings.TrimSpace(publishedAccessHost) == "",
	}
}

func buildRuntimeSingleContainerCreateRequestForChallenge(imageName string, env map[string]string) *runtimeports.TopologyCreateRequest {
	return &runtimeports.TopologyCreateRequest{
		SubnetPool: runtimeports.SubnetPoolSingleContainer,
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{
				Key:             "default",
				Image:           imageName,
				Env:             cloneCompositionStringMap(env),
				IsEntryPoint:    true,
				NetworkKeys:     []string{runtimecontracts.TopologyDefaultNetworkKey},
				ServiceProtocol: runtimecontracts.ChallengeTargetProtocolHTTP,
			},
		},
	}
}

func cloneCompositionStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneCompositionResourceLimits(input *runtimecontracts.ResourceLimits) *runtimecontracts.ResourceLimits {
	if input == nil {
		return nil
	}
	return &runtimecontracts.ResourceLimits{
		CPUQuota:  input.CPUQuota,
		Memory:    input.Memory,
		PidsLimit: input.PidsLimit,
	}
}

func cloneCompositionTrafficPolicies(input []runtimecontracts.TopologyTrafficPolicy) []runtimecontracts.TopologyTrafficPolicy {
	if len(input) == 0 {
		return nil
	}
	output := make([]runtimecontracts.TopologyTrafficPolicy, len(input))
	for idx, policy := range input {
		output[idx] = policy
		output[idx].Ports = append([]int(nil), policy.Ports...)
	}
	return output
}
