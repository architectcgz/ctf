package domain

import (
	"reflect"
	"testing"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
)

func TestBuildRuntimeTopologyPlanSplitsNetworkByBroadDeny(t *testing.T) {
	plan := BuildRuntimeTopologyPlan(challengecontracts.TopologySpec{
		Networks: []challengecontracts.TopologyNetwork{
			{Key: "backend", Internal: true},
		},
		Nodes: []challengecontracts.TopologyNode{
			{Key: "web", NetworkKeys: []string{"backend"}},
			{Key: "db", NetworkKeys: []string{"backend"}},
			{Key: "cache", NetworkKeys: []string{"backend"}},
		},
		Policies: []challengecontracts.TopologyTrafficPolicy{
			{SourceNodeKey: "web", TargetNodeKey: "cache", Action: challengecontracts.TopologyPolicyActionDeny},
		},
	})

	expected := map[string][]string{
		"web":   {"backend__web__db"},
		"db":    {"backend__web__db", "backend__db__cache"},
		"cache": {"backend__db__cache"},
	}
	if !reflect.DeepEqual(plan.NodeNetworkKeys, expected) {
		t.Fatalf("unexpected node network plan: %+v", plan.NodeNetworkKeys)
	}
	if len(plan.Networks) != 2 {
		t.Fatalf("unexpected runtime network count: %+v", plan.Networks)
	}
}

func TestBuildRuntimeTopologyPlanUsesAllowListModeForFineGrainedAllow(t *testing.T) {
	plan := BuildRuntimeTopologyPlan(challengecontracts.TopologySpec{
		Networks: []challengecontracts.TopologyNetwork{
			{Key: "backend", Internal: true},
		},
		Nodes: []challengecontracts.TopologyNode{
			{Key: "web", NetworkKeys: []string{"backend"}},
			{Key: "db", NetworkKeys: []string{"backend"}},
			{Key: "cache", NetworkKeys: []string{"backend"}},
		},
		Policies: []challengecontracts.TopologyTrafficPolicy{
			{SourceNodeKey: "web", TargetNodeKey: "db", Action: challengecontracts.TopologyPolicyActionAllow, Protocol: challengecontracts.TopologyPolicyProtocolTCP, Ports: []int{3306}},
		},
	})

	expected := map[string][]string{
		"web":   {"backend__web__db"},
		"db":    {"backend__web__db"},
		"cache": {"backend__cache"},
	}
	if !reflect.DeepEqual(plan.NodeNetworkKeys, expected) {
		t.Fatalf("unexpected allow-list network plan: %+v", plan.NodeNetworkKeys)
	}
}
