package commands

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	challengeports "ctf-platform/internal/module/challenge/ports"
)

type dockerCommandRunner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
}

type execDockerCommandRunner struct{}

func (execDockerCommandRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		return output.String(), fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(output.String()))
	}
	return output.String(), nil
}

type DockerCLIImageBuilder struct {
	runner dockerCommandRunner
}

func NewDockerCLIImageBuilder() *DockerCLIImageBuilder {
	return &DockerCLIImageBuilder{runner: execDockerCommandRunner{}}
}

func newDockerCLIImageBuilderWithRunner(runner dockerCommandRunner) *DockerCLIImageBuilder {
	if runner == nil {
		runner = execDockerCommandRunner{}
	}
	return &DockerCLIImageBuilder{runner: runner}
}

func (b *DockerCLIImageBuilder) Build(ctx context.Context, contextPath, dockerfilePath, localRef string) error {
	_, err := b.runner.Run(ctx, "docker", "build", "-f", dockerfilePath, "-t", localRef, contextPath)
	return err
}

func (b *DockerCLIImageBuilder) Tag(ctx context.Context, sourceRef, targetRef string) error {
	_, err := b.runner.Run(ctx, "docker", "tag", sourceRef, targetRef)
	return err
}

func (b *DockerCLIImageBuilder) Push(ctx context.Context, targetRef string) error {
	_, err := b.runner.Run(ctx, "docker", "push", targetRef)
	return err
}

func (b *DockerCLIImageBuilder) Pull(ctx context.Context, targetRef string) error {
	_, err := b.runner.Run(ctx, "docker", "pull", targetRef)
	return err
}

func (b *DockerCLIImageBuilder) Inspect(ctx context.Context, targetRef string) (challengeports.ImageInspectResult, error) {
	output, err := b.runner.Run(ctx, "docker", "image", "inspect", "--format", "{{.Size}}", targetRef)
	if err != nil {
		return challengeports.ImageInspectResult{}, err
	}
	size, err := strconv.ParseInt(strings.TrimSpace(output), 10, 64)
	if err != nil {
		return challengeports.ImageInspectResult{}, fmt.Errorf("parse docker image size for %s: %w", targetRef, err)
	}
	return challengeports.ImageInspectResult{Size: size}, nil
}
