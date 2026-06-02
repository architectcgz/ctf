package commands

import (
	"context"
	"testing"
	"time"

	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
)

type typedNilCleanupEngine struct{}

func (*typedNilCleanupEngine) StopContainer(context.Context, string, time.Duration) error {
	return nil
}

func (*typedNilCleanupEngine) RemoveContainer(context.Context, string, bool) error {
	return nil
}

func (*typedNilCleanupEngine) RemoveNetwork(context.Context, string) error {
	return nil
}

func (*typedNilCleanupEngine) RemoveACLRules(context.Context, []runtimecontracts.InstanceRuntimeACLRule) error {
	return nil
}

func (*typedNilCleanupEngine) RemoveACL(context.Context, *runtimecontracts.InstanceRuntimeACLHandle) error {
	return nil
}

type typedNilProvisioningEngine struct{}

func (*typedNilProvisioningEngine) CreateNetwork(context.Context, string, map[string]string, bool, bool, string) (string, error) {
	return "", nil
}

func (*typedNilProvisioningEngine) ListNetworkSubnets(context.Context) ([]string, error) {
	return nil, nil
}

func (*typedNilProvisioningEngine) CreateContainer(context.Context, *runtimecontracts.ContainerConfig) (string, error) {
	return "", nil
}

func (*typedNilProvisioningEngine) ResolveServicePort(context.Context, string, int) (int, error) {
	return 0, nil
}

func (*typedNilProvisioningEngine) ConnectContainerToNetwork(context.Context, string, string) error {
	return nil
}

func (*typedNilProvisioningEngine) InspectContainerNetworkIPs(context.Context, string) (map[string]string, error) {
	return nil, nil
}

func (*typedNilProvisioningEngine) StartContainer(context.Context, string) error {
	return nil
}

func (*typedNilProvisioningEngine) StopContainer(context.Context, string, time.Duration) error {
	return nil
}

func (*typedNilProvisioningEngine) RemoveContainer(context.Context, string, bool) error {
	return nil
}

func (*typedNilProvisioningEngine) RemoveNetwork(context.Context, string) error {
	return nil
}

func (*typedNilProvisioningEngine) ApplyACLRules(context.Context, []runtimecontracts.InstanceRuntimeACLRule) error {
	return nil
}

func (*typedNilProvisioningEngine) ApplyACL(context.Context, *runtimecontracts.InstanceRuntimeACLHandle, []runtimecontracts.InstanceRuntimeACLRule) error {
	return nil
}

func TestNewRuntimeCleanupServiceTreatsTypedNilEngineAsNil(t *testing.T) {
	t.Parallel()

	var typedNil *typedNilCleanupEngine
	service := NewRuntimeCleanupService(typedNil, nil, nil)
	if service.engine != nil {
		t.Fatalf("expected typed nil engine to be normalized to nil, got %#v", service.engine)
	}
}

func TestNewProvisioningServiceTreatsTypedNilDependenciesAsNil(t *testing.T) {
	t.Parallel()

	var typedNilEngine *typedNilProvisioningEngine
	service := NewProvisioningService(nil, typedNilEngine, nil, nil)
	if service.engine != nil {
		t.Fatalf("expected typed nil engine to be normalized to nil, got %#v", service.engine)
	}
}
