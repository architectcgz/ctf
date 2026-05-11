package infrastructure

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"

	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/go-connections/nat"

	"ctf-platform/internal/config"
	runtimedomain "ctf-platform/internal/module/runtime/domain"
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

func TestValidateReusableNetwork(t *testing.T) {
	t.Parallel()

	expectedLabels := runtimedomain.ChallengeInstanceLabels(runtimedomain.ComposeServiceAWD)
	tests := []struct {
		name    string
		network networktypes.Inspect
		wantErr bool
	}{
		{
			name: "legacy managed network without compose labels is reusable",
			network: networktypes.Inspect{
				Internal: true,
				Labels: map[string]string{
					runtimedomain.ProjectLabelKey:           runtimedomain.ProjectLabelValue,
					runtimedomain.ManagedByLabelKey:         runtimedomain.ManagedByLabelValue,
					runtimedomain.ChallengeInstanceLabelKey: runtimedomain.ChallengeInstanceLabelValue,
				},
			},
		},
		{
			name: "compose managed network is reusable",
			network: networktypes.Inspect{
				Internal: true,
				Labels: map[string]string{
					runtimedomain.ProjectLabelKey:           runtimedomain.ProjectLabelValue,
					runtimedomain.ManagedByLabelKey:         runtimedomain.ManagedByLabelValue,
					runtimedomain.ChallengeInstanceLabelKey: runtimedomain.ChallengeInstanceLabelValue,
					runtimedomain.ComposeProjectLabelKey:    runtimedomain.ProjectLabelValue,
					runtimedomain.ComposeServiceLabelKey:    runtimedomain.ComposeServiceAWD,
				},
			},
		},
		{
			name: "non managed network is rejected",
			network: networktypes.Inspect{
				Internal: true,
				Labels: map[string]string{
					runtimedomain.ProjectLabelKey: runtimedomain.ProjectLabelValue,
				},
			},
			wantErr: true,
		},
		{
			name: "compose service mismatch is rejected",
			network: networktypes.Inspect{
				Internal: true,
				Labels: map[string]string{
					runtimedomain.ProjectLabelKey:           runtimedomain.ProjectLabelValue,
					runtimedomain.ManagedByLabelKey:         runtimedomain.ManagedByLabelValue,
					runtimedomain.ChallengeInstanceLabelKey: runtimedomain.ChallengeInstanceLabelValue,
					runtimedomain.ComposeProjectLabelKey:    runtimedomain.ProjectLabelValue,
					runtimedomain.ComposeServiceLabelKey:    runtimedomain.ComposeServiceJeopardy,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateReusableNetwork("ctf-awd-contest-8", expectedLabels, true, tt.network)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateReusableNetwork() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
