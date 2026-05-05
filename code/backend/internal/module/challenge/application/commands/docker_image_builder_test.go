package commands

import (
	"context"
	"reflect"
	"testing"
)

type dockerRunnerCall struct {
	name string
	args []string
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
