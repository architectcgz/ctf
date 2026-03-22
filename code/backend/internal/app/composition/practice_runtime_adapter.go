package composition

import (
	"context"
	"reflect"

	"ctf-platform/internal/model"
	practiceModule "ctf-platform/internal/module/practice"
	runtimeapp "ctf-platform/internal/module/runtime/application"
)

type practiceRuntimeTopologyBridge interface {
	CleanupRuntime(instance *model.Instance) error
	CreateTopology(ctx context.Context, req *runtimeapp.TopologyCreateRequest) (*runtimeapp.TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
}

type practiceRuntimeInstanceServiceAdapter struct {
	service practiceRuntimeTopologyBridge
}

func newPracticeRuntimeInstanceServiceAdapter(service practiceRuntimeTopologyBridge) *practiceRuntimeInstanceServiceAdapter {
	if isNilPracticeRuntimeTopologyBridge(service) {
		return nil
	}
	return &practiceRuntimeInstanceServiceAdapter{service: service}
}

func isNilPracticeRuntimeTopologyBridge(service practiceRuntimeTopologyBridge) bool {
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

func (a *practiceRuntimeInstanceServiceAdapter) CleanupRuntime(instance *model.Instance) error {
	if a == nil || a.service == nil {
		return nil
	}
	return a.service.CleanupRuntime(instance)
}

func (a *practiceRuntimeInstanceServiceAdapter) CreateTopology(ctx context.Context, req *practiceModule.TopologyCreateRequest) (*practiceModule.TopologyCreateResult, error) {
	if a == nil || a.service == nil || req == nil {
		return nil, nil
	}

	result, err := a.service.CreateTopology(ctx, toRuntimeTopologyCreateRequest(req))
	if err != nil {
		return nil, err
	}
	return fromRuntimeTopologyCreateResult(result), nil
}

func (a *practiceRuntimeInstanceServiceAdapter) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error) {
	if a == nil || a.service == nil {
		return "", "", 0, 0, nil
	}
	return a.service.CreateContainer(ctx, imageName, env, reservedHostPort)
}

func toRuntimeTopologyCreateRequest(req *practiceModule.TopologyCreateRequest) *runtimeapp.TopologyCreateRequest {
	if req == nil {
		return nil
	}

	networks := make([]runtimeapp.TopologyCreateNetwork, 0, len(req.Networks))
	for _, network := range req.Networks {
		networks = append(networks, runtimeapp.TopologyCreateNetwork{
			Key:      network.Key,
			Internal: network.Internal,
		})
	}

	nodes := make([]runtimeapp.TopologyCreateNode, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		nodes = append(nodes, runtimeapp.TopologyCreateNode{
			Key:          node.Key,
			Image:        node.Image,
			Env:          cloneStringMap(node.Env),
			ServicePort:  node.ServicePort,
			IsEntryPoint: node.IsEntryPoint,
			NetworkKeys:  append([]string(nil), node.NetworkKeys...),
			Resources:    cloneResourceLimits(node.Resources),
		})
	}

	return &runtimeapp.TopologyCreateRequest{
		Networks:         networks,
		Nodes:            nodes,
		Policies:         append([]model.TopologyTrafficPolicy(nil), req.Policies...),
		ReservedHostPort: req.ReservedHostPort,
	}
}

func fromRuntimeTopologyCreateResult(result *runtimeapp.TopologyCreateResult) *practiceModule.TopologyCreateResult {
	if result == nil {
		return nil
	}
	return &practiceModule.TopologyCreateResult{
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
