package runtimeadapters

import (
	"context"
	"reflect"

	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

type practiceRuntimeCleaner interface {
	CleanupRuntime(instance *instanceentity.Instance) error
}

type practiceRuntimeProvisioner interface {
	CreateTopology(ctx context.Context, req *runtimecontracts.TopologyCreateRequest) (*runtimecontracts.TopologyCreateResult, error)
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

func (a *PracticeRuntimeService) CleanupRuntime(instance *instanceentity.Instance) error {
	if a == nil || a.cleaner == nil {
		return nil
	}
	return a.cleaner.CleanupRuntime(instance)
}

func (a *PracticeRuntimeService) CreateTopology(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
	if a == nil || a.provisioner == nil || req == nil {
		return nil, nil
	}

	result, err := a.provisioner.CreateTopology(ctx, &runtimecontracts.TopologyCreateRequest{
		Networks:                   toRuntimeTopologyNetworks(req.Networks),
		Nodes:                      toRuntimeTopologyNodes(req.Nodes),
		Policies:                   append([]runtimecontracts.TopologyTrafficPolicy(nil), req.Policies...),
		SubnetPool:                 req.SubnetPool,
		OwnerInstanceID:            req.OwnerInstanceID,
		ReservedHostPort:           req.ReservedHostPort,
		DisableEntryPortPublishing: req.DisableEntryPortPublishing,
		ContainerName:              req.ContainerName,
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

func toRuntimeTopologyNetworks(items []practiceports.TopologyCreateNetwork) []runtimecontracts.TopologyCreateNetwork {
	result := make([]runtimecontracts.TopologyCreateNetwork, 0, len(items))
	for _, item := range items {
		result = append(result, runtimecontracts.TopologyCreateNetwork{
			Key:      item.Key,
			Name:     item.Name,
			Subnet:   item.Subnet,
			Internal: item.Internal,
			Shared:   item.Shared,
		})
	}
	return result
}

func toRuntimeTopologyNodes(items []practiceports.TopologyCreateNode) []runtimecontracts.TopologyCreateNode {
	result := make([]runtimecontracts.TopologyCreateNode, 0, len(items))
	for _, item := range items {
		result = append(result, runtimecontracts.TopologyCreateNode{
			Key:             item.Key,
			Image:           item.Image,
			Env:             cloneStringMap(item.Env),
			Command:         append([]string(nil), item.Command...),
			WorkingDir:      item.WorkingDir,
			ServicePort:     item.ServicePort,
			ServiceProtocol: item.ServiceProtocol,
			IsEntryPoint:    item.IsEntryPoint,
			NetworkKeys:     append([]string(nil), item.NetworkKeys...),
			NetworkAliases:  append([]string(nil), item.NetworkAliases...),
			Mounts:          append([]runtimecontracts.ContainerMount(nil), item.Mounts...),
			Resources:       cloneResourceLimits(item.Resources),
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

func cloneResourceLimits(input *runtimecontracts.ResourceLimits) *runtimecontracts.ResourceLimits {
	if input == nil {
		return nil
	}
	return &runtimecontracts.ResourceLimits{
		CPUQuota:  input.CPUQuota,
		Memory:    input.Memory,
		PidsLimit: input.PidsLimit,
	}
}
