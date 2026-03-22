package app

import (
	"context"
	"reflect"

	"ctf-platform/internal/model"
	practiceModule "ctf-platform/internal/module/practice"
	runtimeapp "ctf-platform/internal/module/runtime/application"
)

type practiceRuntimeTopologyBridgeForTest interface {
	CleanupRuntime(instance *model.Instance) error
	CreateTopology(ctx context.Context, req *runtimeapp.TopologyCreateRequest) (*runtimeapp.TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
}

type practiceRuntimeInstanceServiceAdapterForTest struct {
	service practiceRuntimeTopologyBridgeForTest
}

func newPracticeRuntimeInstanceServiceAdapterForTest(service practiceRuntimeTopologyBridgeForTest) *practiceRuntimeInstanceServiceAdapterForTest {
	if service == nil {
		return nil
	}
	value := reflect.ValueOf(service)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		if value.IsNil() {
			return nil
		}
	}
	return &practiceRuntimeInstanceServiceAdapterForTest{service: service}
}

func (a *practiceRuntimeInstanceServiceAdapterForTest) CleanupRuntime(instance *model.Instance) error {
	if a == nil || a.service == nil {
		return nil
	}
	return a.service.CleanupRuntime(instance)
}

func (a *practiceRuntimeInstanceServiceAdapterForTest) CreateTopology(ctx context.Context, req *practiceModule.TopologyCreateRequest) (*practiceModule.TopologyCreateResult, error) {
	if a == nil || a.service == nil || req == nil {
		return nil, nil
	}

	result, err := a.service.CreateTopology(ctx, &runtimeapp.TopologyCreateRequest{
		Networks:         toRuntimeTopologyNetworksForTest(req.Networks),
		Nodes:            toRuntimeTopologyNodesForTest(req.Nodes),
		Policies:         append([]model.TopologyTrafficPolicy(nil), req.Policies...),
		ReservedHostPort: req.ReservedHostPort,
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return &practiceModule.TopologyCreateResult{
		PrimaryContainerID: result.PrimaryContainerID,
		NetworkID:          result.NetworkID,
		AccessURL:          result.AccessURL,
		RuntimeDetails:     result.RuntimeDetails,
	}, nil
}

func (a *practiceRuntimeInstanceServiceAdapterForTest) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error) {
	if a == nil || a.service == nil {
		return "", "", 0, 0, nil
	}
	return a.service.CreateContainer(ctx, imageName, env, reservedHostPort)
}

func toRuntimeTopologyNetworksForTest(items []practiceModule.TopologyCreateNetwork) []runtimeapp.TopologyCreateNetwork {
	result := make([]runtimeapp.TopologyCreateNetwork, 0, len(items))
	for _, item := range items {
		result = append(result, runtimeapp.TopologyCreateNetwork{
			Key:      item.Key,
			Internal: item.Internal,
		})
	}
	return result
}

func toRuntimeTopologyNodesForTest(items []practiceModule.TopologyCreateNode) []runtimeapp.TopologyCreateNode {
	result := make([]runtimeapp.TopologyCreateNode, 0, len(items))
	for _, item := range items {
		result = append(result, runtimeapp.TopologyCreateNode{
			Key:          item.Key,
			Image:        item.Image,
			Env:          cloneStringMapForTest(item.Env),
			ServicePort:  item.ServicePort,
			IsEntryPoint: item.IsEntryPoint,
			NetworkKeys:  append([]string(nil), item.NetworkKeys...),
			Resources:    cloneResourceLimitsForTest(item.Resources),
		})
	}
	return result
}

func cloneStringMapForTest(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneResourceLimitsForTest(input *model.ResourceLimits) *model.ResourceLimits {
	if input == nil {
		return nil
	}
	return &model.ResourceLimits{
		CPUQuota:  input.CPUQuota,
		Memory:    input.Memory,
		PidsLimit: input.PidsLimit,
	}
}
