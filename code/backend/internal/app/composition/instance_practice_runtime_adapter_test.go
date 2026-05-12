package composition

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

func TestPracticeRuntimeTopologyAdapterPreservesAWDNetworkFields(t *testing.T) {
	req := &practiceports.TopologyCreateRequest{
		ContainerName: "ctf-workspace-workspace-c8-t15-s21-r1",
		Networks: []practiceports.TopologyCreateNetwork{
			{Key: model.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-8", Shared: true},
		},
		Nodes: []practiceports.TopologyCreateNode{
			{
				Key:            "web",
				Command:        []string{"tail", "-f", "/dev/null"},
				WorkingDir:     "/workspace",
				IsEntryPoint:   true,
				NetworkKeys:    []string{model.TopologyDefaultNetworkKey},
				NetworkAliases: []string{"awd-c8-t15-s21"},
				Mounts: []model.ContainerMount{
					{Source: "ctf-workspace-root-c8-t15-s21-r1-src", Target: "/workspace/src", ReadOnly: false},
				},
			},
		},
		DisableEntryPortPublishing: true,
	}

	got := toRuntimeTopologyCreateRequestFromPractice(req)
	if len(got.Networks) != 1 || got.Networks[0].Name != "ctf-awd-contest-8" || !got.Networks[0].Shared {
		t.Fatalf("expected AWD network fields to be preserved, got %+v", got.Networks)
	}
	if len(got.Nodes) != 1 || len(got.Nodes[0].NetworkAliases) != 1 || got.Nodes[0].NetworkAliases[0] != "awd-c8-t15-s21" {
		t.Fatalf("expected AWD network aliases to be preserved, got %+v", got.Nodes)
	}
	if len(got.Nodes[0].Command) != 3 || got.Nodes[0].Command[0] != "tail" || got.Nodes[0].WorkingDir != "/workspace" {
		t.Fatalf("expected runtime command and working dir to be preserved, got %+v", got.Nodes[0])
	}
	if len(got.Nodes[0].Mounts) != 1 || got.Nodes[0].Mounts[0].Target != "/workspace/src" {
		t.Fatalf("expected runtime mounts to be preserved, got %+v", got.Nodes[0].Mounts)
	}
	if got.ContainerName != "ctf-workspace-workspace-c8-t15-s21-r1" {
		t.Fatalf("expected container name to be preserved, got %+v", got)
	}
}

func TestPracticeRuntimeTopologyAdapterPreservesWorkspaceShellFields(t *testing.T) {
	req := &practiceports.TopologyCreateRequest{
		ContainerName: "workspace-companion",
		Nodes: []practiceports.TopologyCreateNode{
			{
				Key:             "workspace",
				Image:           "python:3.12-alpine",
				Env:             map[string]string{"LANG": "C.UTF-8"},
				Command:         []string{"/bin/sh", "-lc", "apk add --no-cache git vim nano && exec tail -f /dev/null"},
				WorkingDir:      "/workspace",
				ServicePort:     22,
				ServiceProtocol: model.ChallengeTargetProtocolTCP,
				IsEntryPoint:    true,
				NetworkKeys:     []string{model.TopologyDefaultNetworkKey},
				NetworkAliases:  []string{"awd-c8-t15-s21-workspace"},
				Mounts: []model.ContainerMount{
					{Source: "workspace-src", Target: "/workspace/src"},
					{Source: "workspace-data", Target: "/workspace/data", ReadOnly: true},
				},
			},
		},
	}

	got := toRuntimeTopologyCreateRequestFromPractice(req)
	if got.ContainerName != "workspace-companion" {
		t.Fatalf("expected container name preserved, got %+v", got)
	}
	if len(got.Nodes) != 1 {
		t.Fatalf("expected one node, got %+v", got.Nodes)
	}
	node := got.Nodes[0]
	if !reflect.DeepEqual(node.Command, req.Nodes[0].Command) {
		t.Fatalf("expected command preserved, got %+v", node.Command)
	}
	if node.WorkingDir != req.Nodes[0].WorkingDir {
		t.Fatalf("expected working dir preserved, got %q", node.WorkingDir)
	}
	if !reflect.DeepEqual(node.Mounts, req.Nodes[0].Mounts) {
		t.Fatalf("expected mounts preserved, got %+v", node.Mounts)
	}
	if !reflect.DeepEqual(node.Env, req.Nodes[0].Env) {
		t.Fatalf("expected env preserved, got %+v", node.Env)
	}
}

func TestPracticeRuntimeServiceAdapterInspectManagedContainerDelegatesToEngine(t *testing.T) {
	adapter := newPracticeRuntimeServiceAdapter(nil, nil, &stubPracticeRuntimeEngine{
		inspectFn: func(ctx context.Context, containerID string) (*runtimeports.ManagedContainerState, error) {
			if containerID != "workspace-ctr" {
				t.Fatalf("unexpected container inspect target: %s", containerID)
			}
			return &runtimeports.ManagedContainerState{
				ID:      containerID,
				Exists:  true,
				Running: true,
				Status:  "running",
			}, nil
		},
	})

	got, err := adapter.InspectManagedContainer(context.Background(), "workspace-ctr")
	if err != nil {
		t.Fatalf("InspectManagedContainer() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected managed container state")
	}
	if got.ID != "workspace-ctr" || !got.Exists || !got.Running || got.Status != "running" {
		t.Fatalf("unexpected managed container state: %+v", got)
	}
}

func TestPracticeRuntimeServiceAdapterInspectManagedContainerPropagatesErrors(t *testing.T) {
	wantErr := errors.New("inspect failed")
	adapter := newPracticeRuntimeServiceAdapter(nil, nil, &stubPracticeRuntimeEngine{
		inspectFn: func(context.Context, string) (*runtimeports.ManagedContainerState, error) {
			return nil, wantErr
		},
	})

	got, err := adapter.InspectManagedContainer(context.Background(), "workspace-ctr")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected inspect error %v, got %v", wantErr, err)
	}
	if got != nil {
		t.Fatalf("expected nil managed container state on error, got %+v", got)
	}
}

type stubPracticeRuntimeEngine struct {
	inspectFn func(context.Context, string) (*runtimeports.ManagedContainerState, error)
}

func (s *stubPracticeRuntimeEngine) InspectManagedContainer(ctx context.Context, containerID string) (*runtimeports.ManagedContainerState, error) {
	if s.inspectFn == nil {
		return nil, nil
	}
	return s.inspectFn(ctx, containerID)
}
