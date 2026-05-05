package commands

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

type dockerRunnerCall struct {
	name string
	args []string
	env  []string
}

type fakeDockerRunner struct {
	calls  []dockerRunnerCall
	output string
	err    error
}

func (r *fakeDockerRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	r.calls = append(r.calls, dockerRunnerCall{name: name, args: append([]string(nil), args...)})
	return r.output, r.err
}

func (r *fakeDockerRunner) RunWithEnv(ctx context.Context, env []string, name string, args ...string) (string, error) {
	r.calls = append(r.calls, dockerRunnerCall{name: name, args: append([]string(nil), args...), env: append([]string(nil), env...)})
	return r.output, r.err
}

func TestDockerCLIImageBuilderBuildUsesExplicitDockerfileAndContext(t *testing.T) {
	runner := &fakeDockerRunner{}
	builder := newDockerCLIImageBuilderWithRunner(runner)

	if err := builder.Build(context.Background(), "/tmp/pkg/docker", "/tmp/pkg/docker/Dockerfile", "jeopardy/web:v1"); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	want := []dockerRunnerCall{{
		name: "docker",
		args: []string{"build", "-f", "/tmp/pkg/docker/Dockerfile", "-t", "jeopardy/web:v1", "/tmp/pkg/docker"},
	}}
	if !reflect.DeepEqual(runner.calls, want) {
		t.Fatalf("calls = %+v, want %+v", runner.calls, want)
	}
}

func TestDockerCLIImageBuilderInspectParsesSize(t *testing.T) {
	runner := &fakeDockerRunner{output: "12345\n"}
	builder := newDockerCLIImageBuilderWithRunner(runner)

	result, err := builder.Inspect(context.Background(), "registry.example.edu/jeopardy/web:v1")
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if result.Size != 12345 {
		t.Fatalf("Size = %d, want 12345", result.Size)
	}
}

func TestDockerCLIImageBuilderPushUsesIsolatedDockerConfig(t *testing.T) {
	runner := &fakeDockerRunner{}
	builder := newDockerCLIImageBuilderWithRunnerAndConfig(runner, DockerCLIImageBuilderConfig{
		RegistryServer: "registry.example.edu",
		Username:       "ctf",
		Password:       "registry-token",
	})

	if err := builder.Push(context.Background(), "registry.example.edu/jeopardy/web:v1"); err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if len(runner.calls) != 1 {
		t.Fatalf("calls len = %d, want 1", len(runner.calls))
	}
	call := runner.calls[0]
	if call.name != "docker" || !reflect.DeepEqual(call.args, []string{"push", "registry.example.edu/jeopardy/web:v1"}) {
		t.Fatalf("call = %+v", call)
	}
	if len(call.env) != 1 || !strings.HasPrefix(call.env[0], "DOCKER_CONFIG=") {
		t.Fatalf("env = %+v, want isolated DOCKER_CONFIG", call.env)
	}
	for _, arg := range call.args {
		if strings.Contains(arg, "registry-token") {
			t.Fatalf("docker args leaked registry password: %+v", call.args)
		}
	}
}

func TestDockerCLIImageBuilderPullUsesIsolatedDockerConfigForIdentityToken(t *testing.T) {
	runner := &fakeDockerRunner{}
	builder := newDockerCLIImageBuilderWithRunnerAndConfig(runner, DockerCLIImageBuilderConfig{
		RegistryServer: "registry.example.edu",
		IdentityToken:  "registry-identity-token",
	})

	if err := builder.Pull(context.Background(), "registry.example.edu/awd/demo:c1"); err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	if len(runner.calls) != 1 {
		t.Fatalf("calls len = %d, want 1", len(runner.calls))
	}
	call := runner.calls[0]
	if call.name != "docker" || !reflect.DeepEqual(call.args, []string{"pull", "registry.example.edu/awd/demo:c1"}) {
		t.Fatalf("call = %+v", call)
	}
	if len(call.env) != 1 || !strings.HasPrefix(call.env[0], "DOCKER_CONFIG=") {
		t.Fatalf("env = %+v, want isolated DOCKER_CONFIG", call.env)
	}
	for _, arg := range call.args {
		if strings.Contains(arg, "registry-identity-token") {
			t.Fatalf("docker args leaked registry token: %+v", call.args)
		}
	}
}
