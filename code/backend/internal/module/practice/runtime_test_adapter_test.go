package practice

import (
	"context"
	"reflect"

	"ctf-platform/internal/model"
	runtimeapp "ctf-platform/internal/module/runtime/application"
)

type runtimeTopologyBridgeForTest interface {
	CleanupRuntime(instance *model.Instance) error
	CreateTopology(ctx context.Context, req *runtimeapp.TopologyCreateRequest) (*runtimeapp.TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
}

type runtimeInstanceServiceAdapterForTest struct {
	service runtimeTopologyBridgeForTest
}

func newRuntimeInstanceServiceAdapterForTest(service runtimeTopologyBridgeForTest) *runtimeInstanceServiceAdapterForTest {
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
	return &runtimeInstanceServiceAdapterForTest{service: service}
}

func (a *runtimeInstanceServiceAdapterForTest) CleanupRuntime(instance *model.Instance) error {
	if a == nil || a.service == nil {
		return nil
	}
	return a.service.CleanupRuntime(instance)
}

func (a *runtimeInstanceServiceAdapterForTest) CreateTopology(ctx context.Context, req *TopologyCreateRequest) (*TopologyCreateResult, error) {
	if a == nil || a.service == nil || req == nil {
		return nil, nil
	}

	result, err := a.service.CreateTopology(ctx, &runtimeapp.TopologyCreateRequest{
		Networks:         toRuntimeTopologyNetworksForPracticeTest(req.Networks),
		Nodes:            toRuntimeTopologyNodesForPracticeTest(req.Nodes),
		Policies:         append([]model.TopologyTrafficPolicy(nil), req.Policies...),
		ReservedHostPort: req.ReservedHostPort,
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return &TopologyCreateResult{
		PrimaryContainerID: result.PrimaryContainerID,
		NetworkID:          result.NetworkID,
		AccessURL:          result.AccessURL,
		RuntimeDetails:     result.RuntimeDetails,
	}, nil
}

func (a *runtimeInstanceServiceAdapterForTest) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error) {
	if a == nil || a.service == nil {
		return "", "", 0, 0, nil
	}
	return a.service.CreateContainer(ctx, imageName, env, reservedHostPort)
}

func toRuntimeTopologyNetworksForPracticeTest(items []TopologyCreateNetwork) []runtimeapp.TopologyCreateNetwork {
	result := make([]runtimeapp.TopologyCreateNetwork, 0, len(items))
	for _, item := range items {
		result = append(result, runtimeapp.TopologyCreateNetwork{
			Key:      item.Key,
			Internal: item.Internal,
		})
	}
	return result
}

func toRuntimeTopologyNodesForPracticeTest(items []TopologyCreateNode) []runtimeapp.TopologyCreateNode {
	result := make([]runtimeapp.TopologyCreateNode, 0, len(items))
	for _, item := range items {
		result = append(result, runtimeapp.TopologyCreateNode{
			Key:          item.Key,
			Image:        item.Image,
			Env:          cloneStringMapForPracticeTest(item.Env),
			ServicePort:  item.ServicePort,
			IsEntryPoint: item.IsEntryPoint,
			NetworkKeys:  append([]string(nil), item.NetworkKeys...),
			Resources:    cloneResourceLimitsForPracticeTest(item.Resources),
		})
	}
	return result
}

func cloneStringMapForPracticeTest(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneResourceLimitsForPracticeTest(input *model.ResourceLimits) *model.ResourceLimits {
	if input == nil {
		return nil
	}
	return &model.ResourceLimits{
		CPUQuota:  input.CPUQuota,
		Memory:    input.Memory,
		PidsLimit: input.PidsLimit,
	}
}
