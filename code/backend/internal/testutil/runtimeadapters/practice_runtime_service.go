package runtimeadapters

import (
	"context"
	"reflect"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type practiceRuntimeCleaner interface {
	CleanupRuntime(instance *model.Instance) error
}

type practiceRuntimeProvisioner interface {
	CreateTopology(ctx context.Context, req *runtimeports.TopologyCreateRequest) (*runtimeports.TopologyCreateResult, error)
	CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
}

// PracticeRuntimeService 为测试提供 practice 所需的 runtime bridge。
type PracticeRuntimeService struct {
	cleaner     practiceRuntimeCleaner
	provisioner practiceRuntimeProvisioner
}

// NewPracticeRuntimeService 创建 practice runtime 测试桥接。
func NewPracticeRuntimeService(cleaner practiceRuntimeCleaner, provisioner practiceRuntimeProvisioner) *PracticeRuntimeService {
	if isNilDependency(cleaner) && isNilDependency(provisioner) {
		return nil
	}
	return &PracticeRuntimeService{
		cleaner:     cleaner,
		provisioner: provisioner,
	}
}

func (a *PracticeRuntimeService) CleanupRuntime(instance *model.Instance) error {
	if a == nil || a.cleaner == nil {
		return nil
	}
	return a.cleaner.CleanupRuntime(instance)
}

func (a *PracticeRuntimeService) CreateTopology(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
	if a == nil || a.provisioner == nil || req == nil {
		return nil, nil
	}

	result, err := a.provisioner.CreateTopology(ctx, &runtimeports.TopologyCreateRequest{
		Networks:         toRuntimeTopologyNetworks(req.Networks),
		Nodes:            toRuntimeTopologyNodes(req.Nodes),
		Policies:         append([]model.TopologyTrafficPolicy(nil), req.Policies...),
		ReservedHostPort: req.ReservedHostPort,
	})
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	return &practiceports.TopologyCreateResult{
		PrimaryContainerID: result.PrimaryContainerID,
		NetworkID:          result.NetworkID,
		AccessURL:          result.AccessURL,
		RuntimeDetails:     result.RuntimeDetails,
	}, nil
}

func (a *PracticeRuntimeService) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error) {
	if a == nil || a.provisioner == nil {
		return "", "", 0, 0, nil
	}
	return a.provisioner.CreateContainer(ctx, imageName, env, reservedHostPort)
}

func isNilDependency(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func toRuntimeTopologyNetworks(items []practiceports.TopologyCreateNetwork) []runtimeports.TopologyCreateNetwork {
	result := make([]runtimeports.TopologyCreateNetwork, 0, len(items))
	for _, item := range items {
		result = append(result, runtimeports.TopologyCreateNetwork{
			Key:      item.Key,
			Internal: item.Internal,
		})
	}
	return result
}

func toRuntimeTopologyNodes(items []practiceports.TopologyCreateNode) []runtimeports.TopologyCreateNode {
	result := make([]runtimeports.TopologyCreateNode, 0, len(items))
	for _, item := range items {
		result = append(result, runtimeports.TopologyCreateNode{
			Key:          item.Key,
			Image:        item.Image,
			Env:          cloneStringMap(item.Env),
			ServicePort:  item.ServicePort,
			IsEntryPoint: item.IsEntryPoint,
			NetworkKeys:  append([]string(nil), item.NetworkKeys...),
			Resources:    cloneResourceLimits(item.Resources),
		})
	}
	return result
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
