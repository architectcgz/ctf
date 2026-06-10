package agentserver

import (
	"context"
	"io"
	"testing"
	"time"

	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
)

func TestServiceHealthReflectsAvailableDependencies(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		hostExecutor     runtimeports.RuntimeHostExecutor
		sandboxExecutor  runtimeports.SandboxExecutor
		wantReady        bool
		wantCapabilities []string
	}{
		{
			name:             "missing all dependencies",
			wantReady:        false,
			wantCapabilities: []string{},
		},
		{
			name:             "host executor only",
			hostExecutor:     &healthTestRuntimeHostExecutor{},
			wantReady:        false,
			wantCapabilities: []string{"runtime_host_execution", "interactive_exec"},
		},
		{
			name:             "checker runner only",
			sandboxExecutor:  healthTestSandboxExecutor{},
			wantReady:        false,
			wantCapabilities: []string{"checker_runner"},
		},
		{
			name:             "all dependencies",
			hostExecutor:     &healthTestRuntimeHostExecutor{},
			sandboxExecutor:  healthTestSandboxExecutor{},
			wantReady:        true,
			wantCapabilities: []string{"runtime_host_execution", "checker_runner", "interactive_exec"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := NewService(tc.hostExecutor, tc.sandboxExecutor).Health(context.Background(), nil)
			if err != nil {
				t.Fatalf("Health() error = %v", err)
			}
			if resp.Ready != tc.wantReady {
				t.Fatalf("Health().Ready = %v, want %v", resp.Ready, tc.wantReady)
			}
			if len(resp.Capabilities) != len(tc.wantCapabilities) {
				t.Fatalf("Health().Capabilities = %v, want %v", resp.Capabilities, tc.wantCapabilities)
			}
			for i := range tc.wantCapabilities {
				if resp.Capabilities[i] != tc.wantCapabilities[i] {
					t.Fatalf("Health().Capabilities = %v, want %v", resp.Capabilities, tc.wantCapabilities)
				}
			}
		})
	}
}

type healthTestSandboxExecutor struct{}

func (healthTestSandboxExecutor) RunSandboxExec(context.Context, runtimeports.SandboxExecJob) (runtimeports.SandboxExecResult, error) {
	return runtimeports.SandboxExecResult{}, nil
}

type healthTestRuntimeHostExecutor struct{}

func (*healthTestRuntimeHostExecutor) CreateNetwork(context.Context, string, map[string]string, bool, bool, string) (string, error) {
	return "", nil
}

func (*healthTestRuntimeHostExecutor) ListNetworkSubnets(context.Context) ([]string, error) {
	return nil, nil
}

func (*healthTestRuntimeHostExecutor) CreateContainer(context.Context, *runtimecontracts.ContainerConfig) (string, error) {
	return "", nil
}

func (*healthTestRuntimeHostExecutor) ResolveServicePort(context.Context, string, int) (int, error) {
	return 0, nil
}

func (*healthTestRuntimeHostExecutor) ConnectContainerToNetwork(context.Context, string, string) error {
	return nil
}

func (*healthTestRuntimeHostExecutor) InspectContainerNetworkIPs(context.Context, string) (map[string]string, error) {
	return nil, nil
}

func (*healthTestRuntimeHostExecutor) StartContainer(context.Context, string) error {
	return nil
}

func (*healthTestRuntimeHostExecutor) StopContainer(context.Context, string, time.Duration) error {
	return nil
}

func (*healthTestRuntimeHostExecutor) RemoveContainer(context.Context, string, bool) error {
	return nil
}

func (*healthTestRuntimeHostExecutor) RemoveNetwork(context.Context, string) error {
	return nil
}

func (*healthTestRuntimeHostExecutor) ApplyACLRules(context.Context, []runtimecontracts.InstanceRuntimeACLRule) error {
	return nil
}

func (*healthTestRuntimeHostExecutor) ApplyACL(context.Context, *runtimecontracts.InstanceRuntimeACLHandle, []runtimecontracts.InstanceRuntimeACLRule) error {
	return nil
}

func (*healthTestRuntimeHostExecutor) RemoveACLRules(context.Context, []runtimecontracts.InstanceRuntimeACLRule) error {
	return nil
}

func (*healthTestRuntimeHostExecutor) RemoveACL(context.Context, *runtimecontracts.InstanceRuntimeACLHandle) error {
	return nil
}

func (*healthTestRuntimeHostExecutor) WriteFileToContainer(context.Context, string, string, []byte) error {
	return nil
}

func (*healthTestRuntimeHostExecutor) ReadFileFromContainer(context.Context, string, string, int64) ([]byte, error) {
	return nil, nil
}

func (*healthTestRuntimeHostExecutor) ListDirectoryFromContainer(context.Context, string, string, int) ([]runtimecontracts.ContainerDirectoryEntry, error) {
	return nil, nil
}

func (*healthTestRuntimeHostExecutor) ExecContainerCommand(context.Context, string, []string, []byte, int64) ([]byte, error) {
	return nil, nil
}

func (*healthTestRuntimeHostExecutor) InspectImageSize(context.Context, string) (int64, error) {
	return 0, nil
}

func (*healthTestRuntimeHostExecutor) RemoveImage(context.Context, string) error {
	return nil
}

func (*healthTestRuntimeHostExecutor) ListManagedContainers(context.Context) ([]runtimecontracts.ManagedContainer, error) {
	return nil, nil
}

func (*healthTestRuntimeHostExecutor) InspectManagedContainer(context.Context, string) (*runtimecontracts.ManagedContainerState, error) {
	return nil, nil
}

func (*healthTestRuntimeHostExecutor) ListManagedContainerStats(context.Context) ([]runtimecontracts.ManagedContainerStat, error) {
	return nil, nil
}

func (*healthTestRuntimeHostExecutor) ExecContainerInteractive(context.Context, string, []string, io.Reader, io.Writer) error {
	return nil
}
