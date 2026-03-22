package practice

import (
	"context"
	"reflect"

	"ctf-platform/internal/model"
	runtime "ctf-platform/internal/module/runtime"
)

type topologyCreateNode struct {
	Key          string
	Image        string
	Env          map[string]string
	ServicePort  int
	IsEntryPoint bool
	NetworkKeys  []string
	Resources    *model.ResourceLimits
}

type topologyCreateNetwork struct {
	Key      string
	Internal bool
}

type topologyCreateRequest struct {
	Networks         []topologyCreateNetwork
	Nodes            []topologyCreateNode
	Policies         []model.TopologyTrafficPolicy
	ReservedHostPort int
}

type topologyCreateResult struct {
	PrimaryContainerID string
	NetworkID          string
	AccessURL          string
	RuntimeDetails     model.InstanceRuntimeDetails
}

type runtimeTopologyBridge interface {
	CleanupRuntime(instance *model.Instance) error
	CreateTopology(ctx context.Context, req *runtime.TopologyCreateRequest) (*runtime.TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
}

type runtimeInstanceServiceAdapter struct {
	service runtimeTopologyBridge
}

func NewRuntimeInstanceServiceAdapter(service runtimeTopologyBridge) runtimeInstanceService {
	if isNilRuntimeTopologyBridge(service) {
		return nil
	}
	return &runtimeInstanceServiceAdapter{service: service}
}

func isNilRuntimeTopologyBridge(service runtimeTopologyBridge) bool {
	if service == nil {
		return true
	}
	value := reflect.ValueOf(service)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (a *runtimeInstanceServiceAdapter) CleanupRuntime(instance *model.Instance) error {
	if a == nil || a.service == nil {
		return nil
	}
	return a.service.CleanupRuntime(instance)
}

func (a *runtimeInstanceServiceAdapter) CreateTopology(ctx context.Context, req *topologyCreateRequest) (*topologyCreateResult, error) {
	if a == nil || a.service == nil || req == nil {
		return nil, nil
	}

	result, err := a.service.CreateTopology(ctx, toRuntimeTopologyCreateRequest(req))
	if err != nil {
		return nil, err
	}
	return fromRuntimeTopologyCreateResult(result), nil
}

func (a *runtimeInstanceServiceAdapter) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error) {
	if a == nil || a.service == nil {
		return "", "", 0, 0, nil
	}
	return a.service.CreateContainer(ctx, imageName, env, reservedHostPort)
}

func toRuntimeTopologyCreateRequest(req *topologyCreateRequest) *runtime.TopologyCreateRequest {
	if req == nil {
		return nil
	}

	networks := make([]runtime.TopologyCreateNetwork, 0, len(req.Networks))
	for _, network := range req.Networks {
		networks = append(networks, runtime.TopologyCreateNetwork{
			Key:      network.Key,
			Internal: network.Internal,
		})
	}

	nodes := make([]runtime.TopologyCreateNode, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		nodes = append(nodes, runtime.TopologyCreateNode{
			Key:          node.Key,
			Image:        node.Image,
			Env:          cloneStringMap(node.Env),
			ServicePort:  node.ServicePort,
			IsEntryPoint: node.IsEntryPoint,
			NetworkKeys:  append([]string(nil), node.NetworkKeys...),
			Resources:    cloneResourceLimits(node.Resources),
		})
	}

	return &runtime.TopologyCreateRequest{
		Networks:         networks,
		Nodes:            nodes,
		Policies:         append([]model.TopologyTrafficPolicy(nil), req.Policies...),
		ReservedHostPort: req.ReservedHostPort,
	}
}

func fromRuntimeTopologyCreateResult(result *runtime.TopologyCreateResult) *topologyCreateResult {
	if result == nil {
		return nil
	}
	return &topologyCreateResult{
		PrimaryContainerID: result.PrimaryContainerID,
		NetworkID:          result.NetworkID,
		AccessURL:          result.AccessURL,
		RuntimeDetails:     result.RuntimeDetails,
	}
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneResourceLimits(input *model.ResourceLimits) *model.ResourceLimits {
	if input == nil {
		return nil
	}
	return &model.ResourceLimits{
		CPUQuota:  input.CPUQuota,
		Memory:    input.Memory,
		PidsLimit: input.PidsLimit,
	}
}
