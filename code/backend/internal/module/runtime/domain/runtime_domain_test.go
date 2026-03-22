package domain

import (
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/model"
)

func TestExtractManagedResourcesPrefersRuntimeDetails(t *testing.T) {
	t.Parallel()

	instance := &model.Instance{
		ContainerID: "legacy-web",
		NetworkID:   "legacy-net",
		RuntimeDetails: `{
			"containers":[
				{"container_id":"web-ctr"},
				{"container_id":"db-ctr"},
				{"container_id":"web-ctr"}
			],
			"networks":[
				{"network_id":"net-a"},
				{"network_id":"net-b"},
				{"network_id":"net-a"}
			],
			"acl_rules":[
				{"comment":"ctf:acl:test"}
			]
		}`,
	}

	resources := ExtractManagedResources(instance)
	if len(resources.ContainerIDs) != 2 || resources.ContainerIDs[0] != "web-ctr" || resources.ContainerIDs[1] != "db-ctr" {
		t.Fatalf("unexpected container ids: %+v", resources.ContainerIDs)
	}
	if len(resources.NetworkIDs) != 2 || resources.NetworkIDs[0] != "net-a" || resources.NetworkIDs[1] != "net-b" {
		t.Fatalf("unexpected network ids: %+v", resources.NetworkIDs)
	}
	if len(resources.ACLRules) != 1 || resources.ACLRules[0].Comment != "ctf:acl:test" {
		t.Fatalf("unexpected acl rules: %+v", resources.ACLRules)
	}
}

func TestExtractManagedResourcesFallsBackToLegacyFields(t *testing.T) {
	t.Parallel()

	instance := &model.Instance{
		ContainerID: "legacy-web",
		NetworkID:   "legacy-net",
	}

	resources := ExtractManagedResources(instance)
	if len(resources.ContainerIDs) != 1 || resources.ContainerIDs[0] != "legacy-web" {
		t.Fatalf("unexpected container ids: %+v", resources.ContainerIDs)
	}
	if len(resources.NetworkIDs) != 1 || resources.NetworkIDs[0] != "legacy-net" {
		t.Fatalf("unexpected network ids: %+v", resources.NetworkIDs)
	}
	if len(resources.ACLRules) != 0 {
		t.Fatalf("expected no acl rules, got %+v", resources.ACLRules)
	}
}

func TestRemainingHelpersClampAtZero(t *testing.T) {
	t.Parallel()

	now := time.Now()
	if got := RemainingExtends(1, 3); got != 0 {
		t.Fatalf("RemainingExtends() = %d, want 0", got)
	}
	if got := RemainingTime(now.Add(-time.Second), now); got != 0 {
		t.Fatalf("RemainingTime() = %d, want 0", got)
	}
}

func TestResolveTopologyACLRulesGeneratesAllowAndFallbackDeny(t *testing.T) {
	t.Parallel()

	details := model.InstanceRuntimeDetails{
		Networks: []model.InstanceRuntimeNetwork{
			{Key: model.TopologyDefaultNetworkKey, Name: "runtime-net"},
		},
		Containers: []model.InstanceRuntimeContainer{
			{NodeKey: "web", ContainerID: "web-ctr", NetworkKeys: []string{model.TopologyDefaultNetworkKey}},
			{NodeKey: "db", ContainerID: "db-ctr", NetworkKeys: []string{model.TopologyDefaultNetworkKey}},
		},
	}
	ipsByContainerID := map[string]map[string]string{
		"web-ctr": {"runtime-net": "172.30.0.2"},
		"db-ctr":  {"runtime-net": "172.30.0.3"},
	}

	rules, err := ResolveTopologyACLRules([]model.TopologyTrafficPolicy{
		{
			SourceNodeKey: "web",
			TargetNodeKey: "db",
			Action:        model.TopologyPolicyActionAllow,
			Protocol:      model.TopologyPolicyProtocolTCP,
			Ports:         []int{3306},
		},
	}, details, ipsByContainerID)
	if err != nil {
		t.Fatalf("ResolveTopologyACLRules() error = %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 acl rules, got %+v", rules)
	}
	if rules[0].Action != model.TopologyPolicyActionAllow || rules[0].Protocol != model.TopologyPolicyProtocolTCP {
		t.Fatalf("unexpected allow rule: %+v", rules[0])
	}
	if rules[1].Action != model.TopologyPolicyActionDeny || rules[1].Protocol != model.TopologyPolicyProtocolAny {
		t.Fatalf("unexpected fallback deny rule: %+v", rules[1])
	}
}

func TestResolveTopologyACLRulesRejectsAllowWithoutSharedRuntimeNetwork(t *testing.T) {
	t.Parallel()

	details := model.InstanceRuntimeDetails{
		Networks: []model.InstanceRuntimeNetwork{
			{Key: "web-net", Name: "web-net"},
			{Key: "db-net", Name: "db-net"},
		},
		Containers: []model.InstanceRuntimeContainer{
			{NodeKey: "web", ContainerID: "web-ctr", NetworkKeys: []string{"web-net"}},
			{NodeKey: "db", ContainerID: "db-ctr", NetworkKeys: []string{"db-net"}},
		},
	}
	ipsByContainerID := map[string]map[string]string{
		"web-ctr": {"web-net": "172.30.0.2"},
		"db-ctr":  {"db-net": "172.30.0.3"},
	}

	_, err := ResolveTopologyACLRules([]model.TopologyTrafficPolicy{
		{
			SourceNodeKey: "web",
			TargetNodeKey: "db",
			Action:        model.TopologyPolicyActionAllow,
			Protocol:      model.TopologyPolicyProtocolTCP,
			Ports:         []int{3306},
		},
	}, details, ipsByContainerID)
	if err == nil || !strings.Contains(err.Error(), "no shared runtime network") {
		t.Fatalf("expected no shared runtime network error, got %v", err)
	}
}
