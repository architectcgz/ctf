package infrastructure

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/docker/docker/api/types/registry"
	"github.com/docker/go-connections/nat"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
)

func TestBuildSecurityOpts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		seccomp string
		want    []string
	}{
		{
			name:    "empty uses docker default seccomp",
			seccomp: "",
			want:    []string{"no-new-privileges:true"},
		},
		{
			name:    "default uses docker default seccomp",
			seccomp: "default",
			want:    []string{"no-new-privileges:true"},
		},
		{
			name:    "trimmed default uses docker default seccomp",
			seccomp: " default ",
			want:    []string{"no-new-privileges:true"},
		},
		{
			name:    "unconfined is passed through",
			seccomp: "unconfined",
			want:    []string{"seccomp=unconfined", "no-new-privileges:true"},
		},
		{
			name:    "custom profile is passed through",
			seccomp: "/etc/docker/seccomp/ctf.json",
			want:    []string{"seccomp=/etc/docker/seccomp/ctf.json", "no-new-privileges:true"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := buildSecurityOpts(tt.seccomp)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("buildSecurityOpts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultSecurityConfigUsesNormalizedSecurityOpts(t *testing.T) {
	t.Parallel()

	cfg := &config.ContainerConfig{
		ReadonlyRootfs:      true,
		AllowedCapabilities: []string{"CHOWN"},
		RunAsUser:           "1000:1000",
		Seccomp:             "default",
	}

	got := DefaultSecurityConfig(cfg)
	want := []string{"no-new-privileges:true"}
	if !reflect.DeepEqual(got.SecurityOpt, want) {
		t.Fatalf("DefaultSecurityConfig().SecurityOpt = %v, want %v", got.SecurityOpt, want)
	}
}

func TestResolveContainerFilePathUsesWorkingDirForRelativePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		workingDir string
		filePath   string
		want       string
	}{
		{
			name:       "relative path uses container working dir",
			workingDir: "/app",
			filePath:   "app.py",
			want:       "/app/app.py",
		},
		{
			name:       "relative nested path uses container working dir",
			workingDir: "/app",
			filePath:   "src/main.py",
			want:       "/app/src/main.py",
		},
		{
			name:       "empty working dir falls back to root",
			workingDir: "",
			filePath:   "app.py",
			want:       "/app.py",
		},
		{
			name:       "absolute path is preserved",
			workingDir: "/app",
			filePath:   "/etc/hosts",
			want:       "/etc/hosts",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := resolveContainerFilePath(tt.workingDir, tt.filePath)
			if got != tt.want {
				t.Fatalf("resolveContainerFilePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildImagePullRegistryAuthMatchesConfiguredRegistry(t *testing.T) {
	t.Parallel()

	auth := buildImagePullRegistryAuth("registry.example.edu/ctf/awd-supply-ticket:v1", config.ContainerRegistryConfig{
		Enabled:  true,
		Server:   "https://registry.example.edu/",
		Username: "ctf",
		Password: "registry-token",
	})
	if auth == "" {
		t.Fatal("buildImagePullRegistryAuth() returned empty auth for configured registry")
	}

	decoded, err := base64.URLEncoding.DecodeString(auth)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}

	var got registry.AuthConfig
	if err := json.Unmarshal(decoded, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got.ServerAddress != "registry.example.edu" {
		t.Fatalf("ServerAddress = %q, want registry.example.edu", got.ServerAddress)
	}
	if got.Username != "ctf" {
		t.Fatalf("Username = %q, want ctf", got.Username)
	}
	if got.Password != "registry-token" {
		t.Fatalf("Password = %q, want registry-token", got.Password)
	}
}

func TestBuildImagePullRegistryAuthSkipsUnmatchedRegistry(t *testing.T) {
	t.Parallel()

	auth := buildImagePullRegistryAuth("docker.io/library/nginx:latest", config.ContainerRegistryConfig{
		Enabled:  true,
		Server:   "registry.example.edu",
		Username: "ctf",
		Password: "registry-token",
	})
	if auth != "" {
		t.Fatalf("buildImagePullRegistryAuth() = %q, want empty auth for public registry", auth)
	}
}

func TestSelectServicePort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		exposedPorts  nat.PortSet
		preferredPort int
		want          int
	}{
		{
			name:          "preferred port wins when exposed",
			exposedPorts:  nat.PortSet{"8080/tcp": struct{}{}, "80/tcp": struct{}{}},
			preferredPort: 8080,
			want:          8080,
		},
		{
			name:          "single exposed port is used",
			exposedPorts:  nat.PortSet{"80/tcp": struct{}{}},
			preferredPort: 8080,
			want:          80,
		},
		{
			name:          "web port preferred over arbitrary lowest port",
			exposedPorts:  nat.PortSet{"80/tcp": struct{}{}, "3306/tcp": struct{}{}},
			preferredPort: 8080,
			want:          80,
		},
		{
			name:          "preferred port returned when image exposes nothing",
			exposedPorts:  nat.PortSet{},
			preferredPort: 8080,
			want:          8080,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := selectServicePort(tt.exposedPorts, tt.preferredPort)
			if got != tt.want {
				t.Fatalf("selectServicePort() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestResolveContainerResourceLimitsClonesInput(t *testing.T) {
	t.Parallel()

	input := &model.ResourceLimits{
		CPUQuota:  1.5,
		Memory:    256 * 1024 * 1024,
		PidsLimit: 128,
	}

	resolved, err := resolveContainerResourceLimits(input, &config.ContainerConfig{})
	if err != nil {
		t.Fatalf("resolveContainerResourceLimits() error = %v", err)
	}
	if resolved == input {
		t.Fatal("expected resolved resource limits to be a clone, got original pointer")
	}

	resolved.Memory = 512 * 1024 * 1024
	if input.Memory != 256*1024*1024 {
		t.Fatalf("expected input memory to stay unchanged, got %d", input.Memory)
	}
}

func TestResolveContainerResourceLimitsUsesDefaults(t *testing.T) {
	t.Parallel()

	resolved, err := resolveContainerResourceLimits(nil, &config.ContainerConfig{
		DefaultCPUQuota:  2,
		DefaultMemory:    512 * 1024 * 1024,
		DefaultPidsLimit: 256,
	})
	if err != nil {
		t.Fatalf("resolveContainerResourceLimits() error = %v", err)
	}
	if resolved.CPUQuota != 2 || resolved.Memory != 512*1024*1024 || resolved.PidsLimit != 256 {
		t.Fatalf("unexpected resolved defaults: %+v", resolved)
	}
}

func TestResolveContainerSecurityConfigClonesInput(t *testing.T) {
	t.Parallel()

	input := &model.SecurityConfig{
		ReadonlyRootfs: true,
		CapDrop:        []string{"ALL"},
		CapAdd:         []string{"NET_BIND_SERVICE"},
		SecurityOpt:    []string{"no-new-privileges:true"},
		User:           "1000:1000",
	}

	resolved := resolveContainerSecurityConfig(input, &config.ContainerConfig{})
	if resolved == input {
		t.Fatal("expected resolved security config to be a clone, got original pointer")
	}
	if &resolved.CapDrop[0] == &input.CapDrop[0] {
		t.Fatal("expected CapDrop slice to be cloned")
	}
	if &resolved.CapAdd[0] == &input.CapAdd[0] {
		t.Fatal("expected CapAdd slice to be cloned")
	}
	if &resolved.SecurityOpt[0] == &input.SecurityOpt[0] {
		t.Fatal("expected SecurityOpt slice to be cloned")
	}
}

func TestResolveContainerSecurityConfigUsesDefaults(t *testing.T) {
	t.Parallel()

	resolved := resolveContainerSecurityConfig(nil, &config.ContainerConfig{
		ReadonlyRootfs:      true,
		AllowedCapabilities: []string{"CHOWN"},
		RunAsUser:           "1000:1000",
		Seccomp:             "default",
	})
	if !resolved.ReadonlyRootfs || resolved.User != "1000:1000" {
		t.Fatalf("unexpected resolved defaults: %+v", resolved)
	}
	if !reflect.DeepEqual(resolved.CapDrop, []string{"ALL"}) {
		t.Fatalf("unexpected CapDrop defaults: %v", resolved.CapDrop)
	}
	if !reflect.DeepEqual(resolved.CapAdd, []string{"CHOWN"}) {
		t.Fatalf("unexpected CapAdd defaults: %v", resolved.CapAdd)
	}
}
