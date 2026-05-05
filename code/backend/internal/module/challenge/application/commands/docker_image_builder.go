package commands

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	challengeports "ctf-platform/internal/module/challenge/ports"
)

type dockerCommandRunner interface {
	Run(ctx context.Context, name string, args ...string) (string, error)
	RunWithEnv(ctx context.Context, env []string, name string, args ...string) (string, error)
}

type execDockerCommandRunner struct{}

func (execDockerCommandRunner) Run(ctx context.Context, name string, args ...string) (string, error) {
	return runDockerCommand(ctx, nil, name, args...)
}

func (execDockerCommandRunner) RunWithEnv(ctx context.Context, env []string, name string, args ...string) (string, error) {
	return runDockerCommand(ctx, env, name, args...)
}

func runDockerCommand(ctx context.Context, env []string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	if err := cmd.Run(); err != nil {
		return output.String(), fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(output.String()))
	}
	return output.String(), nil
}

type DockerCLIImageBuilderConfig struct {
	RegistryServer string
	Username       string
	Password       string
	IdentityToken  string
}

type DockerCLIImageBuilder struct {
	runner dockerCommandRunner
	config DockerCLIImageBuilderConfig
}

func NewDockerCLIImageBuilder() *DockerCLIImageBuilder {
	return &DockerCLIImageBuilder{runner: execDockerCommandRunner{}}
}

func NewDockerCLIImageBuilderWithConfig(config DockerCLIImageBuilderConfig) *DockerCLIImageBuilder {
	return &DockerCLIImageBuilder{runner: execDockerCommandRunner{}, config: config}
}

func newDockerCLIImageBuilderWithRunner(runner dockerCommandRunner) *DockerCLIImageBuilder {
	return newDockerCLIImageBuilderWithRunnerAndConfig(runner, DockerCLIImageBuilderConfig{})
}

func newDockerCLIImageBuilderWithRunnerAndConfig(runner dockerCommandRunner, config DockerCLIImageBuilderConfig) *DockerCLIImageBuilder {
	if runner == nil {
		runner = execDockerCommandRunner{}
	}
	return &DockerCLIImageBuilder{runner: runner, config: config}
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
	env, cleanup, err := b.registryAuthEnv()
	if err != nil {
		return err
	}
	defer cleanup()
	_, err = b.runner.RunWithEnv(ctx, env, "docker", "push", targetRef)
	return err
}

func (b *DockerCLIImageBuilder) Pull(ctx context.Context, targetRef string) error {
	env, cleanup, err := b.registryAuthEnv()
	if err != nil {
		return err
	}
	defer cleanup()
	_, err = b.runner.RunWithEnv(ctx, env, "docker", "pull", targetRef)
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

func (b *DockerCLIImageBuilder) registryAuthEnv() ([]string, func(), error) {
	cleanup := func() {}
	config := b.config
	server := strings.Trim(strings.TrimSpace(config.RegistryServer), "/")
	username := strings.TrimSpace(config.Username)
	password := strings.TrimSpace(config.Password)
	identityToken := strings.TrimSpace(config.IdentityToken)
	if server == "" || (identityToken == "" && username == "" && password == "") {
		return nil, cleanup, nil
	}
	if identityToken == "" && (username == "" || password == "") {
		return nil, cleanup, fmt.Errorf("docker registry auth requires username/password or identity token")
	}

	dockerConfigDir, err := os.MkdirTemp("", "ctf-docker-auth-*")
	if err != nil {
		return nil, cleanup, fmt.Errorf("create docker auth config: %w", err)
	}
	cleanup = func() {
		_ = os.RemoveAll(dockerConfigDir)
	}
	if err := os.Chmod(dockerConfigDir, 0o700); err != nil {
		cleanup()
		return nil, func() {}, fmt.Errorf("secure docker auth config: %w", err)
	}

	auth := dockerAuthEntry{}
	if identityToken != "" {
		auth.IdentityToken = identityToken
	} else {
		auth.Auth = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	}
	content, err := json.Marshal(dockerConfigFile{Auths: map[string]dockerAuthEntry{server: auth}})
	if err != nil {
		cleanup()
		return nil, func() {}, fmt.Errorf("encode docker auth config: %w", err)
	}
	configPath := filepath.Join(dockerConfigDir, "config.json")
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		cleanup()
		return nil, func() {}, fmt.Errorf("write docker auth config: %w", err)
	}
	return []string{"DOCKER_CONFIG=" + dockerConfigDir}, cleanup, nil
}

type dockerConfigFile struct {
	Auths map[string]dockerAuthEntry `json:"auths"`
}

type dockerAuthEntry struct {
	Auth          string `json:"auth,omitempty"`
	IdentityToken string `json:"identitytoken,omitempty"`
}
