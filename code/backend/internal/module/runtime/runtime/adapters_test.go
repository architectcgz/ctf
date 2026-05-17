package runtime

import (
	"testing"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

func TestRuntimeChallengeTopologyAdapterPreservesRuntimeFields(t *testing.T) {
	req := &challengeTopologyCreateRequestStub{
		Networks: []challengeTopologyCreateNetworkStub{
			{Key: model.TopologyDefaultNetworkKey, Internal: true},
		},
		Nodes: []challengeTopologyCreateNodeStub{
			{
				Key:             "web",
				Image:           "ctf/web:latest",
				Env:             map[string]string{"MODE": "awd"},
				ServicePort:     8080,
				ServiceProtocol: model.ChallengeTargetProtocolHTTP,
				IsEntryPoint:    true,
				NetworkKeys:     []string{model.TopologyDefaultNetworkKey},
				Resources:       &model.ResourceLimits{CPUQuota: 50000, Memory: 256 * 1024 * 1024, PidsLimit: 128},
			},
		},
		Policies: []model.TopologyTrafficPolicy{
			{Action: model.TopologyPolicyActionAllow, Protocol: model.TopologyPolicyProtocolTCP, Ports: []int{8080}},
		},
	}

	got := toRuntimeChallengeTopologyCreateRequest(req.toPorts(), "host-gateway.internal")
	if len(got.Networks) != 1 || !got.Networks[0].Internal {
		t.Fatalf("expected challenge runtime network fields to be preserved, got %+v", got.Networks)
	}
	if len(got.Nodes) != 1 || got.Nodes[0].Image != "ctf/web:latest" {
		t.Fatalf("expected AWD network aliases to be preserved, got %+v", got.Nodes)
	}
	if got.Nodes[0].ServicePort != 8080 || got.Nodes[0].ServiceProtocol != model.ChallengeTargetProtocolHTTP {
		t.Fatalf("expected runtime service fields to be preserved, got %+v", got.Nodes[0])
	}
	if got.Nodes[0].Resources == nil || got.Nodes[0].Resources.CPUQuota != 50000 {
		t.Fatalf("expected runtime resources to be preserved, got %+v", got.Nodes[0].Resources)
	}
	if len(got.Policies) != 1 || len(got.Policies[0].Ports) != 1 || got.Policies[0].Ports[0] != 8080 {
		t.Fatalf("expected policies preserved, got %+v", got.Policies)
	}
	if got.DisableEntryPortPublishing {
		t.Fatalf("expected published entry port when runtime access host is configured, got %+v", got)
	}
}

func TestRuntimeChallengeTopologyAdapterDisablesPublishedEntryPortWithoutAccessHost(t *testing.T) {
	req := &challengeTopologyCreateRequestStub{
		Networks: []challengeTopologyCreateNetworkStub{
			{Key: model.TopologyDefaultNetworkKey},
		},
		Nodes: []challengeTopologyCreateNodeStub{
			{
				Key:             "web",
				Image:           "ctf/web:latest",
				ServicePort:     8080,
				ServiceProtocol: model.ChallengeTargetProtocolHTTP,
				IsEntryPoint:    true,
				NetworkKeys:     []string{model.TopologyDefaultNetworkKey},
			},
		},
	}

	got := toRuntimeChallengeTopologyCreateRequest(req.toPorts(), "")
	if !got.DisableEntryPortPublishing {
		t.Fatalf("expected private entry access without published access host, got %+v", got)
	}
}

type challengeTopologyCreateRequestStub struct {
	Networks []challengeTopologyCreateNetworkStub
	Nodes    []challengeTopologyCreateNodeStub
	Policies []model.TopologyTrafficPolicy
}

type challengeTopologyCreateNetworkStub struct {
	Key      string
	Internal bool
}

type challengeTopologyCreateNodeStub struct {
	Key             string
	Image           string
	Env             map[string]string
	ServicePort     int
	ServiceProtocol string
	IsEntryPoint    bool
	NetworkKeys     []string
	Resources       *model.ResourceLimits
}

func (r *challengeTopologyCreateRequestStub) toPorts() *challengeports.RuntimeTopologyCreateRequest {
	networks := make([]challengeports.RuntimeTopologyCreateNetwork, 0, len(r.Networks))
	for _, network := range r.Networks {
		networks = append(networks, challengeports.RuntimeTopologyCreateNetwork{
			Key:      network.Key,
			Internal: network.Internal,
		})
	}

	nodes := make([]challengeports.RuntimeTopologyCreateNode, 0, len(r.Nodes))
	for _, node := range r.Nodes {
		nodes = append(nodes, challengeports.RuntimeTopologyCreateNode{
			Key:             node.Key,
			Image:           node.Image,
			Env:             node.Env,
			ServicePort:     node.ServicePort,
			ServiceProtocol: node.ServiceProtocol,
			IsEntryPoint:    node.IsEntryPoint,
			NetworkKeys:     node.NetworkKeys,
			Resources:       node.Resources,
		})
	}

	return &challengeports.RuntimeTopologyCreateRequest{
		Networks: networks,
		Nodes:    nodes,
		Policies: append([]model.TopologyTrafficPolicy(nil), r.Policies...),
	}
}
