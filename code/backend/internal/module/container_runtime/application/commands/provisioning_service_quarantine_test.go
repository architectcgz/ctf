package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"ctf-platform/internal/config"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
)

func TestCreateTopologyQuarantinesRuntimeNodeSubnetConflictAndRetriesSameNodePool(t *testing.T) {
	t.Parallel()

	repo := &nodeScopedProvisioningRepo{
		t:       t,
		subnets: []string{"10.11.0.0/29", "10.11.0.8/29"},
	}
	createNetworkCalls := 0
	engine := &stubProvisioningRuntime{
		createNetworkFn: func(ctx context.Context, name string, labels map[string]string, internal bool, allowExisting bool, subnet string) (string, error) {
			createNetworkCalls++
			switch createNetworkCalls {
			case 1:
				if subnet != "10.11.0.0/29" {
					t.Fatalf("first network create subnet = %q, want first node pool subnet", subnet)
				}
				return "", runtimeports.ErrRuntimeNetworkSubnetConflict
			case 2:
				if subnet != "10.11.0.8/29" {
					t.Fatalf("retry network create subnet = %q, want next node pool subnet", subnet)
				}
				return "network-retry", nil
			default:
				t.Fatalf("unexpected CreateNetwork call #%d", createNetworkCalls)
				return "", nil
			}
		},
		createContainerFn: func(ctx context.Context, cfg *runtimecontracts.ContainerConfig) (string, error) {
			return "container-retry", nil
		},
		inspectContainerNetworkIPsFn: func(ctx context.Context, containerID string) (map[string]string, error) {
			return map[string]string{"ctf-net-default": "10.11.0.2"}, nil
		},
	}
	service := NewProvisioningService(repo, engine, &config.ContainerConfig{
		PublicHost: "127.0.0.1",
	}, nil)

	_, err := service.CreateTopology(context.Background(), &runtimecontracts.TopologyCreateRequest{
		RuntimeNodeID:    701,
		OwnerInstanceID:  901,
		ReservedHostPort: 30080,
		SubnetPool:       runtimecontracts.SubnetPoolSingleContainer,
		Networks: []runtimecontracts.TopologyCreateNetwork{
			{Key: runtimecontracts.TopologyDefaultNetworkKey, Name: "ctf-net-default"},
		},
		Nodes: []runtimecontracts.TopologyCreateNode{
			{
				Key:             "default",
				Image:           "ctf/web:v1",
				ServicePort:     8080,
				ServiceProtocol: runtimecontracts.ChallengeTargetProtocolHTTP,
				IsEntryPoint:    true,
				NetworkKeys:     []string{runtimecontracts.TopologyDefaultNetworkKey},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if createNetworkCalls != 2 {
		t.Fatalf("expected subnet conflict retry once, got %d network creates", createNetworkCalls)
	}
	if got := repo.reservedNodeIDs; len(got) != 2 || got[0] != 701 || got[1] != 701 {
		t.Fatalf("expected retry to reserve from same runtime node pool twice, got %+v", got)
	}
	if len(repo.quarantinedSubnets) != 1 || repo.quarantinedSubnets[0] != "701:10.11.0.0/29" {
		t.Fatalf("expected conflicted subnet to be quarantined for node 701 only, got %+v", repo.quarantinedSubnets)
	}
	if got := repo.releasedSubnets; len(got) != 0 {
		t.Fatalf("expected conflicted subnet to be quarantined instead of released, got releases %+v", got)
	}
}

type nodeScopedProvisioningRepo struct {
	t                  *testing.T
	subnets            []string
	reservedNodeIDs    []int64
	quarantinedSubnets []string
	releasedSubnets    []string
}

func (r *nodeScopedProvisioningRepo) ReserveAvailablePort(ctx context.Context, start, end int) (int, error) {
	r.t.Fatalf("unexpected global port reservation")
	return 0, nil
}

func (r *nodeScopedProvisioningRepo) ReleaseReservedPort(ctx context.Context, port int) error {
	r.t.Fatalf("unexpected global port release")
	return nil
}

func (r *nodeScopedProvisioningRepo) ReserveAvailableSubnetForInstance(ctx context.Context, baseCIDR string, subnetMask int, instanceID int64, networkKey string) (string, error) {
	r.t.Fatalf("unexpected legacy instance subnet reservation")
	return "", nil
}

func (r *nodeScopedProvisioningRepo) ReserveAvailableSubnetExcluding(ctx context.Context, baseCIDR string, subnetMask int, excludedSubnets []string) (string, error) {
	r.t.Fatalf("unexpected legacy subnet reservation with exclusions")
	return "", nil
}

func (r *nodeScopedProvisioningRepo) ReserveAvailableSubnetForInstanceExcluding(ctx context.Context, baseCIDR string, subnetMask int, instanceID int64, networkKey string, excludedSubnets []string) (string, error) {
	r.t.Fatalf("unexpected legacy instance subnet reservation with exclusions")
	return "", nil
}

func (r *nodeScopedProvisioningRepo) ReserveAvailableSubnetForNode(ctx context.Context, nodeID int64, poolKind string, instanceID int64, networkKey string) (string, error) {
	r.reservedNodeIDs = append(r.reservedNodeIDs, nodeID)
	if len(r.subnets) == 0 {
		return "", errors.New("no subnets left")
	}
	next := r.subnets[0]
	r.subnets = r.subnets[1:]
	return next, nil
}

func (r *nodeScopedProvisioningRepo) ReleaseReservedSubnet(ctx context.Context, subnet string) error {
	r.releasedSubnets = append(r.releasedSubnets, subnet)
	return nil
}

func (r *nodeScopedProvisioningRepo) ReleaseSubnetForInstance(ctx context.Context, subnet string, instanceID int64) error {
	r.releasedSubnets = append(r.releasedSubnets, subnet)
	return nil
}

func (r *nodeScopedProvisioningRepo) QuarantineSubnet(ctx context.Context, nodeID int64, subnet string, reason string) error {
	r.quarantinedSubnets = append(r.quarantinedSubnets, "701:"+subnet)
	if nodeID != 701 {
		r.t.Fatalf("quarantine nodeID = %d, want 701", nodeID)
	}
	return nil
}

type stubProvisioningRuntime struct {
	createNetworkFn              func(ctx context.Context, name string, labels map[string]string, internal bool, allowExisting bool, subnet string) (string, error)
	listNetworkSubnetsFn         func(ctx context.Context) ([]string, error)
	createContainerFn            func(ctx context.Context, cfg *runtimecontracts.ContainerConfig) (string, error)
	resolveServicePortFn         func(ctx context.Context, imageRef string, preferredPort int) (int, error)
	connectContainerToNetworkFn  func(ctx context.Context, containerID, networkName string) error
	inspectContainerNetworkIPsFn func(ctx context.Context, containerID string) (map[string]string, error)
	startContainerFn             func(ctx context.Context, containerID string) error
	stopContainerFn              func(ctx context.Context, containerID string, timeout time.Duration) error
	removeContainerFn            func(ctx context.Context, containerID string, force bool) error
	removeNetworkFn              func(ctx context.Context, networkID string) error
	applyACLRulesFn              func(ctx context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error
	applyACLFn                   func(ctx context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle, rules []runtimecontracts.InstanceRuntimeACLRule) error
}

func (s *stubProvisioningRuntime) CreateNetwork(ctx context.Context, name string, labels map[string]string, internal bool, allowExisting bool, subnet string) (string, error) {
	if s.createNetworkFn != nil {
		return s.createNetworkFn(ctx, name, labels, internal, allowExisting, subnet)
	}
	return "network", nil
}

func (s *stubProvisioningRuntime) ListNetworkSubnets(ctx context.Context) ([]string, error) {
	if s.listNetworkSubnetsFn != nil {
		return s.listNetworkSubnetsFn(ctx)
	}
	return nil, nil
}

func (s *stubProvisioningRuntime) CreateContainer(ctx context.Context, cfg *runtimecontracts.ContainerConfig) (string, error) {
	if s.createContainerFn != nil {
		return s.createContainerFn(ctx, cfg)
	}
	return "container", nil
}

func (s *stubProvisioningRuntime) ResolveServicePort(ctx context.Context, imageRef string, preferredPort int) (int, error) {
	if s.resolveServicePortFn != nil {
		return s.resolveServicePortFn(ctx, imageRef, preferredPort)
	}
	return preferredPort, nil
}

func (s *stubProvisioningRuntime) ConnectContainerToNetwork(ctx context.Context, containerID, networkName string) error {
	if s.connectContainerToNetworkFn != nil {
		return s.connectContainerToNetworkFn(ctx, containerID, networkName)
	}
	return nil
}

func (s *stubProvisioningRuntime) InspectContainerNetworkIPs(ctx context.Context, containerID string) (map[string]string, error) {
	if s.inspectContainerNetworkIPsFn != nil {
		return s.inspectContainerNetworkIPsFn(ctx, containerID)
	}
	return map[string]string{"ctf-net-default": "10.11.0.2"}, nil
}

func (s *stubProvisioningRuntime) StartContainer(ctx context.Context, containerID string) error {
	if s.startContainerFn != nil {
		return s.startContainerFn(ctx, containerID)
	}
	return nil
}

func (s *stubProvisioningRuntime) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	if s.stopContainerFn != nil {
		return s.stopContainerFn(ctx, containerID, timeout)
	}
	return nil
}

func (s *stubProvisioningRuntime) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	if s.removeContainerFn != nil {
		return s.removeContainerFn(ctx, containerID, force)
	}
	return nil
}

func (s *stubProvisioningRuntime) RemoveNetwork(ctx context.Context, networkID string) error {
	if s.removeNetworkFn != nil {
		return s.removeNetworkFn(ctx, networkID)
	}
	return nil
}

func (s *stubProvisioningRuntime) ApplyACLRules(ctx context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	if s.applyACLRulesFn != nil {
		return s.applyACLRulesFn(ctx, rules)
	}
	return nil
}

func (s *stubProvisioningRuntime) ApplyACL(ctx context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	if s.applyACLFn != nil {
		return s.applyACLFn(ctx, handle, rules)
	}
	return nil
}
